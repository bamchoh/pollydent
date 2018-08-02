package pollydent

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/hajimehoshi/oto"

	"io"
)

var (
	sampleRate    = 16000
	numOfChanel   = 1
	byteParSample = 2
)

type SpeechParams struct {
	Message string
	Voice   string
	Speed   int
}

// Pollydent is structure to manage read aloud
type Pollydent struct {
	config    *PollyConfig
	playMutex *sync.Mutex
	sess      *session.Session
}

// NewPollydent news Polly structure
func NewPollydent(accessKey, secretKey string, config *PollyConfig) (*Pollydent, error) {
	if accessKey == "" || secretKey == "" {
		return nil, errors.New("Access key or Secret key are invalid")
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	sess := session.New(&aws.Config{Credentials: creds})

	if config == nil {
		config = defaultConfig()
	}

	return &Pollydent{
		config:    config,
		playMutex: new(sync.Mutex),
		sess:      sess,
	}, nil
}
func (p *Pollydent) SendToPolly(config SpeechParams) (io.Reader, error) {
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

	player, err := oto.NewPlayer(sampleRate, numOfChanel, byteParSample, len(totalData))
	if err != nil {
		return
	}
	defer player.Close()

	timeCh := make(chan int, 1)

	go func() {
		t := time.Second * time.Duration(1+len(totalData)/(sampleRate*numOfChanel*byteParSample))
		time.Sleep(t)
		timeCh <- 1
	}()

	if _, err = player.Write(totalData); err != nil {
		return
	}

	<-timeCh

	return
}

// ReadAloud reads aloud msg by Polly
func (p *Pollydent) ReadAloud(msg string) (err error) {
	if msgLen := len([]rune(msg)); msgLen > 1500 {
		errMsg := "Message size is %d. Please pass with the length of 1500 or less."
		err = fmt.Errorf(errMsg, msgLen)
		return err
	}

	strm, err := p.SendToPolly(SpeechParams{Message: msg})
	if err != nil {
		return
	}
	p.Play(strm)
	return
}
