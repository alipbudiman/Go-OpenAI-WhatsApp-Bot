package helper

import (
	"errors"
	"fmt"
	"image/jpeg"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"

	"github.com/nickalie/go-webpbin"
	"go.mau.fi/whatsmeow/types"
)

func TrackEvents(evt interface{}) {
	fmt.Println("\033[34m\n", "[(>)", evt, "]", reflect.TypeOf(evt), "\033[0m")
}

func WriteDisplayMenu(from_dm bool) string {
	h := "┎───「 MENU 」"
	h += "\n⊶ help"
	h += "\n⊶ ping"
	h += "\n⊶ help"
	h += "\n⊶ hello world"
	h += "\n⊶ send image"
	h += "\n⊶ send video"
	h += "\n⊶ chat gpt: `question`"
	h += "\n⊶ dalle draw: `question`"
	h += "\n⊶ group broadcast: `message`"
	if !from_dm {
		h += "\n⊶ say: `query`"
		h += "\n⊶ tag all"
		h += "\n⊶ reader <on/off>"
		h += "\n⊶ anti unsend <on/off>"
	}
	return h
}

func FileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

func ConvertJPEtoWEBP(path string) (bool, string) {
		webp_path := strings.Split(path, ".jpe")[0]+".webp"
		cmd := webpbin.NewCWebP().
		Quality(70).
		InputFile(path).
		OutputFile(webp_path)

	if err := cmd.Run(); err != nil {
		return false, path
	}
	return true, webp_path
}

func SenderJIDConvert(jid types.JID) (types.JID, bool) {
	j := fmt.Sprintf("%v", jid)
	x := strings.Split(j, "@")
	y := strings.Split(x[0], ".") 
	z := y[0] + "@" + x[1]
	jid, ok := ParseJIDUser(z)
	if !ok {
		return jid, false
	}
	return jid, true
}

func ConvertJID(args string) (types.JID, bool) {
	jid, ok := ParseJIDUser(args)
	if ok {
		return jid, true
	}
	return jid, false
}

func ParseJIDUser(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			fmt.Println("Fail JID %s: %v", arg, err)
			return recipient, false
		} else if recipient.User == "" {
			fmt.Println("Fail JID %s: no specified", arg)
			return recipient, false
		}
		return recipient, true
	}
}

func RemoveMyJID(listdata []string, myJID types.JID) []string {
	dataArry := listdata
	for _, x := range dataArry {
		if fmt.Sprintf("%v", myJID) == x {
			dataArry = Remove(dataArry, x)
		}
	}
	return dataArry
}

func Remove(s []string, r string) []string {
	new := make([]string, len(s))
	copy(new, s)
	for i, v := range new {
		if v == r {
			return append(new[:i], new[i+1:]...)
		}
	}
	return s
}

func MentionFormat(jid string) string {
	m := strings.Split(jid, ".")[0]
	m = strings.ReplaceAll(m, "@", "")
	return "@" + strings.ReplaceAll(m, "s", "")
}

func LooperMessage(message string, cut_after int) []string {
	var response []string
	k := len(message) / cut_after
	for aa := 0; aa <= k; aa++ {
		start := aa * cut_after
		end := (aa + 1) * cut_after
		if end > len(message) {
			end = len(message)
		}
		message := message[start:end]
		response = append(response, message)
	}
	return response
}

func InArray[T any](s []T, r T) bool {
	for _, x := range s {
		if reflect.DeepEqual(x, r) {
			return true
		}
	}
	return false
}

func RandomStrings(length int) string {
	rand.Seed(time.Now().UnixNano()) // Inisialisasi nilai seed
	var letterBytes string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))] // Pilih karakter acak dari letterBytes
	}
	randomString := string(b)

	return randomString
}

func ConvertJPEtoJPG(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	newFileName := strings.ReplaceAll(path, ".jpe", ".jpg")
	newfile, err := os.Create(newFileName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer newfile.Close()
	err = jpeg.Encode(newfile, img, &jpeg.Options{Quality: 90})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return newFileName, nil
}

func ConvertF4VtoMP4(inputFile string, outputFile string) (string, error) {
	ch := make(chan bool)
	exec.Command("ffmpeg", "-i", inputFile, outputFile).Run()
	go TrackFileTimeOut(10, outputFile, ch)
	isExist := <-ch
	if isExist {
		return outputFile, nil
	}
	return "", errors.New("error, Fail to convert F4v to MP4")

}

func TrackFileTimeOut(time_in_second int, path string, result chan<- bool) {
	for i := 1; i < time_in_second; i++ {
		time.Sleep(1 * time.Second)
		isXyz := FileExists(path)
		if isXyz {
			result <- true
			break
		}
	}
	result <- false
}
