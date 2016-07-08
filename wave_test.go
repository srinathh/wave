package wave

import (
	"bytes"
	"math"
	"testing"
)

// These characteristics describe the wave test data sample
const (
	testSampleCount   = 89
	testChannels      = 2
	testBitsPerSample = 16
	testSamplesPerSec = 44100
)

// testData is a 2 channel 16 bit wave file representing a square wave of amplitude ~ 0.7
var testData = []byte{82, 73, 70, 70, 136, 1, 0, 0, 87, 65, 86, 69, 102, 109, 116, 32, 16, 0, 0, 0, 1, 0, 2, 0, 68, 172, 0, 0, 16, 177, 2, 0, 4, 0, 16, 0, 100, 97, 116, 97, 100, 1, 0, 0, 0, 0, 255, 255, 34, 73, 36, 73, 220, 90, 217, 90, 95, 89, 99, 89, 221, 89, 218, 89, 162, 89, 164, 89, 195, 89, 193, 89, 176, 89, 178, 89, 186, 89, 185, 89, 181, 89, 182, 89, 183, 89, 183, 89, 183, 89, 182, 89, 182, 89, 184, 89, 184, 89, 181, 89, 182, 89, 185, 89, 183, 89, 179, 89, 182, 89, 187, 89, 184, 89, 179, 89, 181, 89, 186, 89, 185, 89, 180, 89, 181, 89, 184, 89, 183, 89, 182, 89, 183, 89, 183, 89, 183, 89, 183, 89, 182, 89, 183, 89, 183, 89, 182, 89, 182, 89, 183, 89, 184, 89, 183, 89, 182, 89, 182, 89, 183, 89, 183, 89, 183, 89, 184, 89, 182, 89, 180, 89, 184, 89, 187, 89, 180, 89, 178, 89, 186, 89, 187, 89, 180, 89, 179, 89, 187, 89, 187, 89, 175, 89, 176, 89, 194, 89, 194, 89, 165, 89, 165, 89, 217, 89, 216, 89, 102, 89, 104, 89, 200, 90, 198, 90, 75, 77, 77, 77, 244, 8, 241, 8, 174, 187, 179, 187, 62, 165, 57, 165, 150, 166, 154, 166, 43, 166, 40, 166, 87, 166, 90, 166, 66, 166, 63, 166, 78, 166, 82, 166, 69, 166, 65, 166, 78, 166, 80, 166, 68, 166, 69, 166, 78, 166, 76, 166, 70, 166, 71, 166, 75, 166, 76, 166, 73, 166, 70, 166, 73, 166, 76, 166, 74, 166, 72, 166, 72, 166, 73, 166, 75, 166, 74, 166, 71, 166, 73, 166, 77, 166, 74, 166, 69, 166, 72, 166, 76, 166, 74, 166, 72, 166, 72, 166, 74, 166, 75, 166, 73, 166, 73, 166, 74, 166, 73, 166, 71, 166, 74, 166, 75, 166, 71, 166, 74, 166, 77, 166, 71, 166, 69, 166, 76, 166, 78, 166, 70, 166, 69, 166, 76, 166, 76, 166, 71, 166, 72, 166, 77, 166, 74, 166, 68, 166, 72, 166, 80, 166, 77, 166, 63, 166, 65, 166, 90, 166, 89, 166, 43, 166, 42, 166, 143, 166, 145, 166, 96, 165, 95, 165, 55, 175, 57, 175, 27, 0, 7, 0}

func TestReadHeader(t *testing.T) {
	r, err := NewReader(bytes.NewReader(testData))
	if err != nil {
		t.Fatalf("Error reading test data header: %s", err)
	}

	if numSamples := r.GetSampleCount(); numSamples != testSampleCount {
		t.Fatalf("Sample count mismatch: want %d: got %d", testSampleCount, numSamples)
	} else {
		t.Logf("Sample Count: %d", numSamples)
	}

	if channels := r.GetChannels(); channels != testChannels {
		t.Fatalf("Channel  mismatch: want %d: got %d", testChannels, channels)
	} else {
		t.Logf("Channels: %d", channels)
	}

	if bitsPerSample := r.GetBitsPerSample(); bitsPerSample != testBitsPerSample {
		t.Fatalf("Channel  mismatch: want %d: got %d", testBitsPerSample, bitsPerSample)
	} else {
		t.Logf("bits per sample: %d", bitsPerSample)
	}
}

func TestReadData(t *testing.T) {
	r, err := NewReader(bytes.NewReader(testData))
	if err != nil {
		t.Fatalf("Error reading test data header: %s", err)
	}

	// since our sample data is a square wave with amplitude around 7, we will
	// count the number of samples where absolute sample value is +/- 0.02
	// and check whether these are at least 90% of the samples
	ctr := 0
	for j := 0; j < r.GetSampleCount(); j++ {
		sample, err := r.ReadInt()
		if err != nil {
			t.Fatal(err)
		}

		normL := math.Abs(float64(sample[0]) / 32768.0)

		t.Logf("L Sample: %d\t Norm L Sample: %f", sample[0], normL)

		if normL > 0.65 && normL < 0.75 {
			ctr = ctr + 1
		}
	}

	ratio := float64(ctr) / float64(r.GetSampleCount())
	if ratio < 0.9 {
		t.Fatalf("unexpected low amplitudes : want %f, got %f", 0.9, ratio)
	} else {
		t.Logf("got amplitude ratio: %f", ratio)
	}
}

func TestWriteData(t *testing.T) {
	r, err := NewReader(bytes.NewReader(testData))
	if err != nil {
		t.Fatalf("Error: creating reader in TestWriteData:%s", err)
	}

	samples := make([][]int64, r.GetSampleCount())

	for j := 0; j < r.GetSampleCount(); j++ {
		sample, err := r.ReadInt()
		if err != nil {
			t.Fatalf("Error: reading testdata in TestWriteData: %s", err)
		}
		samples[j] = sample
	}

	buf := bytes.Buffer{}
	w, err := NewWriter(&buf, testChannels, testSamplesPerSec, testBitsPerSample, testSampleCount)
	if err != nil {
		t.Fatalf("Error: creating wave writer in TestWriteData: %s", err)
	}

	for _, sample := range samples {
		if err := w.WriteInt(sample); err != nil {
			t.Fatalf("Error: writing sample: %s", err)
		}
	}

	want := testData
	got := buf.Bytes()

	if len(got) != len(want) {
		t.Fatalf("got & want data sizes do not match: want %d: got %d", len(want), len(got))
	}

	for j, item := range got {
		if want[j] != item {
			t.Fatalf("mismatch writing data in byte %d:\nwant\n%v\ngot:%v\n", j, want, got)
		}
	}

}
