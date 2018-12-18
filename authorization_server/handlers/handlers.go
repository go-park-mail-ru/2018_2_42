// @title authorization server for technopark game.
// @version 1.0
// @description This is a registration server will be used for our game.
// @contact.email cup.of.software.code@gmail.com
// @BasePath /api/v1

package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2018_2_42/authorization_server/environment"
	"github.com/go-park-mail-ru/2018_2_42/authorization_server/types"
)

// прикрепляем функции с логикой к глобальному окружению, обеспечивая доступ к конфигу и базе данных
type Environment environment.Environment

func sha256hash(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

func init() {
	rand.Seed(time.Now().Unix())
}

func randomToken() string {
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

const defaultAvatarURL = "/images/default.png"

// RegistrationRegular godoc
// @Summary Regular user registration.
// @Description Registrate users with password and statistics.
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param registrationInfo body types.NewUserRegistration true "login password"
// @Success 201 {object} types.ServerResponse
// @Failure 400 {object} types.ServerResponse
// @Failure 409 {object} types.ServerResponse
// @Failure 422 {object} types.ServerResponse
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/user?temporary=false [post]
func (e *Environment) RegistrationRegular(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	registrationInfo := types.NewUserRegistration{}
	err = registrationInfo.UnmarshalJSON(bodyBytes)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "invalid_request_format",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	if registrationInfo.Login == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "empty_login",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	if len(registrationInfo.Password) < 5 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "weak_password",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	userID, err := e.DB.InsertIntoUser(registrationInfo.Login, defaultAvatarURL, false)
	if err != nil {
		if strings.Contains(err.Error(),
			`duplicate key value violates unique constraint "user_login_key"`) {
			w.WriteHeader(http.StatusConflict)
			response, _ := types.ServerResponse{
				Status:  http.StatusText(http.StatusConflict),
				Message: "login_is_not_unique",
			}.MarshalJSON()
			_, _ = w.Write(response)
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	err = e.DB.InsertIntoRegularLoginInformation(userID, sha256hash(registrationInfo.Password))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	err = e.DB.InsertIntoGameStatistics(userID, 0, 0)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// создаём токены авторизации.
	authorizationToken := randomToken()
	err = e.DB.UpsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// Уже нормальный ответ отсылаем.
	http.SetCookie(w, &http.Cookie{
		Name:  "SessionId",
		Value: authorizationToken,
		Path:  "/",
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#Permanent_cookie
		Expires:  time.Now().AddDate(0, 0, 7),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusCreated)

	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_reusable_registration",
	}.MarshalJSON()
	_, _ = w.Write(response)
}

// RegistrationTemporary godoc
// @Summary Temporary user registration.
// @Description Сreates user without statistics and password, stub so that you can play 1 session without creating an account.
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param registrationInfo body types.NewUserRegistration true "login password"
// @Success 200 {object} types.ServerResponse
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/user&temporary=true [post]
func (e *Environment) RegistrationTemporary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	registrationInfo := types.NewUserRegistration{}
	err = registrationInfo.UnmarshalJSON(bodyBytes)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "invalid_request_format",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	if registrationInfo.Login == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "empty_login",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	userID, err := e.DB.InsertIntoUser(registrationInfo.Login, defaultAvatarURL, true)
	if err != nil {
		if strings.Contains(err.Error(),
			`duplicate key value violates unique constraint "user_login_key"`) {
			w.WriteHeader(http.StatusConflict)
			response, _ := types.ServerResponse{
				Status:  http.StatusText(http.StatusConflict),
				Message: "login_is_not_unique",
			}.MarshalJSON()
			_, _ = w.Write(response)
			return
		}
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// создаём токены авторизации.
	authorizationToken := randomToken()
	err = e.DB.UpsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// Уже нормальный ответ отсылаем.
	http.SetCookie(w, &http.Cookie{
		Name:     "SessionId",
		Value:    authorizationToken,
		Path:     "/",
		Expires:  time.Now().AddDate(0, 0, 7),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusCreated)
	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_disposable_registration",
	}.MarshalJSON()
	_, _ = w.Write(response)
}

// LeaderBoard godoc
// @Summary Get liderboard with best user information.
// @Description Return login, avatarAddress, gamesPlayed and wins information for earch user.
// @Tags users
// @Accept application/json
// @Produce application/json
// @Param limit query int false "Lenth of returning user list."
// @Param offset query int false "Offset relative to the leader."
// @Success 200 {array} accessor.PublicUserInformation
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/users [get]
func (e *Environment) LeaderBoard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	getParams := r.URL.Query()
	_ = r.Body.Close()
	limit := 20
	if customLimitStrings, ok := getParams["limit"]; ok {
		if len(customLimitStrings) == 1 {
			if customLimitInt, err := strconv.Atoi(customLimitStrings[0]); err == nil {
				limit = customLimitInt
			}
		}
	}
	offset := 0
	if customOffsetStrings, ok := getParams["offset"]; ok {
		if len(customOffsetStrings) == 1 {
			if customOffsetInt, err := strconv.Atoi(customOffsetStrings[0]); err == nil {
				offset = customOffsetInt
			}
		}
	}

	LeaderBoard, err := e.DB.SelectLeaderBoard(limit, offset)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// Уже нормальный ответ отсылаем.
	w.WriteHeader(http.StatusOK)
	response, _ := LeaderBoard.MarshalJSON()
	_, _ = w.Write(response)
}

// UserProfile godoc
// @Summary Get user information.
// @Description Return login, avatarAddress, gamesPlayed, wins, information.
// @Tags user
// @Accept application/json
// @Produce application/json
// @Param login query string true "login password"
// @Success 200 {object} accessor.PublicUserInformation
// @Failure 422 {object} types.ServerResponse
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/user [get]
func (e *Environment) UserProfile(w http.ResponseWriter, r *http.Request) {
	getParams := r.URL.Query()
	_ = r.Body.Close()
	login := ""
	if loginStrings, ok := getParams["login"]; ok {
		if len(loginStrings) == 1 {
			if login = loginStrings[0]; login != "" {
				// just working on
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				response, _ := types.ServerResponse{
					Status:  http.StatusText(http.StatusUnprocessableEntity),
					Message: "empty_login_field",
				}.MarshalJSON()
				_, _ = w.Write(response)
				return
			}
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			response, _ := types.ServerResponse{
				Status:  http.StatusText(http.StatusUnprocessableEntity),
				Message: "login_must_be_only_1",
			}.MarshalJSON()
			_, _ = w.Write(response)
			return
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "field_login_required",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}

	userProfile, err := e.DB.SelectUserByLogin(login)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	// нормальный ответ
	w.WriteHeader(http.StatusOK)
	response, _ := userProfile.MarshalJSON()
	_, _ = w.Write(response)
}

// Login godoc
// @Summary Login into account.
// @Description Set cookie on client and save them in database.
// @Tags session
// @Accept application/json
// @Produce application/json
// @Param registrationInfo body types.NewUserRegistration true "login password"
// @Success 202 {object} types.ServerResponse
// @Failure 400 {object} types.ServerResponse
// @Failure 403 {object} types.ServerResponse
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/session [post]
func (e *Environment) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	_ = r.Body.Close()
	registrationInfo := types.NewUserRegistration{}
	err = registrationInfo.UnmarshalJSON(bodyBytes)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "invalid_request_format",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	exists, userId, err := e.DB.SelectUserIdByLoginPasswordHash(registrationInfo.Login, sha256hash(registrationInfo.Password))
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	if exists {
		authorizationToken := randomToken()
		err = e.DB.UpsertIntoCurrentLogin(userId, authorizationToken)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			response, _ := types.ServerResponse{
				Status:  http.StatusText(http.StatusInternalServerError),
				Message: "database_error",
			}.MarshalJSON()
			_, _ = w.Write(response)
			return
		}
		// Уже нормальный ответ отсылаем.
		http.SetCookie(w, &http.Cookie{
			Name:     "SessionId",
			Value:    authorizationToken,
			Path:     "/",
			Expires:  time.Now().AddDate(0, 0, 7),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusAccepted)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusAccepted),
			Message: "successful_password_login",
		}.MarshalJSON()
		_, _ = w.Write(response)
	} else {
		w.WriteHeader(http.StatusForbidden)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusFailedDependency),
			Message: "wrong_login_or_password",
		}.MarshalJSON()
		_, _ = w.Write(response)
	}
}

