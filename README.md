# Go-OpenAI-WhatsApp-Bot

[![GO](https://img.shields.io/badge/golang-v1.22.1^-blue)](https://go.dev/)&nbsp;&nbsp;[![UBUNTU](https://img.shields.io/badge/ubuntu-v22.4-orange)](https://releases.ubuntu.com/impish/)&nbsp;&nbsp;[![SOURCE](https://img.shields.io/badge/license-MIT-green)](https://github.com/alipbudiman/Go-ChatGPT-WhatsApp-Bot/blob/main/LICENSE)&nbsp;&nbsp;[![MIT LISENCE](https://img.shields.io/badge/sponsors-WhatsApp-brightgreen)](https://wa.me/6282113791904)

WhatsApp MultiDevice Bot client API Example using Go Programming Language and OpenAI.

The Example of usage [Whatsmeow WhatsApp API](https://github.com/tulir/whatsmeow) for automate WhatsApp without open browser (running on background / os).

# Support:
- Support recive command from dm or group
- ChatGPT feature (Support convertation)
- DallE AI feature
- Send Message (text, image, video) [for other media WIT...]
- Tagall member groups
- Reader check
- Anti Unsend
- Make sticker
- Brouadcast Group
- Edit Message

# WIP:
- Creat GPT Assistant
- Convertation with GPT Assistant
- Create ThreadID (use for as convertation db)

# Chat GPT

![convertation ChatGPT](/assets/img/ss.jpg)

# Dall-E

![DallE draw](/assets/img/dalle.gif)

# Usable bots for (Example and Inspiration)

Free invite & use Whatsapp bot [Wa Nexus](https://wa-nexus.web.app/)

# Requirement

- [Go](https://go.dev/) Version >= 1.19.9

- [FFmpeg](https://ffmpeg.org/)


# Add OpenAI Api-key or Proxy

Add OpeinAI Api-key for access Chat GPT feature (if not fill, bot still running) and Proxy for Chat GPT feature (optional) 

```go
ChatGPTApikey     string = "" // << your apikey here
ChatGPTProxy      string = "" 
```


To get open AI api-key, you can visit open ai platform. [Click here](https://platform.openai.com/account/api-keys)

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
