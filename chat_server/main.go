package main

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/log/log15adapter"
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'
	log "gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"os"
	"strconv"

	"github.com/go-park-mail-ru/2018_2_42/chat_server/acessor"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/hub"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/types"
	"github.com/go-park-mail-ru/2018_2_42/chat_server/websocket_upgrader"
)

func main() {
	port := flag.Uint16("port", 8080, "listen port for websocket server")
	flag.Parse()

	// соединение к базе
	// Инициализируем логгер, результаты выводятся в stdout.
	logger := log15adapter.NewLogger(log.New("module", "pgx"))

	// Вытаскиваем статический объект пакета,
	// агрегатор SQL выражений, которые надо подготовить в базе данных.
	prep := &accessor.Prep

	// Устанавливаем соединение с базой данных.
	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "postgres",
			Password: "",
			Database: "postgres",
			Logger:   logger,
		},
		MaxConnections: 10, // как вокеров
		// Создаём таблицы в базе данных,
		// Компилируем sql запросы для каждого соединения после их установления.
		AfterConnect: prep.Execute,
	})
	if err != nil {
		log.Crit("Unable to create connection pool", "error", err)
		os.Exit(1)
	}
	ConnPool := &accessor.ConnPool{ConnPool: *pool}
	defer ConnPool.Close()

	userHub := hub.Hub{
		SendNewMessage: make(chan types.Message, 1000),
		SendHistory:    make(chan types.HistoryRequest, 1000),
		NewUser:        make(chan *hub.User, 100),
		ConnPool:       ConnPool,
	}

	for i := 0; i < 10; i++ {
		go userHub.HubWorker()
	}

	// Инициализируем upgrader - он превращает соединения в websocket.
	upgrader := websocket_upgrader.NewConnectionUpgrader(&userHub)
	http.HandleFunc("/chat/v1/", upgrader.HttpEntryPoint)
	portStr := strconv.Itoa(int(*port))
	log.Info("Listening on :" + portStr)
	log.Error(http.ListenAndServe(":"+portStr, nil).Error())
	return
}
