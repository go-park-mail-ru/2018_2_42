package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2018_2_42/game_server/connection_upgrader"
	"github.com/go-park-mail-ru/2018_2_42/game_server/rooms_manager"
	"github.com/go-park-mail-ru/2018_2_42/game_server/websocket_testing_page"
)

var port = "8081"

func main() {
	// Инициализируем upgrader - он превращает соединения в websocket.
	upgrader := connection_upgrader.NewConnectionUpgrader()
	roomsManager := rooms_manager.NewRoomsManager()
	go roomsManager.MaintainConnections(upgrader.QueueToGame)
	http.HandleFunc("/game/v1/entrypoint", upgrader.HttpEntryPoint)
	http.HandleFunc("/", websocket_testing_page.WebSocketTestPage)
	log.Println("Listening on :" + port)
	log.Print(http.ListenAndServe(":"+port, nil))
}
