// @title authorization server for technopark game.
// @version 1.0
// @description This is a registration server will be used for our game.
// @contact.email cup.of.software.code@gmail.com
// @BasePath /api/v1

package handlers

import (
	"authorization_server/accessor"
	"authorization_server/types"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

func CommonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		next.ServeHTTP(w, r)
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered in CommonMiddlware: ", r)
			}
		}()
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//get SessionId from cookies
		cookie, err := r.Cookie("SessionId")
		if err != nil || cookie.Value == "" {
			log.Info("Cannot get cookie or it is empty", err)
			http.HandlerFunc(ErrorNotAuthorized).ServeHTTP(w, r)
			return
		}
		exist, err := accessor.Db.CheckAuthToken(cookie.Value)
		if err != nil {
			log.Error(err)
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(types.ServerResponse{
					Status:  http.StatusText(http.StatusInternalServerError),
					Message: "database_error",
				})
			}).ServeHTTP(w, r)
			return
		}
		if !exist {
			http.HandlerFunc(ErrorNotAuthorized).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

const defaultAvatarUrl = "/images/default.png"

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
func RegistrationRegular(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	registrationInfo := types.NewUserRegistration{}
	err := json.NewDecoder(r.Body).Decode(&registrationInfo)
	if err != nil {
		log.Info("Cannot parse request json: ", err)
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
			log.Error(err)
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
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	err = accessor.Db.InsertIntoGameStatistics(userId, 0, 0)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// создаём токены авторизации.
	authorizationToken := randomToken()
	err = accessor.Db.UpsertIntoCurrentLogin(userId, authorizationToken)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// Уже нормальный ответ отсылаем.
	http.SetCookie(w, &http.Cookie{
		Name:  "SessionId",
		Value: authorizationToken,
		// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie#Permanent_cookie
		Expires:  time.Now().AddDate(0, 1, 0),
		Secure:   false, // TODO: Научиться устанавливать https:// сертефикаты
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_reusable_registration",
	})
	return
}

// Регистрация пользователей временная.
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
func RegistrationTemporary(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	registrationInfo := types.NewUserRegistration{}
	err := json.NewDecoder(r.Body).Decode(&registrationInfo)
	if err != nil {
		log.Info("cannot parse request json: ", err)
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
	userId, err := accessor.Db.InsertIntoUser(registrationInfo.Login, defaultAvatarUrl, true)
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
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusInternalServerError),
				Message: "database_error",
			})
			return
		}
	}
	// создаём токены авторизации.
	authorizationToken := randomToken()
	err = accessor.Db.UpsertIntoCurrentLogin(userId, authorizationToken)
	if err != nil {
		log.Error(err)
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
		Expires:  time.Now().AddDate(0, 1, 0),
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
func LeaderBoard(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	getParams := r.URL.Query()
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

	LeaderBoard, err := accessor.Db.SelectLeaderBoard(limit, offset)
	if err != nil {
		log.Error(err)
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
func UserProfile(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	getParams := r.URL.Query()
	login := ""
	if loginStrings, ok := getParams["login"]; ok {
		if len(loginStrings) == 1 {
			if login = loginStrings[0]; login != "" {
				// just working on
			} else {
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).Encode(types.ServerResponse{
					Status:  http.StatusText(http.StatusUnprocessableEntity),
					Message: "empty_login_field",
				})
				return
			}
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(types.ServerResponse{
				Status:  http.StatusText(http.StatusUnprocessableEntity),
				Message: "login_must_be_only_1",
			})
			return
		}
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusUnprocessableEntity),
			Message: "field_login_required",
		})
		return
	}

	userProfile, err := accessor.Db.SelectUserByLogin(login)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	// нормальный ответ
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userProfile)
	return
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
func Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")

	registrationInfo := types.NewUserRegistration{}
	err := json.NewDecoder(r.Body).Decode(&registrationInfo)
	if err != nil {
		log.Info("Cannot parse request json", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "invalid_request_format",
		})
		return
	}
	exists, userId, err := accessor.Db.SelectUserIdByLoginPasswordHash(registrationInfo.Login, sha256hash(registrationInfo.Password))
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	if exists {
		authorizationToken := randomToken()
		err = accessor.Db.UpsertIntoCurrentLogin(userId, authorizationToken)
		if err != nil {
			log.Error(err)
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
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusAccepted),
			Message: "successful_password_login",
		})
	} else {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusFailedDependency),
			Message: "wrong_login_or_password",
		})
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
func Logout(w http.ResponseWriter, r *http.Request) {
	//get sid from cookies
	inCookie, err := r.Cookie("SessionId")
	if err != nil {
		log.Info(err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusUnauthorized),
			Message: "unauthorized_user",
		})
		return
	}
	err = accessor.Db.DropUsersSession(inCookie.Value)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusNotFound),
			Message: "target_session_not_found",
		})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "SessionId",
		Expires:  time.Unix(0, 0),
		Secure:   false, // TODO: Научиться устанавливать https:// сертефикаты
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusOK),
		Message: "successful_logout",
	})
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
func SetAvatar(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	//get SessionId from cookies
	cookie, err := r.Cookie("SessionId")
	_, user, err := accessor.Db.SelectUserBySessionId(cookie.Value)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	r.ParseMultipartForm(0)
	file, handler, err := r.FormFile("avatar")
	if err != nil {
		log.Info(err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusBadRequest),
			Message: "cannot_get_file",
		})
		return
	}
	defer file.Close()
	// /var/www/media/images/login.jpeg
	fileName := user.Login + filepath.Ext(handler.Filename)
	f, err := os.Create(mediaRoot + "/images/" + fileName)
	if err != nil {
		log.Error("Cannot create folder or file: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "cannot_create_file",
		})
		return
	}
	defer f.Close()
	//put avatar path to db
	err = accessor.Db.UpdateUsersAvatarByLogin(user.Login, "/media/images/" + fileName)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.ServerResponse{
			Status:  http.StatusText(http.StatusInternalServerError),
			Message: "database_error",
		})
		return
	}
	io.Copy(f, file)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusCreated),
		Message: "successful_avatar_uploading",
	})
	return
}

func ErrorMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.WriteHeader(http.StatusMethodNotAllowed)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusMethodNotAllowed),
		Message: "this_method_is_not_supported",
	})
}

func ErrorRequiredField(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusBadRequest),
		Message: "field_'temporary'_required",
	})
}

func ErrorNotAuthorized(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(types.ServerResponse{
		Status:  http.StatusText(http.StatusForbidden),
		Message: "unauthorized_user",
	})
}
