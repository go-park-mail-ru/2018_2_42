package handlers

import (
	"auth/database"
	"auth/response"
	"time"

	"github.com/valyala/fasthttp"
)

func DeleteSession(ctx *fasthttp.RequestCtx) {
	recievedCookie := ctx.Request.Header.Cookie("SessionId")
	// log.Println(string(recievedCookie))
	// if len(recievedCookie) == 0 {
	// 	response.ErrorNotAuthorized(ctx)
	// 	return
	// }
	err := database.DropUserSession(string(recievedCookie))
	if err != nil {
		response.ErrorSessionNotFound(ctx)
		return
	}

	cookie := fasthttp.Cookie{}
	cookie.SetHTTPOnly(true)
	cookie.SetSecure(true)
	cookie.SetExpire(time.Now().AddDate(0, 0, 0))
	cookie.SetKey("SessionId")
	ctx.Response.Header.SetCookie(&cookie)
	response.SuccessLogout(ctx)
}
