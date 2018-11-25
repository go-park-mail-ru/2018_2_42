package chat_server

import (
	"github.com/go-park-mail-ru/2018_2_42/chat_server/all_users"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/hub"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/websocket_upgrader"
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Uint16("port", 8080, "listen port for websocket server")
	flag.Parse()

	userHub := hub.Hub{
		SendNewMessage: make(chan types.Message, 1000),
		SendHistory:    make(chan types.HistoryRequest, 1000),
		NewUser:        make(chan *all_users.User, 100),
	}

	// Инициализируем upgrader - он превращает соединения в websocket.
	upgrader := websocket_upgrader.NewConnectionUpgrader(&userHub)
	http.HandleFunc("/chat/v1/", upgrader.HttpEntryPoint)
	portStr := strconv.Itoa(int(*port))
	log.Println("Listening on :" + portStr)
	log.Print(http.ListenAndServe(":"+portStr, nil))
}