// Logout godoc
// @Summary Log registered user out.
// @Description Delete cookie in client and database.
// @Tags session
// @Accept application/json
// @Produce application/json
// @Success 200 {object} types.ServerResponse
// @Failure 404 {object} types.ServerResponse
// @Failure 401 {object} types.ServerResponse
// @Router /api/v1/session [delete]
func (e *Environment) Logout(w http.ResponseWriter, r *http.Request) {
	//get sid from cookies
	inCookie, err := r.Cookie("SessionId")
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusUnauthorized)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusUnauthorized),
			Message: "unauthorized_user",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	err = e.DB.DropUsersSession(inCookie.Value)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusNotFound),
			Message: "target_session_not_found",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "SessionId",
		Expires:  time.Unix(0, 0),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusOK),
		Message: "successful_logout",
	}.MarshalJSON()
	_, _ = w.Write(response)
}

// Корень, куда сохраняются аватарки и прочие загружаемые пользователем ресурсы.
const mediaRoot = "/var/www/media"

func init() {
	if _, err := os.Stat(mediaRoot + "/images"); os.IsNotExist(err) {
		err = os.MkdirAll(mediaRoot+"/images", os.ModePerm)
	}
}

// Logout godoc
// @Summary Upload user avatar.
// @Description Upload avatar from \<form enctype='multipart/form-data' action='/api/v1/avatar'>\<input type="file" name="avatar"></form>.
// @Tags avatar
// @Accept multipart/form-data
// @Produce application/json
// @Success 201 {object} types.ServerResponse
// @Failure 400 {object} types.ServerResponse
// @Failure 401 {object} types.ServerResponse
// @Failure 500 {object} types.ServerResponse
// @Router /api/v1/avatar [post]
func (e *Environment) SetAvatar(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()
	w.Header().Set("Content-Type", "application/json")
	//get SessionId from cookies
	cookie, err := r.Cookie("SessionId")
	if err != nil || cookie.Value == "" {
		log.Print(err)
		w.WriteHeader(http.StatusForbidden)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusForbidden),
			Message: "unauthorized_user",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	exist, user, err := e.DB.SelectUserBySessionId(cookie.Value)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "cannot_create_file",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	if !exist {
		w.WriteHeader(http.StatusForbidden)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusForbidden),
			Message: "unauthorized_user",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	err = r.ParseMultipartForm(0)
	if err != nil {
		log.Print("handlers SetAvatar ParseMultipartForm: " + err.Error())
		return
	}
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "cannot_get_file",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	defer func() { _ = file.Close() }()
	// /var/www/media/images/login.jpeg
	fileName := user.Login + filepath.Ext(handler.Filename)
	f, err := os.Create(mediaRoot + "/images/" + fileName)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "cannot_create_file",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	defer func() { _ = f.Close() }()
	//put avatar path to db
	err = e.DB.UpdateUsersAvatarByLogin(user.Login, "/media/images/"+fileName)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		response, _ := types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		}.MarshalJSON()
		_, _ = w.Write(response)
		return
	}
	_, _ = io.Copy(f, file)
	w.WriteHeader(http.StatusCreated)
	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_avatar_uploading",
	}.MarshalJSON()
	_, _ = w.Write(response)
}

func (e *Environment) ErrorMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.WriteHeader(http.StatusMethodNotAllowed)
	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusMethodNotAllowed),
		Message: "this_method_is_not_supported",
	}.MarshalJSON()
	_, _ = w.Write(response)
}

func (e *Environment) ErrorRequiredField(w http.ResponseWriter, r *http.Request) {
	_ = r.Body.Close()
	w.WriteHeader(http.StatusBadRequest)
	response, _ := types.ServerResponse{
		Status:  http.StatusText(http.StatusBadRequest),
		Message: "field_'temporary'_required",
	}.MarshalJSON()
	_, _ = w.Write(response)
}
