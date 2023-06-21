package feature

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
	"go.mau.fi/whatsmeow/types"
)

var (
	GPTMap        = make(map[types.JID][]map[string]string)
	apikey string = ""
	proxy  string = ""
	model  string = "gpt-3.5-turbo"
)

type GptConfig struct {
	api  string
	prox string
	mod  string
}

func transporter(req *http.Request, proxyURL string) (*http.Response, error) {
	var transport *http.Transport
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client := &http.Client{Transport: transport}
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		return res, nil
	} else {
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func GPTConfig(Model string, Api string, Proxy string) {
	if Api == "<YOUR OPENAI APIKEY>" {
		apikey = ""
	}
	model = model
	proxy = proxy
	apikey = Api
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
	if apikey == "" {
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
		"model":    model,
		"messages": GPTMap[sender],
	})

	if err != nil {
		return "", err
	}
	var payload = bytes.NewBuffer(requestBody)
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", payload)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apikey)
	response, err := transporter(request, proxy)
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
		return "", errors.New("fail to get response Chat GPT Turbo!")
	}
}
