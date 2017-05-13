package pollydent

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/bamchoh/pollydent/player"

	"errors"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// PollyConfig is configuration structure for Polly
type PollyConfig struct {
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Region    string `yaml:"region"`
	Format    string `yaml:"format"`
	Voice     string `yaml:"voice"`
	TextType  string `yaml:"text_type"`
	Polly     *polly.Polly
}

// Polly is structure to manage read aloud
type Polly struct {
	Logger    *log.Logger
	Config    *PollyConfig
	Counter   int
	SendMutex *sync.Mutex
	PlayMutex *sync.Mutex
}

func (p *Polly) sendAws(msg string, speed int) (resp *polly.SynthesizeSpeechOutput, err error) {
	p.SendMutex.Lock()
	defer p.SendMutex.Unlock()
	packedMsg := `<speak><prosody rate="` + strconv.Itoa(speed) + `%"><![CDATA[` + msg + `]]></prosody></speak>`

	resp, err = p.SynthesizeSpeech(packedMsg)
	return
}

func (p *Polly) play(resp *polly.SynthesizeSpeechOutput) (err error) {
	p.PlayMutex.Lock()
	defer p.PlayMutex.Unlock()
	err = player.Play(resp.AudioStream)

	if err != nil {
		p.Logger.Println(err)
	}
	return
}

// ReadAloud reads aloud msg by Polly
func (p *Polly) ReadAloud(msg string) (err error) {
	p.Counter++
	defer func() { p.Counter-- }()

	if p.Counter > 5 {
		p.Logger.Println("Skipped : ", msg)
		return
	}

	speed := 100
	speed += 20 * (p.Counter - 1)

	resp, err := p.sendAws(msg, speed)
	if err != nil {
		p.Logger.Println(err)
		return
	}
	p.play(resp)
	return
}

// NewPolly news Polly structure
func NewPolly(logger *log.Logger, loadFile string) (*Polly, error) {
	var err error

	p := &Polly{
		Logger:    logger,
		SendMutex: new(sync.Mutex),
		PlayMutex: new(sync.Mutex),
	}

	basedir := filepath.Dir(os.Args[0])
	filepath := filepath.Join(basedir, loadFile)
	f, err := os.Open(filepath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer f.Close()

	p.Config, err = p.load(f)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return p, err
}

func (p *Polly) load(r io.Reader) (*PollyConfig, error) {
	var data []byte
	var err error
	var pc PollyConfig

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &pc)
	if err != nil {
		return nil, err
	}

	errTxt := ""
	if pc.AccessKey == "" {
		errTxt += "value of access_key does not set in setting file. "
	}

	if pc.SecretKey == "" {
		errTxt += "value of secret_key does not set in setting file. "
	}

	if errTxt != "" {
		return nil, errors.New(errTxt)
	}

	if pc.Region == "" {
		pc.Region = "us-west-2"
	}

	if pc.Format == "" {
		pc.Format = "mp3"
	}

	if pc.Voice == "" {
		pc.Voice = "Mizuki"
	}

	if pc.TextType == "" {
		pc.TextType = "ssml"
	}

	pc.Polly, err = p.initPolly(&pc)
	if err != nil {
		return nil, err
	}

	return &pc, err
}

func (p *Polly) initPolly(pc *PollyConfig) (*polly.Polly, error) {
	creds := credentials.NewStaticCredentials(pc.AccessKey, pc.SecretKey, "")
	sess := session.New(&aws.Config{Credentials: creds})
	return polly.New(sess, aws.NewConfig().WithRegion(pc.Region)), nil
}

// SynthesizeSpeech is call aws-polly SynthesizeSpeech
func (p *Polly) SynthesizeSpeech(text string) (*polly.SynthesizeSpeechOutput, error) {
	pc := p.Config
	params := &polly.SynthesizeSpeechInput{
		OutputFormat: aws.String(pc.Format),
		Text:         aws.String(text),
		TextType:     aws.String(pc.TextType),
		VoiceId:      aws.String(pc.Voice),
	}
	return pc.Polly.SynthesizeSpeech(params)
}
