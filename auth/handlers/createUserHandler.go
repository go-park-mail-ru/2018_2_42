package handlers

import (
	"auth/errors"

	"github.com/valyala/fasthttp"
)

// CreateUser creates regular/temporary user or returns StatusError.
func CreateUser(ctx *fasthttp.RequestCtx) {
	isTemporary := string(ctx.FormValue("temporary"))

	switch isTemporary {
	case "false":
		helpers.CreateRegularUser(ctx)
	case "true":
		helpers.CreateTemporaryUser(ctx)
	case "":
		errors.ErrorRequiredField(ctx)
	default:
		errors.ErrorMethodNotAllowed(ctx)
	}
}
