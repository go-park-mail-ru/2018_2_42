package connectionUpgrader

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/go-park-mail-ru/2018_2_42/game_server/types"
	"github.com/go-park-mail-ru/2018_2_42/game_server/user_connection"
)

// ConnectionUpgrader ответственен за превращение http соединения в игрока - пользователя.
type ConnectionUpgrader struct {
	// Настройки WebSocket.
	upgrader websocket.Upgrader
	// Канал, в который помещаются соединения с пользователем, что бы передать их в RoomManager
	QueueToGame chan *user_connection.UserConnection
}

// NewConnectionUpgrader - фабричная функция вместо конструктора.
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

var debug = true

// HTTPEntryPoint - входная точка для http соединения.
// Запускается в разных горутинах, только читает из класса.
// Проводит upgrade соединения и проверку cookie полззователя.
func (cu *ConnectionUpgrader) HTTPEntryPoint(w http.ResponseWriter, r *http.Request) {
	log.Printf("New connection: %#v", r)
	// Проверяет SessionId из cookie.
	sessionID, err := r.Cookie("SessionId")
	if err != nil {
		response, _ := types.ServerResponse{
			Status:  "forbidden",
			Message: "missing_sessionid_cookie",
		}.MarshalJSON()
		w.WriteHeader(http.StatusForbidden)
		w.Write(response)
		r.Body.Close()
		return
	}

	var avatar string
	var login string
	if debug {
		// просто создаёт случайный логин
		login = "Anon" + time.Now().Format(time.RFC3339)
		avatar = "/images/default.png"
	}
	// Меняет протокол.
	WSConnection, err := cu.upgrader.Upgrade(w, r, nil)
	if err != nil {
		response, _ := types.ServerResponse{
			Status:  "bad request",
			Message: "error on upgrade connection: " + err.Error(),
		}.MarshalJSON()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(response)
		r.Body.Close()
		return
	}
	connection := &user_connection.UserConnection{
		Login:      login,
		Avatar:     avatar,
		Token:      sessionID.Value,
		Connection: WSConnection,
	}
	cu.QueueToGame <- connection
	return
}
