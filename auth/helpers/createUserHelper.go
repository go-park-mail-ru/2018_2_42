package helpers

import (
	"auth/errors"
	"auth/models"
	"log"

	"github.com/valyala/fasthttp"
)

func CreateRegularUser(ctx *fasthttp.RequestCtx) {
	var userRegistreation models.UserRegistration

	err := userRegistreation.UnmarshalJSON(ctx.PostBody())
	if err != nil {
		log.Printf("Cannot parse request json: %s", err)
		ctx.SetStatusCode(fasthttp.StatusBadRequest) // 400 Bad Request
		errors.ErrorInvalidRequestFormat(ctx)
	}

	if len(userRegistreation.Login) == 0 {
		errors.ErrorBadLogin(ctx)
	}
	if len(userRegistreation.Password) < 5 {
		errors.ErrorBadPassword(ctx)
	}

}

func CreateTemporaryUser(ctx *fasthttp.RequestCtx) {

}
