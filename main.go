package main

import (
	"fmt"

	"github.com/juggleim/imbot-sdk-go/imbotclients"
)

const (
	address        = "ws://127.0.0.1:9002"
	appkey         = "appkey"
	Token1  string = "CgZhcHBrZXkaICAvo1UH53CiwR/aurQXCDBpogz9OGlWbbWDpDsMJ4dn"
	Token2  string = "CgZhcHBrZXkaIHyiKFX87ojypRsjRqk/IPYTkqTNEiuvvABITR/imPaH"
	Token3  string = "CgZhcHBrZXkaIDIBXriAVM4RyD7VLFv8vrR1+efi6LycPMuKqbQ/oVdF"
	Token4  string = "CgZhcHBrZXkaIKTH3MaZdkgLYMLsYpVmt/UT3jQkd2UgGX35LjN26ouz"
	Token5  string = "CgZhcHBrZXkaINYLoPeDJyh0HuZdk3Vx+dNs5RBD2/McgZiDjjyXS2Pm"
)

func main() {
	client := imbotclients.NewImBotClient(address, appkey)
	code, ack := client.Connect(Token1)
	fmt.Println("connect code:", code)
	if ack != nil {
		fmt.Printf("connect ack: code=%d user_id=%s session=%s ext=%s\n", ack.Code, ack.UserId, ack.Session, ack.Ext)
	}
	client.Disconnect()
}
