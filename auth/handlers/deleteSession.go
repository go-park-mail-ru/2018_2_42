package handlers

import (
	"auth/database"
	"auth/helpers"
	"time"

	"github.com/valyala/fasthttp"
)

// DeleteSession deletes a session of user.
func DeleteSession(ctx *fasthttp.RequestCtx) {
	recievedCookie := ctx.Request.Header.Cookie("SessionId")
	if len(recievedCookie) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusForbidden, "403 Forbidden", "user_is_unauthorized")
		return
	}

	err := database.DropUserSession(string(recievedCookie))
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusForbidden, "404 Not Found", "session_not_found")
		return
	}

	cookie := fasthttp.Cookie{}
	cookie.SetHTTPOnly(true)
	cookie.SetExpire(time.Now().AddDate(0, 0, 0))
	cookie.SetKey("SessionId")

	ctx.Response.Header.SetCookie(&cookie)
	helpers.ServerResponse(ctx, fasthttp.StatusOK, "200 OK", "successful_sign_out")
}
