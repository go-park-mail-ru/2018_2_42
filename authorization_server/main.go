package main

import (
	"authorization_server/accessor"
	"authorization_server/handlers"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(true)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
}

func main() {
	// http.Handle("/api/v1/user", handlers.CommonMiddleware(
	// 	http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		switch r.Method {
	// 		case http.MethodGet:
	// 			handlers.UserProfile(w, r)
	// 		case http.MethodPost:
	// 			params := r.URL.Query()
	// 			if isTemporary, ok := params["temporary"]; ok {
	// 				if len(isTemporary) == 1 {
	// 					switch isTemporary[0] {
	// 					case "true":
	// 						handlers.RegistrationTemporary(w, r)
	// 						return
	// 					case "false":
	// 						handlers.RegistrationRegular(w, r)
	// 						return
	// 					}
	// 				}
	// 			}
	// 			handlers.ErrorRequiredField(w, r)
	// 		default:
	// 			handlers.ErrorMethodNotAllowed(w, r)
	// 		}
	// 	})))

	// получить всех пользователей для доски лидеров
	http.Handle("/api/v1/users", handlers.CommonMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// setupResponse(w, r)
			switch r.Method {
			case http.MethodGet:
				handlers.LeaderBoard(w, r)
			case http.MethodOptions:
				return
			default:
				handlers.ErrorMethodNotAllowed(w, r)
			}
		})))

	http.Handle("/api/v1/session", handlers.CommonMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlers.Login(w, r)
			case http.MethodDelete:
				handlers.Logout(w, r)
			case http.MethodOptions:
				return
			default:
				handlers.ErrorMethodNotAllowed(w, r)
			}
		})))

	http.Handle("/api/v1/avatar", handlers.CommonMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlers.AuthMiddleware(http.HandlerFunc(handlers.SetAvatar)).ServeHTTP(w, r)
			case http.MethodOptions:
				return
			default:
				handlers.ErrorMethodNotAllowed(w, r)
			}
		})))

	log.Info("starting server at :8080")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal("failed to start server at :8080 : " + err.Error())
	}

	err = accessor.Db.Close()
	if err != nil {
		log.Fatal("failed to start server at :8080 : " + err.Error())
	}
}
