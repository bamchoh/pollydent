package pollydent

import (
	"os"

	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// PollyConfig is configuration structure for Polly
type PollyConfig struct {
	Region   string `yaml:"region"`
	Format   string `yaml:"format"`
	Voice    string `yaml:"voice"`
	TextType string `yaml:"type"`
	Speed    int
}

func defaultConfig() *PollyConfig {
	return &PollyConfig{
		Region:   "us-west-2",
		Format:   "pcm",
		Voice:    "Mizuki",
		TextType: "ssml",
		Speed:    100,
	}
}

func load(r io.Reader) (*PollyConfig, error) {
	var data []byte
	var err error
	pc := defaultConfig()

	data, err = ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, pc)
	if err != nil {
		return nil, err
	}

	return pc, err
}

func Load(filepath string) (*PollyConfig, error) {
	var err error

	f, err := os.Open(filepath)
	if err != nil {
		return defaultConfig(), err
	}
	defer f.Close()

	return load(f)
}
