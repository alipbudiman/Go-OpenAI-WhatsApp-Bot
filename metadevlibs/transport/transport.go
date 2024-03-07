package transport

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	obj "gowagpt/metadevlibs/object"
)

func Transporter(req *http.Request, proxyURL string) (*http.Response, error) {
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

func Download(url string, folderPath string, customName string) (string, error) {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		os.MkdirAll(folderPath, 0755)
	}
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var fileName string
	if customName != "" {
		fileName = customName
	} else {
		fileName = path.Base(url)
	}
	filePath := filepath.Join(folderPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}
	if strings.Contains(filePath, ".mp3") {
		opusN := obj.OpusEncode(filePath)
		time.Sleep(time.Second * 1)
		err := os.Remove(filePath)
		if err != nil {
			return "", err
		}
		return opusN, nil
	}
	return filePath, nil
}
