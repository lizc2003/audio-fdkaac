[![PkgGoDev](https://pkg.go.dev/badge/github.com/qrtc/fdk-aac-go)](https://pkg.go.dev/github.com/qrtc/fdk-aac-go)

# fdk-aac-go

Go bindings for [fdk-aac](https://github.com/mstorsjo/fdk-aac). A standalone library of the Fraunhofer FDK AAC code from Android.

## Why fdk-aac-go

The purpose of fdk-aac-go is easing the adoption of fdk-aac codec library. Using Go, with just a few lines of code you can implement an application that encode/decode data easy.

##  Is this a new implementation of fdk-aac?

No! We are just exposing the great work done by the research organization of [Fraunhofer IIS](https://www.iis.fraunhofer.de/en/ff/amm/impl.html) as a golang library. All the functionality and implementation still resides in the official fdk-aac project.

# Features

- **Encode PCM to AAC**: Convert raw PCM audio data to AAC format
- **Decode AAC to PCM**: Convert AAC audio data to raw PCM format
- **Multiple AAC Profiles**: Support for AAC-LC, HE-AAC, HE-AACv2, and AAC-ELD
- **Various Transport Formats**: ADTS, Raw, LATM/LOAS, and more
- **WAV File Support**: Direct encoding/decoding from/to WAV files
- **Streaming Support**: Process audio data in chunks without loading entire files
- **Error Handling**: Comprehensive error reporting and validation

# Usage

## Decode AAC frame to PCM

```go
package main

import (
	"fmt"

	fdkaac "github.com/lizc2003/audio-fdkaac"
)

func main() {
	decoder, err := fdkaac.NewDecoder(&fdkaac.DecoderConfig{
		TransportFmt: fdkaac.TtMp4Adts,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer decoder.Close()

	inBuf := []byte{
		// AAC frame data
	}
	// Use decoder's recommended buffer size
	outBuf := make([]byte, decoder.EstimateOutBufBytes(fdkaac.EstimateFrames))

	n, err := decoder.Decode(inBuf, outBuf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Decoded %d bytes of PCM data\n", n)
	
	// Get stream information
	info, err := decoder.GetStreamInfo()
	if err == nil {
		fmt.Printf("Sample Rate: %d, Channels: %d\n", info.SampleRate, info.NumChannels)
	}
	
	fmt.Println(outBuf[0:n])
}
```

## Encode PCM to AAC

```go
package main

import (
	"fmt"

	fdkaac "github.com/lizc2003/audio-fdkaac"
)

func main() {
	encoder, err := fdkaac.NewEncoder(&fdkaac.EncoderConfig{
		TransMux:    fdkaac.TtMp4Adts,
		SampleRate:  44100,
		MaxChannels: 2,
		Bitrate:     128000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer encoder.Close()

	inBuf := []byte{
		// PCM bytes (16-bit samples)
	}
	// Estimate output buffer size based on input size
	outBuf := make([]byte, encoder.EstimateOutBufBytes(len(inBuf)))

	n, nFrames, err := encoder.Encode(inBuf, outBuf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Encoded %d bytes in %d frames\n", n, nFrames)
	
	// Don't forget to flush at the end
	n2, nFrames2, err := encoder.Flush(outBuf[n:])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Flushed %d bytes in %d frames\n", n2, nFrames2)
	fmt.Println(outBuf[0:n+n2])
}

```

# Dependencies

* fdk-aac
