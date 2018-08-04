# pollydent

pollydent is a wrapper of text-to-speech for Amazon Polly and Google Cloud Text-To-Speech

# Requirement

pollydent needs AWS account, and GCP account. Please sign up AWS, and GCP.
if you want to use Google Cloud Text-To-Speech, please install `gcloud` in your PC.
Please install SoX if you are using Linux or OS X.

# Install

```
$ go get github.com/bamchoh/pollydent
```

# Usage

```
package main

import (
	"log"
	polly "github.com/bamchoh/pollydent"
)

func main() {
	p := polly.NewPollydent(
		"<ACCESS_KEY>",
		"<SECRET_KEY>",
		nil,
	)
	p.ReadAloud("こんにちは世界")
}
```

# ACCESS_KEY, SECRET_KEY

If you have no these keys, please create according to https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html

# How to enable Google Cloud Text-To-Speech

Please see Quick Start guide, https://cloud.google.com/text-to-speech/docs/quickstart-protocol

