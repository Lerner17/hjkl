package main

// "wss://chat.wasd.tv/socket.io/?EIO=3&transport=websocket"

// var addr = flag.String("addr", ":8080", "http service address")

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

// var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: "chat.wasd.tv", Path: "/socket.io/", RawQuery: "EIO=3&transport=websocket"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	const jwt = `42["join",{"streamId":738781,"channelId":60117,"jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX3JvbGUiOiJHVUVTVCIsImlhdCI6MTYxOTg2MjU4NSwiZXhwIjoxNjE5ODkxMzg1fQ.I-x3_xf0yZem8TGxV8XDaCmyQR97sokq3EZ49E6WfYA","excludeStickers":true}]`
	if err := c.WriteMessage(websocket.TextMessage, []byte(jwt)); err != nil {
		log.Println("write jwd:", err)
		return
	}

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case _ = <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("2"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// 0{"sid":"S7UykRZ0gNzRWVqmCcku","upgrades":[],"pingInterval":25000,"pingTimeout":5000}
// 40
