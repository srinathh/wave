# wave [![Build Status](https://travis-ci.org/srinathh/wave.svg?branch=master)](https://travis-ci.org/srinathh/wave) [![GoDoc](https://godoc.org/github.com/srinathh/wave?status.svg)](https://godoc.org/github.com/srinathh/wave)

Package wave implements a simple reader and writer standard uncompressed wav files. 
Reading is supported across uncompressed 8, 16, 24 and 32 bit-depth integer formats
and 32 bit floating point formats. Writing is currently supported for 8, 16 and 32 bit 
integers and 32 bit floating point numbers.

The key motivation for creating [yet another](http://godoc.org/?q=wave) wave file reader/
writer was to create a package with minimal, idiomatic and well documented functions 
using standard Go language types which make no assumptions and normalizations of your data.

## Installation
`go get github.com/srinathh/wave`

## Usage
### Reader
The following example demonstrates how to use a Reader. First create a `wave.Reader` encapsulating
an `io.Reader` using the `NewReader()` function. If the reader is able to successfully read and 
validate the header, you can read samples using the `ReadInt()` and `ReadFloat()` functions.
```
r, err := NewReader(bytes.NewReader(testData))
if err != nil {
    fmt.Println(err)
    os.Exit(1)
}

fmt.Printf("Bits Per Sample: %d\n", r.H.GetBitsPerSample())
fmt.Printf("Samples Per Second: %d\n", r.H.GetSamplesPerSec())
fmt.Printf("Number of Channels: %d\n", r.H.GetChannels())
fmt.Printf("Sample Count: %d\n", r.H.GetSampleCount())
fmt.Printf("First 5 samples in Channel 0:")
for j := 0; j < 5; j++ {
    sample, err := r.ReadInt()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    fmt.Printf("\nSample %d: %d", j, sample[0])
}
```
which gives the output:
```
Bits Per Sample: 16
Samples Per Second: 44100
Number of Channels: 2
Sample Count: 89
First 5 samples in Channel 0:
Sample 0: 0
Sample 1: 18722
Sample 2: 23260
Sample 3: 22879
Sample 4: 23005
```
### Writer
The following code snippet demonstrates how to create a writer and write data to it.
The calling function must call close on the underlying writer if required.
```
    // Create a wave file that will contain 1 channel with 8 bit depth audio 
    // recorded at 44100 seconds. The sample will contain 2 seconds of audio
	w, err := NewWriter(&buf, 1, 44100, 8, 88200)
	if err != nil {
		t.Fatalf("Error: creating wave writer in TestWriteData: %s", err)
	}

    // samples is an [][]int64
	for _, sample := range samples {
		if err := w.WriteInt(sample); err != nil {
			t.Fatalf("Error: writing sample: %s", err)
		}
	}
```
## License
Apache 2 License