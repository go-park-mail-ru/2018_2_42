package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"
	flag "github.com/spf13/pflag" // ради gnu style: --flag='value'

	"github.com/go-park-mail-ru/2018_2_42/authorization_server/accessor"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/environment"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/handlers"
)

func registerUsersHandlers(handlersEnv handlers.Environment) {
	http.Handle("/api/v1/users", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// setupResponse(w, r)
			switch r.Method {
			case http.MethodGet:
				handlersEnv.LeaderBoard(w, r)
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))
}

func registerSessionHandlers(handlersEnv handlers.Environment) {
	http.Handle("/api/v1/session", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlersEnv.Login(w, r)
			case http.MethodDelete:
				handlersEnv.Logout(w, r)
			case http.MethodGet:
				handlersEnv.CheckSession(w, r)
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))
}

func registerAvatarHandlers(handlersEnv handlers.Environment) {
	http.Handle("/api/v1/avatar", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlersEnv.SetAvatar(w, r)
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))
}

func registerUserHandlers(handlersEnv handlers.Environment) {
	http.Handle("/api/v1/user", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlersEnv.UserProfile(w, r)
			case http.MethodPost:
				params := r.URL.Query()
				temporary := false
				if isTemporary, ok := params["temporary"]; ok && len(isTemporary) == 1 {
					temporary = isTemporary[0] == "true"
				}
				if temporary {
					handlersEnv.RegistrationTemporary(w, r)
				} else {
					handlersEnv.RegistrationRegular(w, r)
				}
				return
			default:
				handlersEnv.ErrorMethodNotAllowed(w, r)
			}
		}))
}

func main() {
	// получаем конфигурацию из аргументов командной строки
	env := environment.Environment{}
	env.Config.ListeningPort = flag.String(
		"listening-port",
		"8080",
		"port on which the server will listen")
	env.Config.PostgresPath = flag.String(
		"postgres-path",
		"postgres://postgres:@127.0.0.1:5432/postgres?sslmode=disable",
		"full postgres address like 'postgres://postgres:1@127.0.0.1:5432/postgres?sslmode=disable'")
	env.Config.ImagesRoot = flag.String(
		"images-root",
		"/var/www/media/images",
		"the folder in which the downloaded avatars of users will be saved")
	flag.Parse()

	// подключаемся к базе.
	var err error
	handlersEnv := handlers.Environment(env)
	handlersEnv.DB, err = accessor.ConnectToDatabase(*handlersEnv.Config.PostgresPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "accessor.ConnectToDatabase: "))
	}
	err = handlersEnv.DB.InitDatabase()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		err := handlersEnv.DB.Close()
		if err != nil {
			log.Fatal("failed to close Database connection: " + err.Error())
		}
	}()

	// регистрируем обработчики запросов с логикой сервера.
	registerUserHandlers(handlersEnv)
	registerUsersHandlers(handlersEnv)
	registerSessionHandlers(handlersEnv)
	registerAvatarHandlers(handlersEnv)

	// начинаем слушать порт.
	fmt.Println("starting server at :" + *env.Config.ListeningPort)
	log.Println(http.ListenAndServe(":"+*env.Config.ListeningPort, nil))

	return
}
