package main

import (
	"fmt"
	"github.com/lizc2003/fdk-aac-go"
	"os"
)

func main() {
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

	totalBytes, totalFrames, sampleRate, err := fdkaac.EncodeWavStream(in, out, &fdkaac.AacEncoderConfig{
		TransMux: fdkaac.TtMp4Adts,
		Bitrate:  128000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("totalBytes: %d, totalFrames: %d, sampleRate: %d\n", totalBytes, totalFrames, sampleRate)
}
