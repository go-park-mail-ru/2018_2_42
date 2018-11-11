package handlers

import (
	"auth/database"
	"auth/helpers"
	"auth/models"
	"auth/response"
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

func CreateSession(ctx *fasthttp.RequestCtx) {
	var userRegistreation models.UserRegistration
	err := userRegistreation.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		log.Printf("Cannot parse request json: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest) // 400 Bad Request
		response.ErrorInvalidRequestFormat(ctx)
		return
	}

	userID, err := database.SelectUserIdByLoginPasswordHash(userRegistreation.Login, helpers.Sha256hash(userRegistreation.Password))
	if err != nil {
		response.ErrorUserNotFound(ctx)
	}

	authorizationToken := helpers.RandomToken()
	err = database.InsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		response.ErrorUserNotFound(ctx)
		return
	}

	cookie := fasthttp.Cookie{}
	cookie.SetHTTPOnly(true)
	cookie.SetSecure(true)
	cookie.SetExpire(time.Now().AddDate(0, 1, 0))
	cookie.SetValue(authorizationToken)
	cookie.SetKey("SessionId")
	ctx.Response.Header.SetCookie(&cookie)
	response.SuccessLogin(ctx)
}
