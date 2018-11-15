package main

import (
	"log"
	"net/http"

	"github.com/go-park-mail-ru/2018_2_42/game_server/connection_upgrader"
	"github.com/go-park-mail-ru/2018_2_42/game_server/rooms_manager"
)

func main() {
	upgrader := connection_upgrader.NewConnectionUpgrader()
	roomsManager := rooms_manager.NewRoomsManager()
	go roomsManager.MaintainConnections(upgrader.QueueToGame)
	http.HandleFunc("/game/v1/entrypoint", upgrader.HttpEntryPoint)
	log.Println("Listening on :8081")
	log.Print(http.ListenAndServe(":8081", nil))
}
