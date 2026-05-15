package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/juggleim/imbot-sdk-go/imbotclients"
	"github.com/juggleim/imbot-sdk-go/imbotclients/pbdefines/pbobjs"
)

const (
// address        = "ws://127.0.0.1:9002"
// appkey         = "appkey"
// Token1  string = "CgZhcHBrZXkaICAvo1UH53CiwR/aurQXCDBpogz9OGlWbbWDpDsMJ4dn"
// Token2  string = "CgZhcHBrZXkaIHyiKFX87ojypRsjRqk/IPYTkqTNEiuvvABITR/imPaH"
// Token3  string = "CgZhcHBrZXkaIDIBXriAVM4RyD7VLFv8vrR1+efi6LycPMuKqbQ/oVdF"
// Token4  string = "CgZhcHBrZXkaIKTH3MaZdkgLYMLsYpVmt/UT3jQkd2UgGX35LjN26ouz"
// Token5  string = "CgZhcHBrZXkaINYLoPeDJyh0HuZdk3Vx+dNs5RBD2/McgZiDjjyXS2Pm"
)

func main() {
	address := "wss://ws.juggle.im"
	appkey := "nwm6fxqt2aeebhb7"
	token := "ChBud202ZnhxdDJhZWViaGI3GhCK-7Ic0Van0kxFd3Q9tAyF"
	client := imbotclients.NewImBotClient(address, appkey)
	client.OnMessageCallBack = func(msg *pbobjs.DownMsg) {
		fmt.Printf("received message: %+v\n", msg)
	}
	code, ack := client.Connect(token)
	fmt.Println("connect code:", code)
	if ack != nil {
		fmt.Printf("connect ack: code=%d user_id=%s session=%s ext=%s\n", ack.Code, ack.UserId, ack.Session, ack.Ext)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("message listener started, press Ctrl+C to exit")
	<-sigCh

	client.Disconnect()
}
