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
	address = "ws://127.0.0.1:9002"
	appkey  = "appkey"
)

func main() {
	token := "CgZhcHBrZXkaIGvZMSW_6pEgEc_TCzR72mG6hWjgpR1WivvfEBH5jyOd"
	client := imbotclients.NewImBotClient(address, appkey, imbotclients.TransportMode_WebSocket)
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
