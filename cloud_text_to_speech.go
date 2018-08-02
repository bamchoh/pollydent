package pollydent

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hajimehoshi/oto"
)

type Request struct {
	Input       input       `json:"input"`
	Voice       voice       `json:"voice"`
	AudioConfig audioConfig `json:"audioConfig"`
}

type input struct {
	Text string `json:"text"`
}

type voice struct {
	LanguageCode string `json:"languageCode"`
	Name         string `json:"name"`
	SsmlGender   string `json:"ssmlGender"`
}

type audioConfig struct {
	AudioEncoding   string `json:"audioEncoding"`
	SampleRateHertz int    `json:"sampleRateHertz"`
}

type Response struct {
	AudioContent string `json:"audioContent"`
}

type MP3Wrapper struct {
	io.Reader
}

func (w *MP3Wrapper) Close() error {
	return nil
}

func _main() {
	var err error
	reqData := Request{
		Input: input{
			Text: "Android is a mobile operating system developed by Google, based on the Linux kernel and designed primarily for touchscreen mobile devices such as smartphones and tablets.",
		},
		Voice: voice{
			LanguageCode: "en-US",
			Name:         "en-US-Wavenet-C",
			SsmlGender:   "FEMALE",
		},
		AudioConfig: audioConfig{
			AudioEncoding:   "LINEAR16",
			SampleRateHertz: 16000,
		},
	}
	body, err := json.Marshal(reqData)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("gcloud", "auth", "application-default", "print-access-token")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	token := strings.Split(string(out), "\r\n")[0]

	req, err := http.NewRequest(
		"POST",
		"https://texttospeech.googleapis.com/v1beta1/text:synthesize",
		bytes.NewReader(body),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var resData Response
	err = dec.Decode(&resData)
	if err != nil {
		panic(err)
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(resData.AudioContent))

	audioData, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	p, err := oto.NewPlayer(16000, 1, 2, len(audioData))
	if err != nil {
		panic(err)
	}
	defer p.Close()

	if _, err := p.Write(audioData); err != nil {
		panic(err)
	}
}
