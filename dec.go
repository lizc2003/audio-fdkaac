package fdkaac

import (
	"errors"
	"fmt"
	"unsafe"
)

/*
#include "deps/include/aacdecoder_lib.h"

AAC_DECODER_ERROR aacDecoder_ConfigRawWrapped(HANDLE_AACDECODER self,
			UCHAR *conf, const UINT length) {
	return aacDecoder_ConfigRaw(self, &conf, &length);
}

AAC_DECODER_ERROR aacDecoder_DecodeWrapped(HANDLE_AACDECODER self,
			UCHAR *pBuffer, const UINT bufferSize, UCHAR *pOut, INT outSize, UINT *bytesDecode, UINT *bytesValid) {
	INT frameSize = 0;
	AAC_DECODER_ERROR errNo;
	errNo = aacDecoder_Fill(self, &pBuffer, &bufferSize, bytesValid);
	if (errNo != AAC_DEC_OK) {
		return errNo;
	}

	*bytesDecode = 0;
	for(;;) {
		errNo = aacDecoder_DecodeFrame(self, (INT_PCM *)pOut, outSize/2, 0);
		if (errNo != AAC_DEC_OK) {
			if (errNo == AAC_DEC_NOT_ENOUGH_BITS) {
				// No more complete frames in internal buffer
				break;
			}
			return errNo;
		}

		if (frameSize == 0) {
			CStreamInfo* info = aacDecoder_GetStreamInfo(self);
			frameSize = info->frameSize * info->numChannels * 2;
		}

		*bytesDecode += frameSize;
		outSize -= frameSize;
		pOut += frameSize;
	}
	return AAC_DEC_OK;
}

*/
import "C"

const (
	EstimateFrames = 10
)

var decErrors = [...]error{
	C.AAC_DEC_OK:                            nil,
	C.AAC_DEC_OUT_OF_MEMORY:                 errors.New("heap returned NULL pointer or output buffer is invalid"),
	C.AAC_DEC_UNKNOWN:                       errors.New("error condition is of unknown reason"),
	C.AAC_DEC_TRANSPORT_SYNC_ERROR:          errors.New("the transport decoder had synchronization problems"),
	C.AAC_DEC_NOT_ENOUGH_BITS:               errors.New("the input buffer ran out of bits"),
	C.AAC_DEC_INVALID_HANDLE:                errors.New("the handle passed to the function call was invalid"),
	C.AAC_DEC_UNSUPPORTED_AOT:               errors.New("the AOT found in the configuration is not supported"),
	C.AAC_DEC_UNSUPPORTED_FORMAT:            errors.New("the bitstream format is not supported"),
	C.AAC_DEC_UNSUPPORTED_ER_FORMAT:         errors.New("the error resilience tool format is not supported"),
	C.AAC_DEC_UNSUPPORTED_EPCONFIG:          errors.New("the error protection format is not supported"),
	C.AAC_DEC_UNSUPPORTED_MULTILAYER:        errors.New("more than one layer for AAC scalable is not supported"),
	C.AAC_DEC_UNSUPPORTED_CHANNELCONFIG:     errors.New("the channel configuration is not supported"),
	C.AAC_DEC_UNSUPPORTED_SAMPLINGRATE:      errors.New("the sample rate specified in the configuration is not supported"),
	C.AAC_DEC_INVALID_SBR_CONFIG:            errors.New("the SBR configuration is not supported"),
	C.AAC_DEC_SET_PARAM_FAIL:                errors.New("the parameter could not be set"),
	C.AAC_DEC_NEED_TO_RESTART:               errors.New("the decoder needs to be restarted"),
	C.AAC_DEC_OUTPUT_BUFFER_TOO_SMALL:       errors.New("the provided output buffer is too small"),
	C.AAC_DEC_TRANSPORT_ERROR:               errors.New("the transport decoder encountered an unexpected error"),
	C.AAC_DEC_PARSE_ERROR:                   errors.New("error while parsing the bitstream"),
	C.AAC_DEC_UNSUPPORTED_EXTENSION_PAYLOAD: errors.New("error while parsing the extension payload of the bitstream"),
	C.AAC_DEC_DECODE_FRAME_ERROR:            errors.New("the parsed bitstream value is out of range"),
	C.AAC_DEC_CRC_ERROR:                     errors.New("the embedded CRC did not match"),
	C.AAC_DEC_INVALID_CODE_BOOK:             errors.New("an invalid codebook was signaled"),
	C.AAC_DEC_UNSUPPORTED_PREDICTION:        errors.New("predictor found, but not supported in the AAC Low Complexity profile"),
	C.AAC_DEC_UNSUPPORTED_CCE:               errors.New("a CCE element was found which is not supported"),
	C.AAC_DEC_UNSUPPORTED_LFE:               errors.New("a LFE element was found which is not supported"),
	C.AAC_DEC_UNSUPPORTED_GAIN_CONTROL_DATA: errors.New("gain control data found but not supported"),
	C.AAC_DEC_UNSUPPORTED_SBA:               errors.New("SBA found but currently not supported in the BSAC profile"),
	C.AAC_DEC_TNS_READ_ERROR:                errors.New("error while reading TNS data"),
	C.AAC_DEC_RVLC_ERROR:                    errors.New("error while decoding error resilient data"),
	C.AAC_DEC_ANC_DATA_ERROR:                errors.New("non severe error concerning the ancillary data handling"),
	C.AAC_DEC_TOO_SMALL_ANC_BUFFER:          errors.New("the registered ancillary data buffer is too small to receive the parsed data"),
	C.AAC_DEC_TOO_MANY_ANC_ELEMENTS:         errors.New("more than the allowed number of ancillary data elements should be written to buffer"),
}

