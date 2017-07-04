package pollydent

import (
	"os"

	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// PollyConfig is configuration structure for Polly
type PollyConfig struct {
	Region   string
	format   string
	Voice    string `yaml:"voice"`
	textType string
	Speed    int
}

func load(r io.Reader) (*PollyConfig, error) {
	var data []byte
	var err error
	pc := PollyConfig{
		Region:   "us-west-2",
		format:   "pcm",
		Voice:    "Mizuki",
		textType: "ssml",
		Speed:    100,
	}

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &pc)
	if err != nil {
		return nil, err
	}

	return &pc, err
}

func Load(filepath string) (*PollyConfig, error) {
	var err error

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return load(f)
}
