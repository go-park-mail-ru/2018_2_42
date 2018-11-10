package response

import (
	"auth/models"

	"github.com/valyala/fasthttp"
)

func ErrorRequiredField(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest) // 400 Bad Request
	errorResp := models.ServerResponse{
		Status:  "400 Bad Request",
		Message: "field_temporary_required",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorMethodNotAllowed(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed) // 405 Method Not Allowed
	errorResp := models.ServerResponse{
		Status:  "405 Method Not Allowed",
		Message: "this_method_is_not_supported",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorNotAuthorized(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusForbidden) // 403 Forbidden
	errorResp := models.ServerResponse{
		Status:  "403 Forbidden",
		Message: "unauthorized_user",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorDataBase(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusInternalServerError) // 500 Internal Server
	errorResp := models.ServerResponse{
		Status:  "500 Internal Server",
		Message: "database_error",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorInvalidRequestFormat(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusBadRequest) // 400 Bad Request
	errorResp := models.ServerResponse{
		Status:  "400 Bad Request",
		Message: "invalid_request_format",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorBadLogin(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusUnprocessableEntity) // 422 Unprocessable Entity
	errorResp := models.ServerResponse{
		Status:  "422 Unprocessable Entity",
		Message: "empty_login",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorWeakPassword(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusUnprocessableEntity) // 422 Unprocessable Entity
	errorResp := models.ServerResponse{
		Status:  "422 Unprocessable Entity",
		Message: "weak_password",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}

func ErrorNotUniqLogin(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusUnprocessableEntity) // 422 Unprocessable Entity
	errorResp := models.ServerResponse{
		Status:  "422 Unprocessable Entity",
		Message: "login_is_not_unique",
	}
	buf, _ := errorResp.MarshalJSON()
	ctx.SetBody(buf)
}
