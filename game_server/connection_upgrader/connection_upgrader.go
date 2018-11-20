package connection_upgrader

import (
	"encoding/json"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/types"
	"github.com/go-park-mail-ru/2018_2_42/game_server/user_connection"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// Ответственность: превращение http соединения в игрока - пользователя.
type ConnectionUpgrader struct {
	// Настройки WebSocket.
	upgrader websocket.Upgrader
	// Канал, в который помещаются соединения с пользователем, что бы передать их в RoomManager
	QueueToGame chan *user_connection.UserConnection
}

// Фабричная функция вместо конструктора.
func NewConnectionUpgrader() (cu *ConnectionUpgrader) {
	cu = &ConnectionUpgrader{
		upgrader: websocket.Upgrader{
			HandshakeTimeout: time.Duration(1 * time.Second),
			CheckOrigin: func(r *http.Request) bool { // Токен не проверяется.
				return true
			},
			EnableCompression: true,
		},
		QueueToGame: make(chan *user_connection.UserConnection, 50),
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
		response, _ := json.Marshal(types.ServerResponse{
			Status:  "forbidden",
			Message: "missing_sessionid_cookie",
		})
		w.WriteHeader(http.StatusForbidden)
		w.Write(response)
		r.Body.Close()
		return
	}
	// проверяет наличие этого пользователя в базе
	// TODO: реализовать проверку пользователя. Для этого нужен handler в authorisation server
	login := "Anon"

	// Меняет протокол.
	WSConnection, err := cu.upgrader.Upgrade(w, r, nil)
	if err != nil {
		response, _ := json.Marshal(types.ServerResponse{
			Status:  "bad request",
			Message: "error on upgrade connection: " + err.Error(),
		})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		r.Body.Close()
		return
	}
	connection := &user_connection.UserConnection{
		Login:      login,
		Token:      sessionId.Value,
		Connection: WSConnection,
	}
	cu.QueueToGame <- connection
	return
}
