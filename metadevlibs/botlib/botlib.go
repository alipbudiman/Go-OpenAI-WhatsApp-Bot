package botlib

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
)

type CLIENT struct {
	Client *whatsmeow.Client
}

var (
	Client *CLIENT
)

func (cl *CLIENT) SendMessageV2(evt interface{}, msg *string) {
	v := evt.(*events.Message)
	resp := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: msg,
			ContextInfo: &waProto.ContextInfo{
				StanzaId:    &v.Info.ID,
				Participant: proto.String(v.Info.MessageSource.Sender.String()),
			},
		},
	}
	cl.Client.SendMessage(context.Background(), v.Info.Sender, resp)
}

func (cl *CLIENT) SendTextMessage(jid types.JID, text string) {
	cl.Client.SendMessage(context.Background(), jid, &waProto.Message{Conversation: proto.String(text)})
}

func ResizeImage(imagePath string, thumbnailPath string, h int, w int, q int) string {
	img, err := imaging.Open(imagePath)
	if err != nil {
		fmt.Println("failed to open image:")
	}
	img = imaging.Fit(img, w, h, imaging.Lanczos)
	err = imaging.Save(img, thumbnailPath, imaging.JPEGQuality(q))
	if err != nil {
		fmt.Println("failed to save thumbnail")
	}
	return thumbnailPath

}

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return false
	}
	return true
}

func TrackFileTimeOut(time_in_second int, path string, result chan<- bool) {
	for i := 1; i < time_in_second; i++ {
		time.Sleep(1 * time.Second)
		isXyz := fileExists(path)
		if isXyz {
			result <- true
			break
		}
	}
	result <- false
}

func CreateVideoThumbnail(path string, thumbnailPath string) (string, error) {
	cmd := exec.Command("ffmpeg", "-i", path, "-ss", "00:00:01.000", "-vframes", "1", thumbnailPath)
	err3 := cmd.Run()
	if err3 != nil {
		fmt.Printf("Failed create thumbnail: %v", err3)
		return "", err3
	}
	ch := make(chan bool)
	go TrackFileTimeOut(10, thumbnailPath, ch)
	isExist := <-ch
	if isExist {
		return thumbnailPath, nil
	}
	return "", errors.New("Fail create thumbnail")

}

func (cl *CLIENT) SendImageMessage(jid types.JID, path string, caption string) {
	data, err1 := ioutil.ReadFile(path)
	if err1 != nil {
		fmt.Printf("Failed to read %s: %v", path, err1)
		return
	}
	uploaded, err2 := cl.Client.Upload(context.Background(), data, whatsmeow.MediaImage)
	if err2 != nil {
		fmt.Printf("Failed to upload file: %v", err2)
		return
	}
	ThumbnailPath := strings.ReplaceAll(path, ".jpg", "ThumbnailPath.jpg")
	imgRe := ResizeImage(path, ThumbnailPath, 30, 30, 90)
	imgData, _ := ioutil.ReadFile(imgRe)
	msg := &waProto.Message{ImageMessage: &waProto.ImageMessage{
		JpegThumbnail: imgData,
		Caption:       proto.String(caption),
		Url:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(http.DetectContentType(data)),
		FileEncSha256: uploaded.FileEncSHA256,
		FileSha256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
	}}
	cl.Client.SendMessage(context.Background(), jid, msg)
	os.Remove(ThumbnailPath)
}

func (cl *CLIENT) SendVideoMessage(jid types.JID, path string, caption string) {
	data, err1 := ioutil.ReadFile(path)
	if err1 != nil {
		fmt.Printf("Failed to read %s: %v", path, err1)
		return
	}
	uploaded, err2 := cl.Client.Upload(context.Background(), data, whatsmeow.MediaVideo)
	if err2 != nil {
		fmt.Printf("Failed to upload file: %v", err2)
		return
	}
	thumbnailPath := strings.ReplaceAll(path, ".mp4", "Thumbnail.jpg")
	thumbnailPath, err := CreateVideoThumbnail(path, thumbnailPath)
	if err != nil {
		fmt.Printf("Failed to upload file: %v", err)
		return
	}
	RethumbnailPath := strings.ReplaceAll(thumbnailPath, "Thumbnail.jpg", "ReThumbnail.jpg")
	imgRe := ResizeImage(thumbnailPath, RethumbnailPath, 30, 30, 90)
	os.Remove(thumbnailPath)
	imgData, _ := ioutil.ReadFile(imgRe)
	respX, _ := cl.Client.Upload(context.Background(), imgData, whatsmeow.MediaImage)
	msg := &waProto.Message{VideoMessage: &waProto.VideoMessage{
		Caption:             proto.String(caption),
		Url:                 proto.String(uploaded.URL),
		DirectPath:          proto.String(uploaded.DirectPath),
		MediaKey:            uploaded.MediaKey,
		Mimetype:            proto.String(http.DetectContentType(data)),
		FileEncSha256:       uploaded.FileEncSHA256,
		FileSha256:          uploaded.FileSHA256,
		FileLength:          proto.Uint64(uint64(len(data))),
		JpegThumbnail:       imgData,
		ThumbnailDirectPath: &respX.DirectPath,
		ThumbnailSha256:     respX.FileSHA256,
		ThumbnailEncSha256:  respX.FileEncSHA256,
	}}
	cl.Client.SendMessage(context.Background(), jid, msg)
	os.Remove(RethumbnailPath)
}

func (cl *CLIENT) SendMention(jid types.JID, text string, mentionList []string) {
	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: proto.String(text),
			ContextInfo: &waProto.ContextInfo{
				MentionedJid: mentionList,
			},
		},
	}
	cl.Client.SendMessage(context.Background(), jid, msg)
}

func (cl *CLIENT) GetGroup(jid types.JID) *types.GroupInfo {
	data, _ := cl.Client.GetGroupInfo(jid)
	return data
}

func (cl *CLIENT) GetMemberList(jid types.JID) []string {
	memberList := []string{}
	data := cl.GetGroup(jid)
	for _, x := range data.Participants {
		cnv := fmt.Sprintf("%v", x)
		cnv = strings.ReplaceAll(cnv, "{", "")
		cnv = strings.ReplaceAll(cnv, "}", "")
		listCnv := strings.Split(cnv, " ")
		memberList = append(memberList, listCnv[0])
	}
	return memberList
}
