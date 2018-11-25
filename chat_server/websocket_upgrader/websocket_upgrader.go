package websocket_upgrader

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2018_2_42/chat_server/hub"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
)

// Ответственность: превращение http соединения в игрока - пользователя.
type ConnectionUpgrader struct {
	// Настройки WebSocket.
	upgrader websocket.Upgrader
	// hub.NewUser chan *all_users.User - канал, в который помещаются соединения с пользователем, что бы передать их в RoomManager
	Hub *hub.Hub
}

// Фабричная функция вместо конструктора.
func NewConnectionUpgrader(hub *hub.Hub) (cu *ConnectionUpgrader) {
	cu = &ConnectionUpgrader{
		upgrader: websocket.Upgrader{
			HandshakeTimeout: time.Duration(1 * time.Second),
			CheckOrigin: func(r *http.Request) bool { // Токен не проверяется.
				return true
			},
			EnableCompression: true,
		},
		Hub: hub,
	}
	return
}

// Handler, входная точка для http соединения.
// Запускается в разных горутинах, только читает из класса.
func (cu *ConnectionUpgrader) HttpEntryPoint(w http.ResponseWriter, r *http.Request) {
	log.Printf("New connection: %#v", r)
	// Проверяет sessionid из cookie.
	sessionId, err := r.Cookie("sessionid")
	if err != nil {
		print("sessionId = " + sessionId.Value + "but registration not implemented!")
	}
	// проверяет наличие этого пользователя в базе
	// TODO: реализовать проверку пользователя. Для этого нужен jrpc handler в authorisation server
	login := "Anon" + time.Now().Format(time.RFC3339)

	// Меняет протокол.
	WSConnection, err := cu.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error on upgrade connection"))
		r.Body.Close()
		return
	}
	connection := &hub.User{
		Connection: WSConnection,
		Login:      login,
		ToUser:     make(chan types.Messages),
		Hub:        cu.Hub,
	}
	log.Println("new user " + connection.Login + "created")
	cu.Hub.NewUser <- connection
	return
}
