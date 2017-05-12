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
	"os"

	"github.com/bamchoh/pollydent"
)

func main() {
	f, _ := os.Create("pollydent.log")
	logger := log.New(f, "pollydent:", 0)
	p, _ := pollydent.NewPolly(logger, "pollydent.yml")
	p.ReadAloud("こんにちは世界")
}
```

# Configuration YAML file

pollydent needs YAML file for accessing to Amazon Polly. Mandatory configuration is here:
```
access_key: "<ACCESS KEY>"
secret_key: "<SECRET KEY>"
```

If you have no these keys, please create according to https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html

