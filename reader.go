package wavcodec

import (
	"encoding/binary"
	"fmt"
	"io"
)

/*
WaveHeader is a struct that holds Header information for a standard PCM Wave file.
The structure is used for reading & writing Wave File information. Reference sources:
https://blogs.msdn.microsoft.com/dawate/2009/06/23/intro-to-audio-programming-part-2-demystifying-the-wav-format/
*/
type WaveHeader struct {
	RiffID        [4]byte // *must be RIFF
	DataSize      uint32  // *36 + SubChunk2Size, or : 4 + (8 + SubChunk1Size) + (8 + SubChunk2Size)
	RiffType      [4]byte // *must be WAVE  *
	FmtChunkID    [4]byte // *must be "fmt "
	FmtChunkSize  uint32  // *must be 16 for PCM
	AudioFmt      uint16  // *will be 1 for PCM, if not 1, compression is used
	Channels      uint16  // Number of channels of audio
	SamplesPerSec uint32  // Sampling rate
	BytesPerSec   uint32  // SampleRate * NumChannels * BitsPerSample/8
	BlockAlign    uint16  // NumChannels * BitsPerSample/8
	BitsPerSample uint16  // 8, 16, 24  (int) or 32 (int/float)
	DataChunkID   [4]byte // *must be "data"
	DataChunkSize uint32  //= NumSamples * NumChannels * BitsPerSample/8
}

func byte4Cmp(b [4]byte, s string) bool {
	if len(s) != 4 {
		panic(fmt.Sprintf("comparing a [4]byte with string %s: size %d", s, len(s)))
	}
	for j := 0; j < 4; j++ {
		if b[j] != s[j] {
			return false
		}
	}
	return true
}

// Validate checks wheather a WaveHeader represents a supported Wave format
func (h WaveHeader) Validate() error {

	if !byte4Cmp(h.RiffID, "RIFF") {
		return fmt.Errorf("unexpected RiffID in header: want %s, got %s", "RIFF", h.RiffID)
	}

	if !byte4Cmp(h.RiffType, "WAVE") {
		return fmt.Errorf("unexpected RiffType in header: want %s, got %s", "WAVE", h.RiffType)
	}

	if !byte4Cmp(h.FmtChunkID, "fmt ") {
		return fmt.Errorf("unexpected FmtChunkId in header: want %s, got %s", "fmt ", h.FmtChunkID)
	}

	if !byte4Cmp(h.DataChunkID, "data") {
		return fmt.Errorf("unexpected DataChunkId in header: want %s, got %s", "data", h.DataChunkID)
	}

	if h.FmtChunkSize != 16 {
		return fmt.Errorf("unexpected FmtChunkSize in header: want %d, got %d", 16, h.FmtChunkSize)
	}

	if h.DataSize != 36+h.DataChunkSize {
		return fmt.Errorf("unexpected DataSize in header: want %d, got %d", 36+h.DataChunkSize, h.DataSize)
	}

	if h.AudioFmt != 1 {
		return fmt.Errorf("only uncompressed PCM formats are supported. got AudioFmt: %d", h.AudioFmt)
	}
	if h.BitsPerSample != 8 && h.BitsPerSample != 16 && h.BitsPerSample != 24 && h.BitsPerSample != 32 {
		return fmt.Errorf("only 8, 16, 24 or 32 bit sampels are supported. got BitsPerSample: %d", h.BitsPerSample)
	}
	return nil
}

// Reader is a reader for Wave file that encapsulates an io.Reader
type Reader struct {
	R      io.Reader
	Header WaveHeader
}

// NewReader creates a new wave reader by attempting to read the wave file header
// from the provided io.Reader. The wave samples can be subsequently read  from the
// wave reader using ReadInt or ReadFloat
func NewReader(r io.Reader) (*Reader, error) {
	h := WaveHeader{}
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("error decoding Header in NewReader:%s", err)
	}
	if err := h.Validate(); err != nil {
		return nil, fmt.Errorf("could not validate header: %s", err)
	}
	return &Reader{r, h}, nil
}

// GetSampleCount calculates the number of samples int he wave file
// using information on DataChunk size, number of channels & bit per
// sample information contained in the wave header.
func (r *Reader) GetSampleCount() int {
	return int(r.Header.DataChunkSize) / int(r.Header.Channels*(r.Header.BitsPerSample/8))
}

// GetBitsPerSample returns the number of bits per sample in the wave file.
func (r *Reader) GetBitsPerSample() int {
	return int(r.Header.BitsPerSample)
}

// GetChannels returns the number of channels in the wave file.
func (r *Reader) GetChannels() int {
	return int(r.Header.Channels)
}

// ReadInt reads the data from the wave file as a raw integer
// respecting BitsPerSample without performing any normalization.
// It returns the data read as int64 as a convenience to allow
// data processing without overflows.
func (r *Reader) ReadInt() ([]int64, error) {

	ret := make([]int64, r.Header.Channels)
	for j := 0; j < int(r.Header.Channels); j++ {
		switch r.Header.BitsPerSample {
		case 8:
			var data int8
			if err := binary.Read(r.R, binary.LittleEndian, &data); err != nil {
				return nil, err
			}
			ret[j] = int64(data)
		case 16:
			var data int16
			if err := binary.Read(r.R, binary.LittleEndian, &data); err != nil {
				return nil, err
			}
			ret[j] = int64(data)
		case 32:
			var data int32
			if err := binary.Read(r.R, binary.LittleEndian, &data); err != nil {
				return nil, err
			}
			ret[j] = int64(data)

		case 24:
			var data1 uint16
			var data2 int8
			if err := binary.Read(r.R, binary.LittleEndian, &data1); err != nil {
				return nil, err
			}
			if err := binary.Read(r.R, binary.LittleEndian, &data2); err != nil {
				return nil, err
			}
			ret[j] = int64(data2)<<16 + int64(data1)
		default:
			panic(fmt.Sprintf("unknonw bits per sample %d", r.Header.BitsPerSample))
		}
	}
	return ret, nil
}

// ReadRawFloat reads the data from the wave file as 32 bit floating
// point numbers. It returns 64 bit floats for computation convenience.
func (r *Reader) ReadRawFloat() ([]float64, error) {
	if r.Header.BitsPerSample != 32 {
		return nil, fmt.Errorf("unexpected BitsPerSample in ReadRawFloat: want 32, got %d", r.Header.BitsPerSample)
	}

	data := make([]float32, r.Header.Channels)

	if err := binary.Read(r.R, binary.LittleEndian, data); err != nil {
		return nil, fmt.Errorf("error reading data in ReadRawFloat: %s", err)
	}

	ret := make([]float64, r.Header.Channels)
	for j, item := range data {
		ret[j] = float64(item)
	}
	return ret, nil
}