// getDecError safely converts C error code to Go error
func getDecError(errNo C.AAC_DECODER_ERROR) error {
	if int(errNo) >= 0 && int(errNo) < len(decErrors) {
		return decErrors[errNo]
	}
	return fmt.Errorf("unknown decoder error: %d", errNo)
}

// PcmDualChannelOutputMode defines how the decoder processes two channel signals.
type PcmDualChannelOutputMode int

const (
	PcmDualChannelLeaveBoth PcmDualChannelOutputMode = iota
	PcmDualChannelMonoCH1
	PcmDualChannelMonoCH2
	PcmDualChannelMix
)

// PcmLimiterMode enable signal level limiting.
type PcmLimiterMode int

const (
	PcmLimiterAutoConfig PcmLimiterMode = iota
	PcmLimiterEnable
	PcmLimiterDisable
)

// Meta data profile.
type MetaDataProfile int

const (
	MdProfileMpegStandard MetaDataProfile = iota
	MdProfileMpegLegacy
	MdProfileMpegLegacyPrio
	MdProfileAribJapan
)

// Error concealment: Processing method.
type ConcealMethod int

const (
	ConcealSpectralMuting ConcealMethod = iota
	ConcealNoiseSubstitution
	ConcealEnergyInterpolation
)

// MPEG-4 DRC: Default presentation mode.
type DrcDefaultPresentationMode int

const (
	DrcParameterHandlingDisabled DrcDefaultPresentationMode = iota
	DrcParameterHandlingEnabled
	DrcPresentationMode1Default
	DrcPresentationMode2Default
)

// Quadrature Mirror Filter (QMF) Bank processing mode.
type QmfLowpowerMode int

const (
	QmfLowpowerInternal QmfLowpowerMode = iota
	QmfLowpowerComplex
	QmfLowpowerReal
)

