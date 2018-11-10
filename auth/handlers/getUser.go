package handlers

import (
	"auth/database"
	"auth/response"

	"github.com/valyala/fasthttp"
)

func GetUser(ctx *fasthttp.RequestCtx) {
	login := string(ctx.FormValue("login"))
	if len(login) == 0 {
		response.ErrorEmptyLoginField(ctx)
	}

	userProfile, err := database.SelectUserByLogin(login)
	if err != nil {
		response.ErrorUserNotFound(ctx)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK) // 200 Status OK
	buf, _ := userProfile.MarshalJSON()
	ctx.SetBody(buf)
}
