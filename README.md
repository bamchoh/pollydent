# pollydent

pollydent is a library to speech to text via Amazon Polly.

# Requirement

pollydent needs AWS account. Please sign up AWS.
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

