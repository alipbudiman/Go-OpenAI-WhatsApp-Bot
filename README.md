# Go-ChatGPT-WhatsApp-Bot

[![GO](https://img.shields.io/badge/golang-v1.19.9^-blue)](https://go.dev/)&nbsp;&nbsp;[![UBUNTU](https://img.shields.io/badge/ubuntu-v18.0-orange)](https://releases.ubuntu.com/impish/)&nbsp;&nbsp;[![SOURCE](https://img.shields.io/badge/license-MIT-green)](https://github.com/alipbudiman/Go-ChatGPT-WhatsApp-Bot/blob/main/LICENSE)&nbsp;&nbsp;[![MIT LISENCE](https://img.shields.io/badge/sponsors-WhatsApp-brightgreen)](https://wa.me/6282113791904)

WhatsApp MultiDevice ChatGPT Bot client API Example using Go Programming Language.

![screen shoot](/assets/img/ss.jpg)

The Example of usage [Whatsmeow WhatsApp API](https://github.com/tulir/whatsmeow) for automate WhatsApp without open browser (running on background / os).

# Usable bots for (Example and Inspiration)

*COMING SOON*

# Requirement

- [Go](https://go.dev/) Version >= 1.19.9

- [FFmpeg](https://ffmpeg.org/)

# Add OPEN AI Token

```go
var (
	Log           *logrus.Logger
	Client        *ClientWrapper
	myJID         types.JID
    //---------here----------------
	ChatGPTApikey string = "" // << INSERT YOUR OPEN AI API HERE
    //---------here----------------
	ChatGPTProxy  string = ""
)
```

To get open AI api-key, you can visit open ai platform. [Click here](https://platform.openai.com/account/api-keys)

# Turn off go module

```
$ go env -w GO111MODULE=off
```

# Run

```
$ go build main.go
$ ./main.go
```

or

```
$ go run main.go
```

# Author

- [Alif budiman](https://github.com/alipbudiman)


# Discussion

- [WhatsApp Group](https://chat.whatsapp.com/Gbe7Y7NHpZXEaLoQRc6WpD)