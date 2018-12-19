package hub

import (
	"log"
	"time"

	"github.com/go-park-mail-ru/2018_2_42/chat_server/acessor"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
)

type Hub struct {
	// очередь на обработку сообщений - сохранение и рассылку.
	// Через неё проходят вообще все сообщения.
	SendNewMessage chan types.Message
	// очередь на запрос истории.
	// Все запросы истории проходят через неё.
	SendHistory chan types.HistoryRequest
	// для новых пользователей.
	NewUser chan *User
	// map со всеми пользователями.  AllUsers[login].(types.User).ToUser <- "все сообщения для него"
	AllUsers AllUsers
	// соединение с базой
	ConnPool *accessor.ConnPool
}

func logString(s *string) string {
	if s != nil {
		return "'" + *s + "'"
	} else {
		return "nil"
	}
}

// вся логика
func (h *Hub) HubWorker() {
	for {
		select {
		case historyRequest := <-h.SendHistory:
			log.Printf("HubWorker historyRequest: From=%s To=%s Before=%d ", logString(historyRequest.From), logString(historyRequest.To), historyRequest.Before)
			messages, err := h.ConnPool.MassagesSelect(historyRequest.To, historyRequest.From, historyRequest.Before)
			if err != nil {
				log.Print("HubWorker historyRequest database err:" + err.Error())
				continue
			}
			user, ok := h.AllUsers.Load(*historyRequest.To)
			if !ok {
				log.Printf("HubWorker historyRequest no user: From %s, To %s. Before %d in %#v", logString(historyRequest.From), logString(historyRequest.To), historyRequest.Before, h.AllUsers)
				continue
			}
			user.ToUser <- messages
		case newMessage := <-h.SendNewMessage:
			log.Printf("HubWorker newMessage: %#v", newMessage)
			now := time.Now()
			id, err := h.ConnPool.MassagesInsert(newMessage.To, newMessage.From, newMessage.Text, now, newMessage.Reply.Id)
			if err != nil {
				log.Print("HubWorker newMessage database err:" + err.Error())
				continue
			}
			newMessage.Id = id
			newMessage.Time = now.Format(time.RFC3339)
			if newMessage.To != nil {
				user, ok := h.AllUsers.Load(*newMessage.To)
				if !ok {
					log.Printf("HubWorker newMessage: User not exist: %#v", newMessage)
					continue
				}
				user.ToUser <- types.Messages{newMessage}
			} else {
				h.AllUsers.Range(func(key string, value *User) bool {
					value.ToUser <- types.Messages{newMessage}
					return true
				})
			}
		case newUser := <-h.NewUser:
			log.Printf("HubWorker newUser: %#v", newUser)
			h.AllUsers.LoadOrStore(newUser.Login, newUser)
			// TODO: проверка перезатирания пользователя
			go newUser.ListeningDemon()
			go newUser.WritingDemon()
		}
	}
	return
}
