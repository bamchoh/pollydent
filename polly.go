package pollydent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/hajimehoshi/oto"

	"io"
)

type SpeechParams struct {
	Message string
	Voice   string
	Speed   int
}

type Speaker interface {
	Send(SpeechParams) (io.Reader, error)
}

type PollySpeaker struct {
	config *PollyConfig
	sess   *session.Session
}

func (p *PollySpeaker) Send(config SpeechParams) (io.Reader, error) {
	var err error

	if config.Speed == 0 {
		config.Speed = p.config.Speed
	}

	if config.Voice == "" {
		config.Voice = p.config.Voice
	}

	text := `<speak><prosody rate="` + strconv.Itoa(config.Speed) + `%"><![CDATA[` + config.Message + `]]></prosody></speak>`

	pol := polly.New(p.sess, aws.NewConfig().WithRegion(p.config.Region))

	params := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(p.config.Format),
		Text:         aws.String(text),
		TextType:     aws.String(p.config.TextType),
		VoiceId:      aws.String(config.Voice),
	}

	resp, err := pol.SynthesizeSpeech(params)
	if err != nil {
		return nil, err
	}
	return resp.AudioStream, nil
}

type GCTTSSpeaker struct {
	config *PollyConfig
	token  string
}

func (p *GCTTSSpeaker) Send(config SpeechParams) (io.Reader, error) {
	var err error

	if config.Speed == 0 {
		config.Speed = p.config.Speed
	}

	if config.Voice == "" {
		config.Voice = p.config.Voice
	}

	text := `<speak><prosody rate="` + strconv.Itoa(config.Speed) + `%"><![CDATA[` + config.Message + `]]></prosody></speak>`

	var v voice
	switch config.Voice {
	case "Mizuki":
		v = voice{
			LanguageCode: "ja-JP",
			Name:         "ja-JP-Wavenet-A",
			SsmlGender:   "FEMALE",
		}
	default:
		v = voice{
			LanguageCode: "en-US",
			Name:         "en-US-Wavenet-C",
			SsmlGender:   "FEMALE",
		}
	}

	reqData := Request{
		Input: input{
			SSML: text,
		},
		Voice: v,
		AudioConfig: audioConfig{
			AudioEncoding:   "LINEAR16",
			SampleRateHertz: 16000,
		},
	}
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	token := getToken()

	req, err := http.NewRequest(
		"POST",
		"https://texttospeech.googleapis.com/v1beta1/text:synthesize",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var resData Response
	err = dec.Decode(&resData)
	if err != nil {
		return nil, err
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(resData.AudioContent))

	return reader, nil
}

// Pollydent is structure to manage read aloud
type Pollydent struct {
	playMutex   *sync.Mutex
	audioConfig AudioConfig
	speaker     Speaker
}

func NewPollydentWithCloudTextToSpeech(config *PollyConfig) (*Pollydent, error) {
	token := getToken()
	return &Pollydent{
		playMutex:   new(sync.Mutex),
		audioConfig: &GCTTSAudioConfig{},
		speaker:     &GCTTSSpeaker{config, token},
	}, nil
}

// NewPollydent news Polly structure
func NewPollydentWithPolly(accessKey, secretKey string, config *PollyConfig) (*Pollydent, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("Access key or Secret key are invalid")
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess := session.New(&aws.Config{Credentials: creds})

	if config == nil {
		config = defaultConfig()
	}

	return &Pollydent{
		playMutex:   new(sync.Mutex),
		audioConfig: &PollyAudioConfig{},
		speaker:     &PollySpeaker{config, sess},
	}, nil
}

func (p *Pollydent) Play(reader io.Reader) (err error) {
	p.playMutex.Lock()
	defer p.playMutex.Unlock()

	totalData := make([]byte, 0)
	for {
		var n int
		data := make([]byte, 65535)
		if n, err = reader.Read(data); err != nil {
			if err != io.EOF {
				return
			}
			totalData = append(totalData, data[:n]...)
			break
		} else {
			totalData = append(totalData, data[:n]...)
		}
	}

	player, err := oto.NewPlayer(
		p.audioConfig.SampleRate(),
		p.audioConfig.NumOfChanel(),
		p.audioConfig.ByteParSample(),
		len(totalData))
	if err != nil {
		return
	}
	defer player.Close()

	timeCh := make(chan int, 1)

	go func() {
		sr := p.audioConfig.SampleRate()
		noc := p.audioConfig.NumOfChanel()
		b := p.audioConfig.ByteParSample()
		t := time.Second * time.Duration(1+len(totalData)/(sr*noc*b))
		time.Sleep(t)
		timeCh <- 1
	}()

	if _, err = player.Write(totalData); err != nil {
		return
	}

	<-timeCh

	return
}

func (p *Pollydent) SendToServer(param SpeechParams) (io.Reader, error) {
	return p.speaker.Send(param)
}

// ReadAloud reads aloud msg by Polly
func (p *Pollydent) ReadAloud(msg string) (err error) {
	if msgLen := len([]rune(msg)); msgLen > 1500 {
		errMsg := "Message size is %d. Please pass with the length of 1500 or less."
		err = fmt.Errorf(errMsg, msgLen)
		return err
	}

	reader, err := p.speaker.Send(SpeechParams{Message: msg})
	if err != nil {
		return
	}
	p.Play(reader)
	return
}
