package fdkaac

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// EncodeWavStream encodes a WAV audio stream into AAC format.
// It reads PCM data from the input reader (wavStream) and writes the encoded AAC data to the output writer (writer).
// The encoding configuration is specified by the config parameter.
// This function parses the WAV header to extract SampleRate and MaxChannels, overriding the values in config.
func EncodeWavStream(wavStream io.Reader, writer io.Writer, config *AacEncoderConfig) (int, int, int, error) {
	sampleRate, numChannels, pcmSize, err := ParseWavHeader(wavStream)
	if err != nil {
		return 0, 0, 0, err
	}
	config.SampleRate = sampleRate
	config.MaxChannels = numChannels
	// Limit the reader to the data size to avoid reading trailing metadata as audio.
	wavStream = io.LimitReader(wavStream, pcmSize)

	encoder, err := CreateAacEncoder(config)
	if err != nil {
		return 0, 0, 0, err
	}
	defer encoder.Close()

	// Buffer for reading input PCM data
	// Read a reasonable chunk size, e.g., aligned with frame size if possible,
	// but simple reading is fine as Encode handles buffering.
	// However, Encode expects input to be byte slice.
	// Let's read in chunks.
	readBufSize := encoder.FrameSize
	inBuf := make([]byte, readBufSize)
	outBuf := make([]byte, encoder.EstimateOutBufBytes(readBufSize))
	totalBytes := 0
	totalFrames := 0

	for {
		n, err := wavStream.Read(inBuf)
		if n > 0 {
			encodedBytes, nFrames, encErr := encoder.Encode(inBuf[:n], outBuf)
			if encErr != nil {
				return 0, 0, 0, encErr
			}
			if encodedBytes > 0 {
				totalBytes += encodedBytes
				totalFrames += nFrames
				if _, wErr := writer.Write(outBuf[:encodedBytes]); wErr != nil {
					return 0, 0, 0, wErr
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, 0, err
		}
	}

	encodedBytes, nFrames, flushErr := encoder.Flush(outBuf)
	if flushErr != nil {
		return 0, 0, 0, flushErr
	}
	if encodedBytes > 0 {
		totalBytes += encodedBytes
		totalFrames += nFrames
		if _, wErr := writer.Write(outBuf[:encodedBytes]); wErr != nil {
			return 0, 0, 0, wErr
		}
	}

	return totalBytes, totalFrames, sampleRate, nil
}

func ParseWavHeader(wavStream io.Reader) (sampleRate int, numChannels int, pcmSize int64, err error) {
	var (
		riffHeader    [12]byte
		chunkHeader   [8]byte
		fmtChunkFound bool
	)

	// Read RIFF header
	if _, err := io.ReadFull(wavStream, riffHeader[:]); err != nil {
		return 0, 0, 0, fmt.Errorf("read RIFF header failed: %w", err)
	}
	if string(riffHeader[0:4]) != "RIFF" || string(riffHeader[8:12]) != "WAVE" {
		return 0, 0, 0, errors.New("invalid WAV header: missing RIFF/WAVE")
	}

	// Loop chunks
	for {
		if _, err := io.ReadFull(wavStream, chunkHeader[:]); err != nil {
			return 0, 0, 0, fmt.Errorf("read chunk header failed: %w", err)
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		if chunkID == "fmt " {
			if chunkSize < 16 {
				return 0, 0, 0, fmt.Errorf("invalid fmt chunk size: %d", chunkSize)
			}
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(wavStream, fmtData); err != nil {
				return 0, 0, 0, fmt.Errorf("read fmt chunk failed: %w", err)
			}

			audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
			numChannels = int(binary.LittleEndian.Uint16(fmtData[2:4]))
			sampleRate = int(binary.LittleEndian.Uint32(fmtData[4:8]))
			bitsPerSample := binary.LittleEndian.Uint16(fmtData[14:16])

			if audioFormat != 1 {
				return 0, 0, 0, fmt.Errorf("unsupported audio format: %d (only PCM supported)", audioFormat)
			}
			if bitsPerSample != SampleBitDepth {
				return 0, 0, 0, fmt.Errorf("unsupported bits per sample: %d (only 16-bit supported)", bitsPerSample)
			}
			fmtChunkFound = true
		} else if chunkID == "data" {
			if !fmtChunkFound {
				return 0, 0, 0, errors.New("data chunk found before fmt chunk")
			}
			// We found data chunk, stop parsing.
			pcmSize = int64(chunkSize)
			break
		} else {
			// Skip other chunks
			if _, err := io.CopyN(io.Discard, wavStream, int64(chunkSize)); err != nil {
				return 0, 0, 0, fmt.Errorf("skip chunk %s failed: %w", chunkID, err)
			}
		}
	}
	return sampleRate, numChannels, pcmSize, nil
}
