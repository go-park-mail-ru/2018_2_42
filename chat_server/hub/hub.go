package hub

import (
	"github.com/go-park-mail-ru/2018_2_42/chat_server/all_users"
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
	NewUser chan *all_users.User
	// map со всеми пользователями.  AllUsers[login].(types.User).ToUser <- "все сообщения для него"
	AllUsers all_users.AllUsers
}
