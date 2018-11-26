package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"

	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
)

type User struct {
	Connection *websocket.Conn
	Login      string
	ToUser     chan types.Messages
	Hub        *Hub // для hub.SendHistory hub.SendNewMessage
}

type AllUsers struct {
	syncMap sync.Map
}

// функции для починки типизации.
func (au *AllUsers) Delete(key string) {
	au.syncMap.Delete(key)
	return
}

func (au *AllUsers) Load(key string) (value *User, ok bool) {
	interfacedValue, ok := au.syncMap.Load(key)
	value = interfacedValue.(*User)
	return
}

func (au *AllUsers) LoadOrStore(key string, value *User) (actual *User, loaded bool) {
	interfacedActual, loaded := au.syncMap.LoadOrStore(key, value)
	actual = interfacedActual.(*User)
	return
}

func (au *AllUsers) Range(f func(key string, value *User) bool) {
	au.syncMap.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*User))
	})
	return
}

func (au *AllUsers) Store(key string, value *User) {
	au.syncMap.Store(key, value)
	return
}

// демоны пользователя.
// слушает из сокета и парсит задачи
func (u *User) ListeningDemon() {
	log.Print("start of '" + u.Login + "' ListeningDemon")
	for {
		_, message, err := u.Connection.ReadMessage()
		if err != nil {
			log.Printf("ListeningDemon ReadMessage error: %#v", err.Error())
			break
		}

		var event types.Event
		err = event.UnmarshalJSON(message)
		if err != nil {
			log.Printf("ListeningDemon UnmarshalJSON %#v %#v: ", string(message), err.Error())
			continue
		}
		if event.Method == "send" {
			messages := types.Messages{}
			err = messages.UnmarshalJSON(event.Parameter)
			if err != nil {
				log.Printf("ListeningDemon UnmarshalJSON %#v %#v: ", string(event.Parameter), err.Error())
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
				log.Printf("ListeningDemon UnmarshalJSON %#v %#v: ", string(event.Parameter), err.Error())
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
	for messages := range u.ToUser {
		log.Printf("WritingDemon of '%s': messages=%#v", u.Login, messages)
		response, err := messages.MarshalJSON()
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
