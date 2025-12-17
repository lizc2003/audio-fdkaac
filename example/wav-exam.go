package main

import (
	"fmt"
	"github.com/lizc2003/audio-fdkaac"
	"os"
)

func main() {
	encodeFromWav()
	decodeToWav()
}

func encodeFromWav() {
	in, err := os.Open("samples/sample.wav")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()

	out, err := os.Create("output.aac")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()

	totalBytes, totalFrames, sampleRate, err := fdkaac.EncodeFromWav(in, out, &fdkaac.EncoderConfig{
		TransMux: fdkaac.TtMp4Adts,
		Bitrate:  128000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("encoded %d bytes, total frames: %d, sample rate: %d\n", totalBytes, totalFrames, sampleRate)
}

func decodeToWav() {
	aacFile, err := os.Open("samples/sample.aac")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aacFile.Close()

	wavFile, err := os.Create("output.wav")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer wavFile.Close()

	config := &fdkaac.DecoderConfig{
		TransportFmt: fdkaac.TtMp4Adts,
	}

	totalBytes, totalSamples, sampleRate, err := fdkaac.DecodeToWav(aacFile, wavFile, config)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("decoded %d bytes, total samples: %d, sample rate: %d\n", totalBytes, totalSamples, sampleRate)
}
