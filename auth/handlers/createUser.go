package handlers

import (
	"auth/database"
	"auth/models"
	"auth/helpers"
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

const (
	defaultAvatarURL = "/images/default.png"
)

// CreateUser creates a new user.
func CreateUser(ctx *fasthttp.RequestCtx) {
	var userRegistreation models.UserRegistration
	err := userRegistreation.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		log.Printf("Cannot parse request json: %s", err)
		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "invalid_request_format")
		return
	}

	if len(userRegistreation.Login) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusUnprocessableEntity, "422 Unprocessable Entity", "empty_login")
		return
	}
	if len(userRegistreation.Password) < 5 {
		helpers.ServerResponse(ctx, fasthttp.StatusUnprocessableEntity, "422 Unprocessable Entity", "weak_password")
		return
	}

	userID, err := database.InsertIntoUser(userRegistreation.Login, helpers.Sha256hash(userRegistreation.Password), defaultAvatarURL, false)
	if err != nil {
		if _, ok := err.(pgx.PgError); ok {
			helpers.ServerResponse(ctx, fasthttp.StatusUnprocessableEntity, "422 Unprocessable Entity", "login_is_not_unique")
			return
		}
		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
		return
	}

	err = database.InsertIntoGameStatistics(userID, 0, 0)
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
		return
	}

	authorizationToken := helpers.RandomToken()
	err = database.InsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
		return
	}

	var cookie fasthttp.Cookie
	cookie.SetKey("SessionId")
	cookie.SetValue(authorizationToken)
	cookie.SetHTTPOnly(true)
	cookie.SetExpire(time.Now().AddDate(0, 1, 0))
	
	ctx.Response.Header.SetCookie(&cookie)
	helpers.ServerResponse(ctx, fasthttp.StatusOK, "200 OK", "successful_sign_up")
}
