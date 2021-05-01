package main

// "wss://chat.wasd.tv/socket.io/?EIO=3&transport=websocket"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lerner17/hjkl/pkg/logger"
)

const (
	scheme = "wss"
	host   = "chat.wasd.tv"
)

// 42["message",{
// 	"id":"aca38137-aa64-4d56-bd3b-2cd288874979",
// 	"user_id":84722,
// 	"message":"на 9 мая готовься заранее, борщик по жирнее и погуще",
// 	"user_login":"Bubelka",
// 	"user_avatar":{
// 		"large":"https://static.wasd.tv/avatars/user/1697.png",
// 		"small":"https://static.wasd.tv/avatars/user/1697.png",
// 		"medium":"https://static.wasd.tv/avatars/user/1697.png"
// 	},
// 	"hash":"cko606lhn000y3a9xv9xbx6x9",
// 	"is_follower":false,
// 	"other_roles":["CHANNEL_SUBSCRIBER"],
// 	"user_channel_role":"CHANNEL_USER",
// 	"channel_id":96081,
// 	"stream_id":738795,
// 	"date_time":"2021-05-01T17:12:43.505Z"
// }]

type UserAvatar struct {
	Large  string `json:"large"`
	Small  string `json:"small"`
	Medium string `json:"medium"`
}
type message struct {
	ID         string     `json:"id"`
	UserID     int        `json:"user_id"`
	Message    string     `json:"message"`
	Avatar     UserAvatar `json:"user_avatar"`
	Hash       string     `json:"hash"`
	IsFollower bool       `json:"is_follower"`
}

// 42["giftsV1",{
// 	"id":"4041470c-b52f-4ff8-85b7-416e1961ab76",
// 	"gift_code":"league7_gift_4",
// 	"price_id":38,
// 	"stream_id":738795,
// 	"channel_id":96081,
// 	"customer_id":899024,
// 	"gift_name":"league7_gift_4",
// 	"gift_description":" ",
// 	"amount":1,
// 	"send_at":"2021-05-01T17:12:44Z",
// 	"gift_graid":1,
// 	"gift_animation":"https://static.wasd.tv/gifts/league_7/league7_gift_4/league_gift_4.gif",
// 	"gift_animation_retina":"https://static.wasd.tv/gifts/league_7/league7_gift_4/league_gift_4_retina.gif"
// }]

type giftsV1 struct {
	ID                  string
	GiftCode            string
	PriceID             int
	StreamID            int
	ChannelID           int
	CustomerID          int
	GiftName            string
	GiftDescription     string
	Amount              int
	GiftGraid           int
	GiftAnimation       string
	GiftAnimationRetina string
}

func parseMessage(msg []byte) {
	switch {
	case bytes.HasPrefix(msg, []byte(`42["message",`)):
		var jsonText = msg[13 : len(msg)-1]
		var payload message

		if err := json.Unmarshal(jsonText, &payload); err != nil {
			fmt.Println("Err Parse: ", err)
		}
		fmt.Printf("%#v\n---\n", payload)
		return
	case bytes.HasPrefix(msg, []byte(`42["giftsV1",`)):
		var jsonText = msg[12 : len(msg)-1]
		var payload giftsV1
		if err := json.Unmarshal(jsonText, &payload); err != nil {
			fmt.Printf("Err Parse Gift: %s, %v\n---\n", jsonText, err)
		}
		fmt.Printf("%#v\n---\n", payload)
	}
}

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logger, err := logger.New("logs/logs.log")
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	u := url.URL{Scheme: scheme, Host: host, Path: "/socket.io/", RawQuery: "EIO=3&transport=websocket"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(fmt.Errorf("dial: %v", err))
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				panic(fmt.Errorf("read: %v", err))
			}
			parseMessage(message)
			logger.Info(string(message))
		}
	}()
	var dobriyactionJWT = fmt.Sprintf(`42["join",{"streamId":%s,"channelId":%s,"jwt":"%s","excludeStickers":true}]`, "738795", "96081", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX3JvbGUiOiJHVUVTVCIsImlhdCI6MTYxOTg2MjYyMCwiZXhwIjoxNjE5ODkxNDIwfQ.GDFgD_v2UEptNjZOa1jMEPNM4L99x2Nl8d2GqIzpfQE")
	if err := c.WriteMessage(websocket.TextMessage, []byte(dobriyactionJWT)); err != nil {
		panic(fmt.Errorf("write jwt: %v", err))
	}

	ticker := time.NewTicker(time.Second * 20)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte("2"))
			if err != nil {
				panic(fmt.Errorf("could not write: %v", err))
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				panic(fmt.Errorf("write close: %v", err))
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
