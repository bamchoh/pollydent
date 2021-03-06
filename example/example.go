package main

import (
	"io"
	"log"
	"os"
	"sync"

	polly "github.com/bamchoh/pollydent"
)

func main() {
	var err error

	config, _ := polly.Load("pollydent.yml")
	var pollydents []*polly.Pollydent

	// Cloud Text To Speech
	var tmp *polly.Pollydent
	tmp, err = polly.NewPollydentWithCloudTextToSpeech(config)
	if err != nil {
		panic(err)
	}
	pollydents = append(pollydents, tmp)

	// Polly
	tmp, err = polly.NewPollydentWithPolly(
		os.Getenv("AWS_ACCESS_KEY"),
		os.Getenv("AWS_SECRET_KEY"),
		config,
	)
	if err != nil {
		panic(err)
	}
	pollydents = append(pollydents, tmp)

	for _, p := range pollydents {
		// Example 1
		// Read Aloud Example
		err = p.ReadAloud("こんにちは世界")
		if err != nil {
			log.Fatal(err)
		}

		// Example 2
		// SendToPolly / Play Example
		params := []polly.SpeechParams{
			polly.SpeechParams{"こんばんわ世界", "Mizuki", 100},
			polly.SpeechParams{"おはようございます世界", "Mizuki", 200},
			polly.SpeechParams{"Hello World", "Joey", 100},
		}

		var strm io.Reader
		var wg sync.WaitGroup
		for _, param := range params {
			wg.Add(1)
			strm, err = p.SendToServer(param)
			if err != nil {
				log.Fatal(err)
			}
			go func(strm2 io.Reader) {
				p.Play(strm2)
				wg.Done()
			}(strm)
		}

		wg.Wait()
	}
}