type DecoderConfig struct {
	// Transport type
	TransportFmt TransportType
	// Defines how the decoder processes two channel signals.
	PcmDualChannelOutputMode PcmDualChannelOutputMode
	// Output buffer channel ordering.
	PcmOutputChannelMappingMpeg bool
	// Enable signal level limiting.
	PcmLimiterMode PcmLimiterMode
	// Signal level limiting attack time in ms.
	PcmLimiterAttackTime int
	// Signal level limiting release time in ms.
	PcmLimiterReleasTime int
	// Minimum number of PCM output channels.
	PcmMinOutputChannels int
	// Maximum number of PCM output channels.
	PcmMaxOutputChannels int
	// Meta data profile.
	MetadataProfile MetaDataProfile
	// Defines the time in ms after which all the bitstream associated meta-data.
	MetadataExpiryTime int
	// Error concealment: Processing method.
	ConcealMethod ConcealMethod
	// MPEG-4 / MPEG-D Dynamic Range Control (DRC):
	// Defines how the boosting DRC factors will be applied to the decoded signal.
	DrcBoostFactor int
	// MPEG-4 DRC: Scaling factor for attenuating gain values.
	DrcAttenuationFactor int
	// MPEG-4 DRC: Defines the level below full-scale to which
	// the output audio signal will be normalized to by the DRC module.
	DrcReferenceLevel int
	// MPEG-4 DRC: Enable DVB specific heavy compression
	EnableDrcHeavyCompression bool
	// MPEG-4 DRC: Default presentation mode.
	DrcDefaultPresentationMode DrcDefaultPresentationMode
	// MPEG-4 DRC: Encoder target level for light.
	DrcEncTargetLevel int
	// MPEG-D DRC: Request a DRC effect type for selection of a DRC set.
	UnidrcSetEffect int
	// MPEG-D DRC: Enable album mode.
	EnableUnidrcAlbumMode bool
	// Quadrature Mirror Filter (QMF) Bank processing mode.
	QmfLowpowerMode QmfLowpowerMode
}

// StreamInfo gives information about the currently decoded audio data.
type StreamInfo struct {
	// The sample rate in Hz of the decoded PCM audio signal.
	SampleRate int
	// The frame length of the decoded PCM audio signal.
	FrameLength int
	// Bytes per frame, including all channels
	FrameBytes int
	// The number of output audio channels before the rendering module.
	NumChannels int
	// Decoder internal members.
	//Sampling rate in Hz without SBR divided by a (ELD) downscale factor if present.
	AacSampleRate int
	// MPEG-2 profile
	Profile int
	// Audio Object Type (from ASC)
	AOT AudioObjectType
	// Channel configuration
	ChannelConfig int
	// Instantaneous bit rate.
	BitRate int
	// Samples per frame for the AAC core (from ASC) divided by a (ELD) downscale factor if present.
	AacSamplesPerFrame int
	// The number of audio channels after AAC core processing (before PS or MPS processing).
	AacNumChannels int
	// Extension Audio Object Type (from ASC)
	ExtAot AudioObjectType
	// Extension sampling rate in Hz (from ASC) divided by a (ELD) downscale factor if present.
	ExtSamplingRate int
	// The number of samples the output is additionally delayed by the decoder.
	OutputDelay int
	// Copy of internal flags. Only to be written by the decoder, and only to be read externally.
	Flags uint
	// epConfig level (from ASC)
	// only level 0 supported, -1 means no ER (e. g. AOT=2, MPEG-2 AAC, etc.)
	EpConfig int8
	// This integer will reflect the estimated amount of lost access units in case aacDecoder_DecodeFrame()
	// returns AAC_DEC_TRANSPORT_SYNC_ERROR.
	NumLostAccessUnits int64
	// This is the number of total bytes that have passed through the decoder.
	NumTotalBytes int64
	// This is the number of total bytes that were considered with errors from numTotalBytes.
	NumBadBytes int64
	// This is the number of total access units that have passed through the decoder.
	NumTotalAccessUnits int64
	// This is the number of total access units that were considered with errors from numTotalBytes.
	NumBadAccessUnits int64
	// DRC program reference level.
	DrcProgRefLev int8
	// DRC presentation mode.
	DrcPresMode int8
}

type Decoder struct {
	ph         C.HANDLE_AACDECODER
	info       *StreamInfo
	remainData []byte
}

