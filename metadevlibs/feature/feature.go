package feature

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	tt "gowagpt/metadevlibs/transport"

	"github.com/tidwall/gjson"
	"go.mau.fi/whatsmeow/types"
)

var (
	GPTMap        = make(map[types.JID][]map[string]string)
	Apikey string = ""
	Proxy  string = ""
	Model  string = "gpt-3.5-turbo"
)

func GPTConfig(model string, api string, proxy string) {
	Model = model
	Proxy = proxy
	Apikey = api
}

func ChatGPTReset(sender types.JID) bool {
	if len(GPTMap[sender]) < 1 {
		return false
	}
	delete(GPTMap, sender)
	return true
}

func ChatGPTHistory(sender types.JID) string {
	if len(GPTMap[sender]) < 1 {
		return "You dont have convertation"
	}
	chatbot := "\n*GPT Converataion*"
	chatbot += fmt.Sprintf("\nUser: %s", strings.Split((fmt.Sprintf("%v", sender)), "@")[0])
	chatbot += fmt.Sprintf("\nCount: %d", (len(GPTMap[sender])-1)/2)
	return chatbot
}

func ChatGPT(sender types.JID, user_chat string) (string, error) {
	if Apikey == "" {
		return "Please input OpenAI Apikey, get apikey here https://platform.openai.com/account/api-keys", nil
	}
	if len(GPTMap[sender]) < 1 {
		newMap := []map[string]string{
			{
				"role":    "system",
				"content": user_chat,
			},
			{
				"role":    "user",
				"content": user_chat,
			},
		}
		GPTMap[sender] = append(GPTMap[sender], newMap...)
	} else {
		newMap := map[string]string{
			"role":    "user",
			"content": user_chat,
		}
		GPTMap[sender] = append(GPTMap[sender], newMap)
	}
	requestBody, err := json.Marshal(map[string]interface{}{
		"model":    Model,
		"messages": GPTMap[sender],
	})

	if err != nil {
		return "", err
	}
	var payload = bytes.NewBuffer(requestBody)
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", payload)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+Apikey)
	response, err := tt.Transporter(request, Proxy)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	var res = gjson.Get(string(body), "choices.0.message.content").String()
	if res != "" {
		newMap := map[string]string{
			"role":    gjson.Get(string(body), "choices.0.message.role").String(),
			"content": res,
		}
		GPTMap[sender] = append(GPTMap[sender], newMap)
		return res, nil
	} else {
		return "", errors.New("fail to get response Chat GPT Turbo")
	}
}

func DallE(prompt string, imageLoad int, size string) ([]string, error) {
	if Apikey == "" {
		return nil, errors.New("please input OpenAI Apikey, get apikey here https://platform.openai.com/account/api-keys")
	}
	requestBody, err := json.Marshal(map[string]interface{}{
		"prompt": prompt,
		"n":      imageLoad,
		"size":   size,
	})
	if err != nil {
		return nil, err
	}
	var payload = bytes.NewBuffer(requestBody)
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations", payload)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+Apikey)
	response, err := tt.Transporter(request, Proxy)
	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		rejected := gjson.Get(string(body), "error.message").String()
		if rejected != "" {
			return nil, errors.New(strings.ToLower(rejected))
		}
		created := gjson.Get(string(body), "created").String()
		if created == "" {
			return nil, errors.New("oops.. something when wrong!")
		}
		result := gjson.Get(string(body), "data.#.url")
		var imgData []string
		for _, url := range result.Array() {
			imgData = append(imgData, url.String())
		}
		return imgData, nil
	}
}
