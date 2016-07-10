package wave

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Header is a struct that holds Header information for a standard PCM Wave file.
// The structure is used for reading & writing Wave File information.
//
// Reference source:
// https://blogs.msdn.microsoft.com/dawate/2009/06/23/intro-to-audio-programming-part-2-demystifying-the-wav-format/
type Header struct {
	RiffID        [4]byte // must be "RIFF"
	DataSize      uint32  // 36 + SubChunk2Size, Little Endian
	RiffType      [4]byte // must be "WAVE"
	FmtChunkID    [4]byte // must be "fmt "
	FmtChunkSize  uint32  // must be 16 for PCM, Little Endian
	AudioFmt      uint16  // must be 1 for standard PCM, else indicates compression, Little Endian
	Channels      uint16  // Number of channels of audio, Little Endian
	SamplesPerSec uint32  // Sampling rate, Little Endian
	BytesPerSec   uint32  // SampleRate * NumChannels * BitsPerSample/8, Little Endian
	BlockAlign    uint16  // NumChannels * BitsPerSample/8, Little Endian
	BitsPerSample uint16  // 8, 16, 24  (int) or 32 (int/float), Little Endian
	DataChunkID   [4]byte // must be "data"
	DataChunkSize uint32  // NumSamples * NumChannels * BitsPerSample/8
}

// byte4Cmp is a utility function that compares [4]byte types with strings
// used in wave header validation.
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

// Validate checks wheather a wave header represents a supported PCM Wave format.
func (h Header) Validate() error {

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

// GetSampleCount calculates the number of samples int he wave file
// using information on DataChunk size, number of channels & bit per
// sample information contained in the wave header.
func (h *Header) GetSampleCount() int {
	return int(h.DataChunkSize) / int(h.Channels*(h.BitsPerSample/8))
}

// GetBitsPerSample returns the number of bits per sample in the wave file.
func (h *Header) GetBitsPerSample() int {
	return int(h.BitsPerSample)
}

// GetChannels returns the number of channels in the wave file.
func (h *Header) GetChannels() int {
	return int(h.Channels)
}

// GetSamplesPerSec gets number of samples per second in the wave file.
func (h *Header) GetSamplesPerSec() int {
	return int(h.SamplesPerSec)
}

// Reader is a reader for Wave file that encapsulates an io.Reader
type Reader struct {
	R io.Reader
	H Header
}

// NewReader creates a new wave reader encapsulating the provided io.Reader.
// When `NewReader()` is called to create a `Reader`, it attempts to read the header information
// from the provided reader and validates if it is a supported format. Samples can then be
// read using `ReadInt()` or `ReadFloat()` functions depending on whether the data is expected
// to be integer or floating point.
func NewReader(r io.Reader) (*Reader, error) {
	h := Header{}
	if err := binary.Read(r, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("error decoding Header in NewReader:%s", err)
	}
	if err := h.Validate(); err != nil {
		return nil, fmt.Errorf("could not validate header: %s", err)
	}
	return &Reader{r, h}, nil
}

// ReadInt reads the data from the wave file as an integer. When reading data, the function respects
// the Bits Per Sample declared in the wave file header. The read functions return a `[]int64`
// where each slice element corresponds to the sample for a channel. The 64 bit types are meant
// to allow headroom for any further audio processing without clipping. The read data is simply
// cast into 64 bit integers and no other normalization or conversion is performed.
func (r *Reader) ReadInt() ([]int64, error) {

	ret := make([]int64, r.H.Channels)
	for j := 0; j < int(r.H.Channels); j++ {
		switch r.H.BitsPerSample {
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
			panic(fmt.Sprintf("unknonw bits per sample %d", r.H.BitsPerSample))
		}
	}
	return ret, nil
}

// ReadFloat reads the data from the wave file as an integer. When reading data, the function respects
// the Bits Per Sample declared in the wave file header. The read functions return a `[]float64`
// where each slice element corresponds to the sample for a channel. The 64 bit types are meant
// to allow headroom for any further audio processing without clipping. The read data is simply
// cast into 64 bit float and no other normalization or conversion is performed.
func (r *Reader) ReadFloat() ([]float64, error) {
	if r.H.BitsPerSample != 32 {
		return nil, fmt.Errorf("unexpected BitsPerSample in ReadRawFloat: want 32, got %d", r.H.BitsPerSample)
	}

	data := make([]float32, r.H.Channels)

	if err := binary.Read(r.R, binary.LittleEndian, data); err != nil {
		return nil, fmt.Errorf("error reading data in ReadRawFloat: %s", err)
	}

	ret := make([]float64, r.H.Channels)
	for j, item := range data {
		ret[j] = float64(item)
	}
	return ret, nil
}
