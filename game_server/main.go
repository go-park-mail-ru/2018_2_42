package main

import (
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'
	"log"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2018_2_42/game_server/connection_upgrader"
	"github.com/go-park-mail-ru/2018_2_42/game_server/rooms_manager"
	"github.com/go-park-mail-ru/2018_2_42/game_server/websocket_test_page"
)

func main() {
	port := flag.Uint16("port", 8080, "listen port for websocket server")
	flag.Parse()

	// Инициализируем upgrader - он превращает соединения в websocket.
	upgrader := connection_upgrader.NewConnectionUpgrader()
	roomsManager := rooms_manager.NewRoomsManager()
	go roomsManager.MaintainConnections(upgrader.QueueToGame)
	http.HandleFunc("/game/v1/entrypoint", upgrader.HttpEntryPoint)
	http.HandleFunc("/", websocket_test_page.WebSocketTestPage)
	portStr := strconv.Itoa(int(*port))
	log.Println("Listening on :" + portStr)
	log.Print(http.ListenAndServe(":"+portStr, nil))
}
