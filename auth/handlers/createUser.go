package handlers

import (
	"auth/response"
	"auth/database"
	"auth/models"
	"auth/helpers"
	"log"
	"time"

	"github.com/jackc/pgx"
	"github.com/valyala/fasthttp"
)

const (
	defaultAvatarUrl = "/images/default.png"
)

func CreateUser(ctx *fasthttp.RequestCtx) {
	var userRegistreation models.UserRegistration
	err := userRegistreation.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		log.Printf("Cannot parse request json: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest) // 400 Bad Request
		response.ErrorInvalidRequestFormat(ctx)
		return
	}

	if len(userRegistreation.Login) == 0 {
		response.ErrorBadLogin(ctx)
		return
	}
	if len(userRegistreation.Password) < 5 {
		response.ErrorWeakPassword(ctx)
		return
	}

	userID, err := database.InsertIntoUser(userRegistreation.Login, helpers.Sha256hash(userRegistreation.Password), defaultAvatarUrl, false)
	if err != nil {
		if _, ok := err.(pgx.PgError); ok {
			response.ErrorNotUniqLogin(ctx)
			return
		}
		log.Println(err)
		response.ErrorDataBase(ctx)
		return
	}

	err = database.InsertIntoGameStatistics(userID, 0, 0)
	if err != nil {
		log.Println(err)
		response.ErrorDataBase(ctx)
		return
	}

	authorizationToken := helpers.RandomToken()
	err = database.InsertIntoCurrentLogin(userID, authorizationToken)
	if err != nil {
		log.Println(err)
		response.ErrorDataBase(ctx)
		return
	}

	var cookie fasthttp.Cookie
	cookie.SetKey("SessionId")
	cookie.SetValue(authorizationToken)
	cookie.SetHTTPOnly(true)
	// cookie.SetSecure(true)
	cookie.SetExpire(time.Now().AddDate(0, 1, 0))
	ctx.Response.Header.SetCookie(&cookie)
	response.SuccessRegistration(ctx)
}
