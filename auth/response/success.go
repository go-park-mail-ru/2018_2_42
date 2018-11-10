package response

import (
	"auth/models"

	"github.com/valyala/fasthttp"
)

func SuccessRegistration(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK) // 200 OK
	errorResp := models.ServerResponse{
		Status:  "200 OK",
		Message: "successful_registration",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}
