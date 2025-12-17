package main

import (
	"fmt"
	"github.com/lizc2003/audio-fdkaac"
	"io"
	"os"
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

	aacFile, err := os.Open("samples/sample.aac")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aacFile.Close()

	pcmBuf := make([]byte, decoder.EstimateOutBufBytes(fdkaac.EstimateFrames))
	totalBytes := 0
	chunk := make([]byte, 2048)

	for {
		n, readErr := aacFile.Read(chunk)
		if n > 0 {
			decodedN, decErr := decoder.Decode(chunk[:n], pcmBuf)
			if decErr != nil {
				fmt.Println(decErr)
				return
			}

			if decodedN == 0 {
				break
			}

			totalBytes += decodedN
		}

		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			fmt.Println(readErr)
			return
		}
	}

	fmt.Printf("Decoded %d bytes of PCM data\n", totalBytes)
}
