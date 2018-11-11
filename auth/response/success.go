package response

import (
	"auth/models"

	"github.com/valyala/fasthttp"
)

func SuccessRegistration(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK) // 200 OK
	errorResp := models.ServerResponse{
		Status:  "200 OK",
		Message: "successful_sign_up",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func SuccessLogin(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK) // 200 OK
	errorResp := models.ServerResponse{
		Status:  "200 OK",
		Message: "successful_sign_in",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func SuccessLogout(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK) // 200 OK
	errorResp := models.ServerResponse{
		Status:  "200 OK",
		Message: "successful_sign_out",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}
