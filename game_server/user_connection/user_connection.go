package user_connection

import "github.com/gorilla/websocket"

// Соединение пользователя, заведомо валидное, за производство отвечает connection_upgrader.
type UserConnection struct {
	Login      string
	Token      string
	Connection *websocket.Conn
}
