package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"github.com/sirupsen/logrus"

	"./metadevlibs/botlib"
	"./metadevlibs/helper"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waLog "go.mau.fi/whatsmeow/util/log"
)

type ClientWrapper struct {
	*botlib.CLIENT
}

var (
	Log    *logrus.Logger
	Client *ClientWrapper
	myJID  types.JID
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
		sender_jid, is_success := helper.SenderJIDConvert(sender)
		if is_success {
			sender = sender_jid
		}
		to := v.Info.Chat
		if to == sender {
			from_dm = true
		}
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
			}
		}
		return
	default:
		fmt.Println(reflect.TypeOf(v))
		fmt.Println(v)
	}
}

func (cl *ClientWrapper) register() {
	cl.Client.AddEventHandler(cl.MessageHandler)
}

func (cl *ClientWrapper) newClient(d *store.Device, l waLog.Logger) {
	cl.Client = whatsmeow.NewClient(d, l)
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:commander.db?_foreign_keys=on", dbLog)
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
