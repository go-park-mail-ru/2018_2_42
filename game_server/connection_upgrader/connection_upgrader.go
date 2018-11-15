package connection_upgrader

import (
	"encoding/json"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/types"
	"github.com/go-park-mail-ru/2018_2_42/game_server/user_connection"
	"github.com/gorilla/websocket"
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
func (cu *ConnectionUpgrader) HttpEntryPoint(w http.ResponseWriter, r *http.Request) {
	// Проверяет login/password из get params.
	params := r.URL.Query()
	var login string
	{
		l, ok := params["login"]
		if ok && len(l) == 1 {
			login = l[0]
		} else {
			response, _ := json.Marshal(types.ServerResponse{
				Status:  "forbidden",
				Message: "missing_login",
			})
			w.Write(response)
			w.WriteHeader(http.StatusForbidden)
			r.Body.Close()
			return
		}
	}
	var token string
	{
		l, ok := params["token"]
		if ok && len(l) == 1 {
			token = l[0]
		} else {
			response, _ := json.Marshal(types.ServerResponse{
				Status:  "forbidden",
				Message: "missing_session_cookie",
			})
			w.Write(response)
			w.WriteHeader(http.StatusForbidden)
			r.Body.Close()
			return
		}
	}
	// проверяет наличие этого пользователя в базе
	// TODO: реализовать проверку пользователя. Для этого нужен handler в authorisation server

	// Меняет протокол.
	WSConnection, err := cu.upgrader.Upgrade(w, r, nil)
	if err != nil {
		response, _ := json.Marshal(types.ServerResponse{
			Status:  "bad request",
			Message: "error on upgrade connection: " + err.Error(),
		})
		w.Write(response)
		w.WriteHeader(http.StatusBadRequest)
		r.Body.Close()
		return
	}

	connection := &user_connection.UserConnection{
		Login:      login,
		Token:      token,
		Connection: WSConnection,
	}

	cu.QueueToGame <- connection
}
