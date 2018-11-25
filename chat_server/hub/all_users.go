package hub

import (
	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type User struct {
	Connection *websocket.Conn
	Login      string
	ToUser     chan types.Messages
	Hub        *Hub // для hub.SendHistory hub.SendNewMessage
}

type AllUsers struct {
	sync.Map
}

// демоны пользователя.
// слушает из сокета и парсит задачи
func (u *User) ListeningDemon() {
	log.Print("start of '" + u.Login + "' ListeningDemon")
	for {
		_, message, err := u.Connection.ReadMessage()
		if err != nil {
			break
		}

		var event types.Event
		err = event.UnmarshalJSON(message)
		if err != nil {
			log.Print(err)
			continue
		}
		if event.Method == "send" {
			messages := types.Messages{}
			err = messages.UnmarshalJSON(event.Parameter)
			if err != nil {
				log.Print(err)
				continue
			}
			for _, message := range messages {
				message.From = &u.Login
				u.Hub.SendNewMessage <- message
			}
			continue
		}
		if event.Method == "history" {
			historyRequest := types.HistoryRequest{}
			err = historyRequest.UnmarshalJSON(event.Parameter)
			if err != nil {
				log.Print(err)
				continue
			}
			historyRequest.To = &u.Login
			u.Hub.SendHistory <- historyRequest
			continue
		}
	}
	// TODO: каскадное удаление.
	log.Print("end of '" + u.Login + "' ListeningDemon")
}

// слушает из канала и маршалит задачи
func (u *User) WritingDemon() {
	log.Print("start of '" + u.Login + "' WritingDemon")
	for masages := range u.ToUser {
		response, err := masages.MarshalJSON()
		if err != nil {
			log.Print(err)
			continue
		}
		err = u.Connection.WriteMessage(websocket.TextMessage, response)
		if err != nil {
			log.Print(err)
			continue
		}
	}
	// TODO: каскадное удаление.
	log.Print("end of '" + u.Login + "' WritingDemon")
}
