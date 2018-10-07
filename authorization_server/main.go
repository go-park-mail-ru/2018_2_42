package main

import (
	"authorization_server/accessor"
	"authorization_server/handlers"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.Handle("/api/v1/user", handlers.CORSMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlers.UserProfile(w, r)
			case http.MethodOptions:
				return
			case http.MethodPost:
				params := r.URL.Query()
				if isTemporary, ok := params["temporary"]; ok {
					if len(isTemporary) == 1 {
						switch isTemporary[0] {
						case "true":
							handlers.RegistrationTemporary(w, r)
							return
						case "false":
							handlers.RegistrationRegular(w, r)
							return
						}
					}
				}
				handlers.ErrorRequiredField(w, r)
			default:
				handlers.ErrorMethodNotAllowed(w, r)
			}
		})))

	// получить всех пользователей для доски лидеров
	http.Handle("/api/v1/users", handlers.CORSMiddleware(
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

	http.Handle("/api/v1/session", handlers.CORSMiddleware(
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

	http.Handle("/api/v1/avatar", handlers.CORSMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				handlers.SetAvatar(w, r)
			case http.MethodOptions:
				return
			default:
				handlers.ErrorMethodNotAllowed(w, r)
			}
		})))

	fmt.Println("starting server at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("failed to start server at :8080 : " + err.Error())
	}

	err = accessor.Db.Close()
	if err != nil {
		log.Fatal("failed to start server at :8080 : " + err.Error())
	}
}
