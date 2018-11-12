package helpers

import (
	"auth/models"

	"github.com/valyala/fasthttp"
)

func ServerResponse(ctx *fasthttp.RequestCtx, statusCode int, statusText, message string) {
	ctx.SetStatusCode(statusCode)
	errorResp := models.ServerResponse{
		Status:  statusText,
		Message: message,
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}
