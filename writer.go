package wave

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Writer creates a writer for wave files encapsulating an io.Writer.
// It supports 8, 16 and 32 bit integer and 32 bit float formats.
type Writer struct {
	W          io.Writer
	H          Header
	ctr        int
	numSamples int
}

// NewWriter creates a wave Writer and writes its header. Subsequent calls to
// WriteInt and WriteFloat should be used to write the actual sample data.
func NewWriter(w io.Writer, channels, samplesPerSec, bitsPerSample, numSamples int) (*Writer, error) {

	subChunk2Size := uint32(numSamples * channels * bitsPerSample / 8)
	h := Header{
		RiffID:        [4]byte{'R', 'I', 'F', 'F'},
		DataSize:      36 + uint32(subChunk2Size),
		RiffType:      [4]byte{'W', 'A', 'V', 'E'},
		FmtChunkID:    [4]byte{'f', 'm', 't', ' '},
		FmtChunkSize:  16,
		AudioFmt:      1,
		Channels:      uint16(channels),
		SamplesPerSec: uint32(samplesPerSec),
		BytesPerSec:   uint32(samplesPerSec * channels * bitsPerSample / 8),
		BlockAlign:    uint16(channels * bitsPerSample / 8),
		BitsPerSample: uint16(bitsPerSample),
		DataChunkID:   [4]byte{'d', 'a', 't', 'a'},
		DataChunkSize: subChunk2Size,
	}

	if err := binary.Write(w, binary.LittleEndian, &h); err != nil {
		return nil, fmt.Errorf("error writing wave header in NewWriter: %s", err)
	}

	return &Writer{w, h, 0, numSamples}, nil
}

// WriteInt writes samples as bitsPerSample integers for 8, 16 and 32 bit PCM.
func (w *Writer) WriteInt(samples []int64) error {
	if len(samples) != int(w.H.Channels) {
		return fmt.Errorf("number of samples != channels in WriteInt: want %d: got %d", w.H.Channels, len(samples))
	}

	if w.ctr+1 > w.numSamples {
		return fmt.Errorf("overflow error: attempting to write too many samples: already wrote %d", w.ctr)
	}

	var reterr error
	switch w.H.BitsPerSample {
	case 8:
		wsamples := make([]int8, w.H.Channels)
		for j, sample := range samples {
			wsamples[j] = int8(sample)
		}
		if err := binary.Write(w.W, binary.LittleEndian, wsamples); err != nil {
			reterr = err
		}
	case 16:
		wsamples := make([]int16, w.H.Channels)
		for j, sample := range samples {
			wsamples[j] = int16(sample)
		}
		if err := binary.Write(w.W, binary.LittleEndian, wsamples); err != nil {
			reterr = err
		}
	case 32:
		wsamples := make([]int32, w.H.Channels)
		for j, sample := range samples {
			wsamples[j] = int32(sample)
		}
		if err := binary.Write(w.W, binary.LittleEndian, wsamples); err != nil {
			reterr = err
		}
	default:
		return fmt.Errorf("unrecognized bitsPerSample: %d", w.H.BitsPerSample)
	}
	if reterr == nil {
		w.ctr++
		return nil
	}
	return fmt.Errorf("error writing sample in WriteInt:%s", reterr)
}

// WriteFloat writes wave samples as 32 bit floating point numbers
func (w *Writer) WriteFloat(samples []float64) error {
	if len(samples) != int(w.H.Channels) {
		return fmt.Errorf("number of samples != channels in WriteInt: want %d: got %d", w.H.Channels, len(samples))
	}

	if w.ctr+1 > w.numSamples {
		return fmt.Errorf("overflow error: attempting to write too many samples: already wrote %d", w.ctr)
	}

	if w.H.BitsPerSample != 32 {
		return fmt.Errorf("only 32 bit floats are supported. bitsPerSample in Header is set to: %d", w.H.BitsPerSample)
	}

	wsamples := make([]float32, w.H.Channels)
	for j, sample := range samples {
		wsamples[j] = float32(sample)
	}
	if err := binary.Write(w.W, binary.LittleEndian, wsamples); err != nil {
		return fmt.Errorf("error writing sample in WriteFloat: %s", err)
	}

	w.ctr++
	return nil
}
