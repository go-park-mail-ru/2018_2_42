// TODO: описать в swagger, https://github.com/swaggo/swag

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"authorization_server/accessor"
	"authorization_server/types"
)

func sha256hash(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

func init() {
	rand.Seed(time.Now().Unix())
}

func randomToken() (string) {
	cookieChars := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz+_")
	result := make([]byte, 20)
	for i := 0; i < 20; {
		key := rand.Uint64()
		for j := 0; j < 10; i, j = i+1, j+1 {
			result[i] = cookieChars[key&63]
			key >>= 6
		}
	}
	return string(result)
}

// Todo: написать конфиг nginx на отдачу статики.
const defaultAvatarUrl = "/static/images/default_avatar.jpg"

// Регистрация пользователей обычная.
func RegistrationRegular(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	registrationInfo := types.NewUserRegistration{}
	err := json.NewDecoder(r.Body).Decode(&registrationInfo)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "invalid_request_format",
		})
		return
	}
	if registrationInfo.Login == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "empty_login",
		})
		return
	}
	if len(registrationInfo.Password) < 5 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "weak_password",
		})
		return
	}
	userId, err := accessor.Db.InsertIntoUser(registrationInfo.Login, defaultAvatarUrl, false)
	if err != nil {
		if strings.Contains(err.Error(),
			`duplicate key value violates unique constraint "user_login_key"`) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusConflict),
				Message: "login_is_not_unique",
			})
			return
		} else {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusInternalServerError),
				Message: "database_error",
			})
			return
		}
	}
	err = accessor.Db.InsertIntoRegularLoginInformation(userId, sha256hash(registrationInfo.Password))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	err = accessor.Db.InsertIntoGameStatistics(userId, 0, 0)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// создаём токены авторизации.
	authorizationToken := randomToken()
	err = accessor.Db.InsertIntoCurrentLogin(userId, authorizationToken)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// Уже нормальный ответ отсылаем.
	http.SetCookie(w, &http.Cookie{
		Name:     "SessionId",
		Value:    authorizationToken,
		Secure:   false, // TODO: Научиться устанавливать https:// сертефикаты
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_disposable_registration",
	})
	return
}

func LeaderBoard(w http.ResponseWriter, r *http.Request)  {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	getParams := r.URL.Query()
	limit := 20
	if customLimitStrings, ok := getParams["limit"]; ok{
		if len(customLimitStrings) == 1{
			if customLimitInt, err := strconv.Atoi(customLimitStrings[0]); err == nil{
				limit = customLimitInt
			}
		}
	}
	offset := 0
	if customOffsetStrings, ok := getParams["offset"]; ok{
		if len(customOffsetStrings) == 1{
			if customOffsetInt, err := strconv.Atoi(customOffsetStrings[0]); err == nil{
				offset = customOffsetInt
			}
		}
	}

	LeaderBoard, err := accessor.Db.SelectLeaderBoard(limit, offset)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// Уже нормальный ответ отсылаем.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LeaderBoard)
	return
}

func main() {
	http.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			RegistrationRegular(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusMethodNotAllowed),
				Message: "this_method_is_not_supported",
			})
		}
	})
	// получить всех пользователей для доски лидеров
	http.HandleFunc("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			LeaderBoard(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusMethodNotAllowed),
				Message: "this_method_is_not_supported",
			})
		}
	})

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
