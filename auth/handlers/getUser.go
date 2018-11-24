package handlers

import (
	"auth/database"
	"auth/helpers"

	"github.com/valyala/fasthttp"
)

// GetUser returns a user's profile.
func GetUser(ctx *fasthttp.RequestCtx) {
	login := string(ctx.FormValue("login"))
	if len(login) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusUnprocessableEntity, "422 Unprocessable Entity", "empty_login_field")
	}

	userProfile, err := database.SelectUserByLogin(login)
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK) // 200 Status OK
	buf, _ := userProfile.MarshalJSON()
	ctx.SetBody(buf)
}
