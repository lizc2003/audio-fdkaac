package main

import (
	"fmt"
	"github.com/lizc2003/fdk-aac-go"
)

func main() {
	encoder, err := fdkaac.CreateAccEncoder(&fdkaac.AacEncoderConfig{
		TransMux:    fdkaac.TtMp4Adts,
		AOT:         fdkaac.AotAacLc,
		SampleRate:  44100,
		MaxChannels: 2,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		encoder.Close()
	}()

	inBuf := []byte{
		// PCM bytes
	}
	outBuf := make([]byte, 4096)

	n, err := encoder.Encode(inBuf, outBuf)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(outBuf[0:n])
}
