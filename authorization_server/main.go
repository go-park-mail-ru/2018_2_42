package main

import (
	"fmt"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/accessor"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/config"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/environment"
	gRPCServer "github.com/go-park-mail-ru/2018_2_42/authorization_server/grpc_authorisation_server"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/handlers"
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'
	"log"
	"net/http"
)

func main() {
	configPath := flag.String("config", "./main.json5", "path of config")
	flag.Parse()

	var err error
	basicEnv := environment.Environment{}
	handlersEnv := handlers.Environment(basicEnv)
	handlersEnv.Config, err = config.ParseConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	handlersEnv.DB, err = accessor.ConnectToDatabase(handlersEnv.Config)
	err = handlersEnv.DB.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	//start grpc_authorisation worker
	serverEnvironment := gRPCServer.ServerEnvironment(basicEnv)
	go func() {
		log.Print("gRPCServer.Worker number start")
		err := gRPCServer.Worker(&serverEnvironment)
		if err != nil {
			log.Print("gRPCServer.Worker stop with error: " + err.Error())
		} else {
			log.Print("gRPCServer.Worker stop successfully")
		}
	}()

	http.Handle("/api/v1/user", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlersEnv.UserProfile(w, r)
			case http.MethodOptions:
				return
			case http.MethodPost:
				params := r.URL.Query()
				temprary := false
				if isTemporary, ok := params["temporary"]; ok && len(isTemporary) == 1 {
					temprary = isTemporary[0] == "true"
				}
				if temprary {
					handlersEnv.RegistrationTemporary(w, r)
				} else {
					handlersEnv.RegistrationRegular(w, r)
				}
				return
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))

	// получить всех пользователей для доски лидеров
	http.Handle("/api/v1/users", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// setupResponse(w, r)
			switch r.Method {
			case http.MethodGet:
				handlersEnv.LeaderBoard(w, r)
			case http.MethodOptions:
				return
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))

	http.Handle("/api/v1/session", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlersEnv.Login(w, r)
			case http.MethodDelete:
				handlersEnv.Logout(w, r)
			case http.MethodOptions:
				return
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))

	http.Handle("/api/v1/avatar", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlersEnv.SetAvatar(w, r)
			case http.MethodOptions:
				return
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))

	fmt.Println("starting server at :8080")

	log.Println(http.ListenAndServe(":8080", nil))

	defer func() {
		err := handlersEnv.DB.Close()
		if err != nil {
			log.Fatal("failed to close Database connection: " + err.Error())
		}
	}()
	return
}
