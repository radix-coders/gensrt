package gensrt

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	speech "cloud.google.com/go/speech/apiv1"
	"google.golang.org/api/option"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

//EncType ...
type EncType int32

//EncType values ...
const (
	// Not specified.
	UNSPECIFIED EncType = 0
	// Uncompressed 16-bit signed little-endian samples (Linear PCM).
	LINEAR16 EncType = 1
	// `FLAC` (Free Lossless Audio Codec)
	FLAC EncType = 2
	// 8-bit samples that compand 14-bit audio samples using G.711 PCMU/mu-law.
	MULAW EncType = 3
	// Adaptive Multi-Rate Narrowband codec. `sample_rate_hertz` must be 8000.
	AMR EncType = 4
	// Adaptive Multi-Rate Wideband codec. `sample_rate_hertz` must be 16000.
	AMR_WB EncType = 5
	// Opus encoded audio frames in Ogg container
	// `sample_rate_hertz` must be one of 8000, 12000, 16000, 24000, or 48000.
	OGG_OPUS EncType = 6
	// Although the use of lossy encodings is not recommended, if a very low
	// bitrate encoding is required, `sample_rate_hertz` must be 16000.
	SPEEX_WITH_HEADER_BYTE EncType = 7
)

//LangCodeType ...
type LangCodeType string

//LangCodeType values ...
const (
	EnUS LangCodeType = "en-US"
)

const srtFormat = `%d
%02v:%02v:%02v,%03v --> %02v:%02v:%02v,%03v
<font color="#808080">%v</font>
`

//Config ...
type Config struct {
	audioPath    string
	languageCode LangCodeType
	encoding     EncType
	sampleRateHz int32
	client       *speech.Client
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

//NewConfig returns config reference for further API usage.
func NewConfig(audioPath, credentialsFilePath string) (cfg *Config, err error) {

	if !fileExists(audioPath) {
		err = fmt.Errorf("Audio File not exist")
		return
	}

	if !fileExists(credentialsFilePath) {
		err = fmt.Errorf("Credentials File not exist")
		return
	}

	cfg = &Config{
		audioPath:    audioPath,
		languageCode: EnUS,
		encoding:     LINEAR16,
		sampleRateHz: 16000,
	}

	ctx := context.Background()
	cliOpt := option.WithCredentialsFile(credentialsFilePath)
	// Creates a client.
	cfg.client, err = speech.NewClient(ctx, cliOpt)
	return
}

//SetLanguage ...
func (cfg *Config) SetLanguage(languageCode LangCodeType) {
	cfg.languageCode = languageCode
}

//SetEncoding ...
func (cfg *Config) SetEncoding(encoding EncType) {
	cfg.encoding = encoding
}

//SetSampleRate ...
func (cfg *Config) SetSampleRate(sampleRateHz int32) {
	cfg.sampleRateHz = sampleRateHz
}

//ProcessRequest generates srt file of audio clip
func (cfg *Config) ProcessRequest() (err error) {
	var (
		resp *speechpb.LongRunningRecognizeResponse
	)
	url, _ := url.ParseRequestURI(cfg.audioPath)
	if url != nil {
		resp, err = cfg.speechToText(true)
	} else {
		resp, err = cfg.speechToText(false)
	}
	err = generateSrt(resp)
	return err
}

func generateSrt(resp *speechpb.LongRunningRecognizeResponse) error {
	// Prints the results.
	count := 1
	f, err := os.Create("output.srt")
	if err != nil {
		return err
	}
	defer f.Close()
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			sHr, sHrRem := alt.Words[0].StartTime.Seconds/3600, alt.Words[0].StartTime.Seconds%3600
			sMin, sMinRem := sHrRem/60, sHrRem%60
			sSec, sNs := sMinRem/60, alt.Words[0].StartTime.Nanos
			eHr, eHrRem := alt.Words[0].EndTime.Seconds/3600, alt.Words[0].EndTime.Seconds%3600
			eMin, eMinRem := eHrRem/60, eHrRem%60
			eSec, eNs := eMinRem/60, alt.Words[0].EndTime.Nanos
			srtLine := fmt.Sprintf(srtFormat, count, sHr, sMin, sSec, sNs, eHr, eMin, eSec, eNs, alt.Transcript)
			count++
			f.WriteString(srtLine)
		}
	}
	return nil
}

func (cfg *Config) speechToText(isURI bool) (*speechpb.LongRunningRecognizeResponse, error) {
	// Send the contents of the audio file with the encoding and
	// and sample rate information to be transcripted.
	req := &speechpb.LongRunningRecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:              speechpb.RecognitionConfig_AudioEncoding(cfg.encoding),
			SampleRateHertz:       int32(cfg.sampleRateHz),
			LanguageCode:          string(cfg.languageCode),
			EnableWordTimeOffsets: true,
		},
	}

	if isURI {
		req.Audio = &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Uri{Uri: cfg.audioPath},
		}
	} else {
		data, err := ioutil.ReadFile(cfg.audioPath)
		if err != nil {
			return nil, err
		}
		req.Audio = &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: data},
		}
	}
	ctx := context.Background()
	op, err := cfg.client.LongRunningRecognize(ctx, req)
	if err != nil {
		return nil, err
	}
	resp, err := op.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