// NewDecoder
func NewDecoder(config *DecoderConfig) (*Decoder, error) {
	config = populateDecConfig(config)

	dec := &Decoder{}
	dec.ph = C.aacDecoder_Open(C.TRANSPORT_TYPE(config.TransportFmt), 1)
	if dec.ph == nil {
		return nil, errors.New("create acc decoder failed")
	}

	var errNo C.AAC_DECODER_ERROR = C.AAC_DEC_OK
	defer func() {
		if errNo != C.AAC_DEC_OK {
			C.aacDecoder_Close(dec.ph)
		}
	}()

	if config.PcmDualChannelOutputMode != PcmDualChannelLeaveBoth {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_DUAL_CHANNEL_OUTPUT_MODE,
			C.int(config.PcmDualChannelOutputMode)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmOutputChannelMappingMpeg {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_OUTPUT_CHANNEL_MAPPING,
			C.int(0)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmLimiterMode != PcmLimiterAutoConfig {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_LIMITER_ENABLE,
			C.int(config.PcmLimiterMode-1)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmLimiterAttackTime > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_LIMITER_ATTACK_TIME,
			C.int(config.PcmLimiterAttackTime)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmLimiterReleasTime > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_LIMITER_RELEAS_TIME,
			C.int(config.PcmLimiterReleasTime)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmMinOutputChannels > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_MIN_OUTPUT_CHANNELS,
			C.int(config.PcmMinOutputChannels)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.PcmMaxOutputChannels > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_PCM_MAX_OUTPUT_CHANNELS,
			C.int(config.PcmMaxOutputChannels)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.MetadataProfile != MdProfileMpegStandard {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_METADATA_PROFILE,
			C.int(config.MetadataProfile)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.MetadataExpiryTime > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_METADATA_EXPIRY_TIME,
			C.int(config.MetadataExpiryTime)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.ConcealMethod != ConcealSpectralMuting {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_CONCEAL_METHOD,
			C.int(config.ConcealMethod)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.DrcBoostFactor > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_BOOST_FACTOR,
			C.int(config.DrcBoostFactor)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.DrcAttenuationFactor > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_ATTENUATION_FACTOR,
			C.int(config.DrcAttenuationFactor)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.DrcReferenceLevel > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_REFERENCE_LEVEL,
			C.int(config.DrcReferenceLevel)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.EnableDrcHeavyCompression {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_HEAVY_COMPRESSION,
			C.int(1)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.DrcDefaultPresentationMode != DrcParameterHandlingDisabled {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_DEFAULT_PRESENTATION_MODE,
			C.int(config.DrcDefaultPresentationMode-1)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.DrcEncTargetLevel > 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_DRC_ENC_TARGET_LEVEL,
			C.int(config.DrcEncTargetLevel)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.UnidrcSetEffect != 0 {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_UNIDRC_SET_EFFECT,
			C.int(config.UnidrcSetEffect)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.EnableUnidrcAlbumMode {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_UNIDRC_ALBUM_MODE,
			C.int(1)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}
	if config.QmfLowpowerMode != QmfLowpowerInternal {
		if errNo = C.aacDecoder_SetParam(dec.ph, C.AAC_QMF_LOWPOWER,
			C.int(config.QmfLowpowerMode)); errNo != C.AAC_DEC_OK {
			return nil, getDecError(errNo)
		}
	}

	return dec, nil
}

// EstimateOutBufBytes returns the recommended output buffer size for decoding.
// The buffer should be large enough to hold multiple AAC frames.
func (dec *Decoder) EstimateOutBufBytes(nFrames int) int {
	// 1 frame: 1024 samples * 8 channels * 2 bytes = 16384 bytes
	return (1024 * 8 * 2) * nFrames
}

// Decode decodes AAC audio data to PCM format.
// It handles partial frames internally and may buffer incomplete data.
// Returns the number of decoded PCM bytes written to output buffer.
func (dec *Decoder) Decode(in, out []byte) (n int, err error) {
	szIn := len(in)
	szOut := len(out)
	if szIn == 0 {
		return 0, errors.New("input buffer is empty")
	}
	if szOut < dec.EstimateOutBufBytes(EstimateFrames) {
		return 0, errors.New("output buffer size is not enough")
	}

	if len(dec.remainData) > 0 {
		in = append(dec.remainData, in...)
		szIn = len(in)
		dec.remainData = nil
	}

	inPtr := (*C.uchar)(unsafe.Pointer(&in[0]))
	inLen := C.uint(szIn)
	bytesValid := inLen
	outPtr := (*C.uchar)(unsafe.Pointer(&out[0]))
	outLen := C.INT(szOut)
	bytesDecoded := C.uint(0)

	if errNo := C.aacDecoder_DecodeWrapped(dec.ph, inPtr, inLen, outPtr, outLen, &bytesDecoded, &bytesValid); errNo != C.AAC_DEC_OK {
		return 0, getDecError(errNo)
	}

	if bytesValid > 0 {
		dec.remainData = append(dec.remainData, in[szIn-int(bytesValid):]...)
	}

	// Get stream info after first successful decode
	if dec.info == nil && bytesDecoded > 0 {
		if dec.info, err = getStreamInfo(dec.ph); err != nil {
			return 0, err
		}
	}

	return int(bytesDecoded), nil
}

// ClearBuffer clears the decoder's internal buffer.
// This is useful when seeking or switching between streams.
func (dec *Decoder) ClearBuffer() error {
	dec.remainData = nil
	return getDecError(C.aacDecoder_SetParam(dec.ph, C.AAC_TPDEC_CLEAR_BUFFER, C.int(1)))
}

// Close releases all resources associated with the decoder.
func (dec *Decoder) Close() {
	if dec.ph != nil {
		C.aacDecoder_Close(dec.ph)
		dec.ph = nil
		dec.remainData = nil
		dec.info = nil
	}
}

// ConfigRaw configures the decoder with raw AAC configuration data (AudioSpecificConfig).
// This is required when using raw AAC transport format (TtMp4Raw).
func (dec *Decoder) ConfigRaw(conf []byte) error {
	if len(conf) == 0 {
		return errors.New("raw config should not be empty")
	}
	confPtr := (*C.uchar)(unsafe.Pointer(&conf[0]))
	length := C.uint(len(conf))
	errNo := C.aacDecoder_ConfigRawWrapped(dec.ph, confPtr, length)
	if errNo != C.AAC_DEC_OK {
		return getDecError(errNo)
	}
	return nil
}

// GetStreamInfo returns stream information after at least one frame has been decoded.
// Returns an error if no frames have been decoded yet.
func (dec *Decoder) GetStreamInfo() (*StreamInfo, error) {
	if dec.info == nil {
		return nil, errors.New("decoder has not decoded first frame yet")
	}
	return dec.info, nil
}

// GetRawStreamInfo retrieves stream information directly from the decoder.
// This can be called after ConfigRaw() without decoding any frames.
func (dec *Decoder) GetRawStreamInfo() (*StreamInfo, error) {
	return getStreamInfo(dec.ph)
}

func getStreamInfo(ph C.HANDLE_AACDECODER) (*StreamInfo, error) {
	originInfo := C.aacDecoder_GetStreamInfo(ph)
	if originInfo == nil {
		return nil, errors.New("get stream info failed")
	}

	si := &StreamInfo{
		SampleRate:          int(originInfo.sampleRate),
		FrameLength:         int(originInfo.frameSize),
		NumChannels:         int(originInfo.numChannels),
		AacSampleRate:       int(originInfo.aacSampleRate),
		Profile:             int(originInfo.profile),
		AOT:                 AudioObjectType(originInfo.aot),
		ChannelConfig:       int(originInfo.channelConfig),
		BitRate:             int(originInfo.bitRate),
		AacSamplesPerFrame:  int(originInfo.aacSamplesPerFrame),
		AacNumChannels:      int(originInfo.aacNumChannels),
		ExtAot:              AudioObjectType(originInfo.extAot),
		ExtSamplingRate:     int(originInfo.extSamplingRate),
		OutputDelay:         int(originInfo.outputDelay),
		Flags:               uint(originInfo.flags),
		EpConfig:            int8(originInfo.epConfig),
		NumLostAccessUnits:  int64(originInfo.numLostAccessUnits),
		NumTotalBytes:       int64(originInfo.numTotalBytes),
		NumBadBytes:         int64(originInfo.numBadBytes),
		NumTotalAccessUnits: int64(originInfo.numTotalAccessUnits),
		NumBadAccessUnits:   int64(originInfo.numBadAccessUnits),
		DrcProgRefLev:       int8(originInfo.drcProgRefLev),
		DrcPresMode:         int8(originInfo.drcPresMode),
	}

	// fdk-aac only supports 16 bits (2 bytes) depth.
	si.FrameBytes = si.FrameLength * si.NumChannels * SampleBitDepth / 8
	return si, nil
}

func populateDecConfig(c *DecoderConfig) *DecoderConfig {
	if c == nil {
		c = &DecoderConfig{}
	}

	return c
}
