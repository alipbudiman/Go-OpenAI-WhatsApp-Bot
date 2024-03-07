package main

import (
	"context"
	"fmt"
	"mime"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"github.com/sirupsen/logrus"

	"gowagpt/metadevlibs/botlib"
	"gowagpt/metadevlibs/feature"
	"gowagpt/metadevlibs/helper"
	"gowagpt/metadevlibs/object"
	"gowagpt/metadevlibs/transport"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientWrapper struct {
	*botlib.CLIENT
}

var (
	Log            *logrus.Logger
	Client         *ClientWrapper
	myJID          types.JID
	ChatGPTApikey  string = "" // << your apikey here
	ChatGPTProxy   string = ""
	checkRead             = make(map[string]int)
	readerTemp            = make(map[string][]string)
	antiUnsend            = make(map[types.JID][]string)
	antiUnsend_img        = make(map[string]*waProto.ImageMessage)
	antiUnsend_vid        = make(map[string]*waProto.VideoMessage)
	antiUnsend_aud        = make(map[string]*waProto.AudioMessage)
	UnsendRead            = make(map[string]int)
	StkConv               = make(map[string]int)
)

func (cl *ClientWrapper) MessageHandler(evt interface{}) {
	helper.TrackEvents(evt)
	switch v := evt.(type) {
	case *events.Message:
		text_in_ExtendedText := v.Message.ExtendedTextMessage.GetText()
		mobile_txt := v.Message.GetConversation()
		pc_txt := v.Message.ExtendedTextMessage.GetText()
		var txtV2 string
		var txt string
		var from_dm bool = false
		switch {
		case text_in_ExtendedText != "":
			txt = strings.ToLower(text_in_ExtendedText)
			txtV2 = text_in_ExtendedText
		case pc_txt != "":
			txt = strings.ToLower(pc_txt)
			txtV2 = pc_txt
		default:
			txt = strings.ToLower(mobile_txt)
			txtV2 = mobile_txt
		}
		sender := v.Info.Sender
		senderSTR := fmt.Sprintf("%v", sender)
		sender_jid, is_success := helper.SenderJIDConvert(sender)
		if is_success {
			sender = sender_jid
		}
		to := v.Info.Chat
		if to == sender {
			from_dm = true
		}

		go func() {
			if StkConv[fmt.Sprintf("%v", to)] == 1 {
				img := v.Message.GetImageMessage()
				if img != nil {
					StkConv[fmt.Sprintf("%v", to)] = 0
					data, err := cl.Client.Download(img)
					if err != nil {
						cl.SendTextMessage(to, err.Error())
						return
					}
					exts, _ := mime.ExtensionsByType(img.GetMimetype())
					path := fmt.Sprintf("%s%s", v.Info.ID, exts[0])
					err = os.WriteFile(path, data, 0600)
					if err != nil {
						cl.SendTextMessage(to, err.Error())
						return
					}
					fmt.Println("Saved image in message to", path)
					webp_success, webp_path := helper.ConvertJPEtoWEBP(path)
					if webp_success {
						cl.SendMention(to, "👾 success create sticker "+helper.MentionFormat(fmt.Sprintf("%v", sender)), []string{fmt.Sprintf("%v", sender)})
						cl.SendStickerMessage(to, webp_path, false)
					} else {
						cl.SendMention(to, "👾 fail create sticker "+helper.MentionFormat(fmt.Sprintf("%v", sender)), []string{fmt.Sprintf("%v", sender)})
					}
					os.Remove(path)
					os.Remove(webp_path)
				}
			}

		}()
		go cl.TrackUnsendMessage(to, v, txtV2, sender.String())
		if txt == "ping" {
			cl.SendTextMessage(to, "Pong!")
		} else if txt == "help" {
			cl.SendTextMessage(to, helper.WriteDisplayMenu(from_dm))
		} else if txt == "send image" {
			cl.SendTextMessage(to, "Loading . . .")
			cl.SendImageMessage(to, "assets/img/img.jpg", ">_<")
		} else if txt == "send video" {
			cl.SendTextMessage(to, "Loading . . .")
			cl.SendVideoMessage(to, "assets/vid/vid.mp4", ">_<")
		} else if strings.HasPrefix(txt, "say: ") {
			spl := txtV2[len("say: "):]
			cl.SendTextMessage(to, spl)
		} else if strings.HasPrefix(txt, "chat gpt: ") {
			cl.SendTextMessage(to, "Process . . .")
			question := txtV2[len("chat gpt: "):]
			responGPT, err := feature.ChatGPT(sender, question, "")
			if err != nil {
				cl.SendTextMessage(to, "Error please check console for detail")
				panic(err)
			}
			buildRespon := "*Chat GPT Response:*"
			buildRespon += "\n" + responGPT
			buildRespon += "\n\n____ [done] ____"
			for _, msg := range helper.LooperMessage(buildRespon, 2000) {
				cl.SendTextMessage(to, msg)
				time.Sleep(1 * time.Second)
			}
			buildConvertation := "*Convertation*"
			buildConvertation += "\nConvertation by: " + helper.MentionFormat(senderSTR)
			buildConvertation += fmt.Sprintf("\nTotal convertation: %d", (len(feature.GPTMap[sender])-1)/2)
			cl.SendMention(to, buildConvertation, []string{senderSTR})

		} else if strings.HasPrefix(txt, "dalle draw: ") {
			cl.SendTextMessage(to, "Process . . .")
			prompt := txtV2[len("dalle draw: "):]
			imgLoad := 3           // Total of DAll-E load image (max 10)
			imgSize := "1024x1024" // Generated images can have a size of "256x256", "512x512", or "1024x1024 "pixels
			responDallE, err := feature.DallE(prompt, imgLoad, imgSize)
			if err != nil {
				cl.SendTextMessage(to, "Error please check console for detail")
				panic(err)
			}
			for i, img := range responDallE {
				fileName := object.GenerateFileName(".jpg")
				data, err := transport.Download(img, "tmp/img", fileName)
				if err != nil {
					cl.SendTextMessage(to, fmt.Sprintf("Error Fail load image %d, please check console for detail", i+1))
					fmt.Println(err.Error())
					continue
				}
				msg := fmt.Sprintf("Total Generate image: %d / %d", imgLoad, i+1)
				msg += fmt.Sprintf("\nTotal word in promp: %d", len(strings.Fields(prompt)))
				cl.SendImageMessage(to, data, msg)
				os.Remove(data)
			}
		} else if strings.HasPrefix(txt, "make sticker") {
			group_id := fmt.Sprintf("%v", to)
			StkConv[group_id] = 1
			cl.SendTextMessage(to, "👾 [Sticker] Please send image")
		} else if strings.HasPrefix(txt, "group broadcast: ") {
			cl.SendTextMessage(to, "Process . . .")
			message := txtV2[len("group broadcast: "):]
			cl.SendMessageToAllGroup(to, message)
			hello_wolrd := []string{
				"Hello, world",
				"Programmed to work and not to feel",
				"Not even sure that this is real",
				"Hello, world",
			}
			res, err := cl.SendTextMessage(to, "Start . . .")
			if err != nil {
				cl.SendTextMessage(to, err.Error())
				return
			}
			for _, hw := range hello_wolrd {
				time.Sleep(time.Second * 1)
				res, err = cl.EditMessage(to, res.ID, hw)
				if err != nil {
					cl.SendTextMessage(to, err.Error())
					break
				}
			}
		}

		if !from_dm {
			if txt == "tag all" {
				cl.SendTextMessage(to, "Loading . . .")
				mem := cl.GetMemberList(to)
				mem = helper.RemoveMyJID(mem, myJID)
				ret := "⌬ Mentionall\n"
				for _, jid := range mem {
					ret += "\n- " + helper.MentionFormat(jid)
				}
				ret += fmt.Sprintf("\n\nTotal %v user", len(mem))
				cl.SendMention(to, ret, mem)
			} else if strings.HasPrefix(txt, "say: ") {
				msg := txtV2[len("say: "):]
				cl.SendTextMessage(to, msg)
			} else if strings.HasPrefix(txt, "reader ") {
				spl := strings.Replace(txt, "reader ", "", 1)
				if spl == "on" {
					readerTemp[to.String()] = []string{}
					checkRead[to.String()] = 1
					cl.SendTextMessage(to, "👾Reader enabled")
				} else if spl == "off" {
					readerTemp[to.String()] = []string{}
					checkRead[to.String()] = 0
					cl.SendTextMessage(to, "👾Reader disable")
				} else {
					cl.SendTextMessage(to, "👾 For Active Reader type `reader on`, Turn off type `reader off`")
				}
			} else if strings.HasPrefix(txt, "anti unsend ") {
				spl := strings.Replace(txt, "anti unsend ", "", 1)
				g := fmt.Sprintf("%v", to)
				if spl == "on" {
					UnsendRead[g] = 1
					cl.SendTextMessage(to, "👾 Anti unsend enabled")
				} else if spl == "off" {
					UnsendRead[g] = 0
					cl.SendTextMessage(to, "👾 Anti unsend disable")
				} else {
					cl.SendTextMessage(to, "👾 For Active Anti Unsend, type `anti unsend on`, Turn off type `anti unsend off`")
				}
			}
		}
		return
	case *events.Receipt:
		go cl.ProcessReader(v)
	default:
		fmt.Println(reflect.TypeOf(v))
		fmt.Println(v)
	}
}

func (cl *ClientWrapper) ProcessReader(evt *events.Receipt) {
	param1 := fmt.Sprintf("%v", evt.Chat)
	if checkRead[param1] == 1 {
		if strings.Contains(evt.Type.GoString(), "events.ReceiptTypeRead") {
			go func() {
				param2 := fmt.Sprintf("%v", evt.Sender)
				if !helper.InArray(readerTemp[param1], param2) && evt.Sender.String() != myJID.String() {
					readerTemp[param1] = append(readerTemp[param1], param2)
					jid, param := helper.ParseJIDUser(param1)
					if param {
						XSiderMsg := "👾 Hello i see you " + helper.MentionFormat(param2)
						cl.SendMention(jid, XSiderMsg, []string{param2})
					}
				}
			}()

		}
	}
}

func (cli *ClientWrapper) TrackUnsendMessage(to types.JID, v *events.Message, text string, sender string) {
	if UnsendRead[fmt.Sprintf("%v", to)] == 1 {
		if strings.Contains(fmt.Sprintf("%v", v), "protocolMessage:{key:{remoteJid:") {
			if strings.Contains(fmt.Sprintf("%v", v), "type:REVOKE") {
				uns_id := v.Message.ProtocolMessage.Key.GetId()
				if uns_id != "" {
					for _, umsg := range antiUnsend[to] {
						Sumsg := strings.Split(umsg, "∬∬∬")
						if Sumsg[0] == uns_id {
							switch {
							case Sumsg[2] == "text":
								cli.SendMention(to, fmt.Sprintf("*User Delete Message*\nType: Text\n%s\nMessage:\n%s", helper.MentionFormat(sender), Sumsg[1]), []string{sender})
								antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
							case Sumsg[2] == "img":
								snd := fmt.Sprintf("%v", sender)
								img := antiUnsend_img[uns_id]
								data, err := cli.Client.Download(img)
								if err != nil {
									cli.SendMention(to, "Fail to restore message "+helper.MentionFormat(snd), []string{snd})
									antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
									delete(antiUnsend_img, uns_id)
									return
								}
								exts, _ := mime.ExtensionsByType(img.GetMimetype())
								path := fmt.Sprintf("%s%s", v.Info.ID, exts[0])
								defer os.Remove(path)
								err = os.WriteFile(path, data, 0600)
								Cvpath, errCv := helper.ConvertJPEtoJPG(path)
								defer os.Remove(Cvpath)
								if errCv != nil {
									cli.SendTextMessage(to, errCv.Error())
									return
								}
								if err != nil {
									cli.SendMention(to, "Fail to restore message "+helper.MentionFormat(snd), []string{snd})
									antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
									delete(antiUnsend_img, uns_id)
									return
								}
								cm := fmt.Sprintf("*User Delete Message*\nType: Image\n%s\nMessage:\n%s", helper.MentionFormat(sender), Sumsg[1])
								cli.SendImageMessage(to, Cvpath, cm)
								antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
								delete(antiUnsend_img, uns_id)
							case Sumsg[2] == "vid":
								snd := fmt.Sprintf("%v", sender)
								vid := antiUnsend_vid[uns_id]
								data, err := cli.Client.Download(vid)
								if err != nil {
									cli.SendMention(to, "Fail to restore message "+helper.MentionFormat(snd), []string{snd})
									antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
									delete(antiUnsend_vid, uns_id)
									return
								}
								exts, _ := mime.ExtensionsByType(vid.GetMimetype())
								path := fmt.Sprintf("%s%s", v.Info.ID, exts[0])
								defer os.Remove(path)
								err = os.WriteFile(path, data, 0600)
								if err != nil {
									cli.SendMention(to, "Fail to restore message "+helper.MentionFormat(snd), []string{snd})
									antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
									delete(antiUnsend_vid, uns_id)
								}
								AddName := helper.RandomStrings(7)
								newName := strings.ReplaceAll(path, ".f4v", AddName+".mp4")
								newPath, err := helper.ConvertF4VtoMP4(path, newName)
								defer os.Remove(newPath)
								if err != nil {
									cli.SendMention(to, "Fail to restore message "+helper.MentionFormat(snd), []string{snd})
									antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
									delete(antiUnsend_vid, uns_id)
								}
								cm := fmt.Sprintf("*User Delete Message*\nType: Video\n%s\nMessage:\n%s", helper.MentionFormat(sender), Sumsg[1])
								cli.SendVideoMessage(to, newPath, cm)
								antiUnsend[to] = helper.Remove(antiUnsend[to], umsg)
								delete(antiUnsend_vid, uns_id)
							}
							return
						}
					}
				}
			}

		}
		if len(antiUnsend[to]) >= 100 {
			antiUnsend[to] = helper.Remove(antiUnsend[to], (antiUnsend[to])[0])
			uns := strings.Split((antiUnsend[to])[0], "∬∬∬")[0]
			delete(antiUnsend_img, uns)
			delete(antiUnsend_vid, uns)
		}
		img := v.Message.GetImageMessage()
		vid := v.Message.GetVideoMessage()
		aud := v.Message.GetAudioMessage()
		if img != nil {
			img_caption := v.Message.ImageMessage.GetCaption()
			antiUnsend_img[v.Info.ID] = img
			antiUnsend[to] = append(antiUnsend[to], fmt.Sprintf("%s∬∬∬%s∬∬∬img", v.Info.ID, img_caption))
			return
		} else if vid != nil {
			vid_caption := v.Message.VideoMessage.GetCaption()
			antiUnsend_vid[v.Info.ID] = vid
			antiUnsend[to] = append(antiUnsend[to], fmt.Sprintf("%s∬∬∬%s∬∬∬vid", v.Info.ID, vid_caption))
			return
		} else if aud != nil {
			antiUnsend_aud[v.Info.ID] = aud
			antiUnsend[to] = append(antiUnsend[to], fmt.Sprintf("%s∬∬∬None∬∬∬aud", v.Info.ID))
			return
		}
		antiUnsend[to] = append(antiUnsend[to], fmt.Sprintf("%s∬∬∬%s∬∬∬text", v.Info.ID, text))
	}
}

func (cl *ClientWrapper) register() {
	cl.Client.AddEventHandler(cl.MessageHandler)
}

func (cl *ClientWrapper) newClient(d *store.Device, l waLog.Logger) {
	cl.Client = whatsmeow.NewClient(d, l)
}

func main() {
	feature.GPTConfig("", ChatGPTApikey, ChatGPTProxy)
	store.DeviceProps.RequireFullSync = proto.Bool(false)
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	if _, err := os.Stat("db/sql"); os.IsNotExist(err) {
		os.MkdirAll("db/sql", 0755)
	}
	container, err := sqlstore.New("sqlite3", "file:db/sql/commander.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	makeJID, _ := helper.ConvertJID(fmt.Sprintf("%v", deviceStore.ID))
	Resjid, _ := helper.SenderJIDConvert(makeJID)
	myJID = Resjid

	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	Client = &ClientWrapper{
		CLIENT: &botlib.CLIENT{},
	}
	Client.newClient(deviceStore, clientLog)
	Client.register()

	if Client.Client.Store.ID == nil {
		qrChan, _ := Client.Client.GetQRChannel(context.Background())
		err = Client.Client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}

	} else {
		err = Client.Client.Connect()
		fmt.Println("Login Success")
		if err != nil {
			panic(err)
		}
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	Client.Client.Disconnect()
}
