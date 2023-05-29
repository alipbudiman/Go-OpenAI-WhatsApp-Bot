package helper

import (
	"fmt"
	"reflect"
	"strings"

	"go.mau.fi/whatsmeow/types"
)

func TrackEvents(evt interface{}) {
	fmt.Println("\033[34m\n", "[(>)", evt, "]", reflect.TypeOf(evt), "\033[0m")
}

func WriteDisplayMenu() string {
	h := "┎───「 MENU 」"
	h += "\n⊶ help"
	h += "\n⊶ ping"
	h += "\n⊶ help"
	h += "\n⊶ send image"
	h += "\n⊶ send video"
	return h
}

func SenderJIDConvert(jid types.JID) (types.JID, bool) {
	j := fmt.Sprintf("%v", jid)
	x := strings.Split(j, "@")
	y := strings.Split(x[0], ".")
	z := y[0] + "@" + x[1]
	jid, ok := parseJID(z)
	if !ok {
		return jid, false
	}
	return jid, true
}

func parseJID(arg string) (types.JID, bool) {
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
