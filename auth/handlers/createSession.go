package handlers

import (
	"auth/database"
	"auth/helpers"
	"auth/models"
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

// CreateSession creates a session for user.
func CreateSession(ctx *fasthttp.RequestCtx) {
	var userRegistreation models.UserRegistration
	err := userRegistreation.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		log.Printf("Cannot parse request json: %s", err)
		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "invalid_request_format")
		return
	}

	userID, err := database.SelectUserIdByLoginPasswordHash(userRegistreation.Login, helpers.Sha256hash(userRegistreation.Password))
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
	}

	authorizationToken := helpers.RandomToken()
	err = database.InsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
		return
	}

	cookie := fasthttp.Cookie{}
	cookie.SetHTTPOnly(true)
	cookie.SetExpire(time.Now().AddDate(0, 1, 0))
	cookie.SetValue(authorizationToken)
	cookie.SetKey("SessionId")

	ctx.Response.Header.SetCookie(&cookie)
	helpers.ServerResponse(ctx, fasthttp.StatusOK, "200 OK", "successful_sign_in")
}
