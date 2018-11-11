package handlers

import (
	"auth/database"
	"auth/response"

	"github.com/valyala/fasthttp"
)

func GetUsers(ctx *fasthttp.RequestCtx) {
	limit := string(ctx.FormValue("limit"))
	offset := string(ctx.FormValue("offset"))

	if len(limit) == 0 || len(offset) == 0 {
		response.ErrorRequiredField(ctx)
	}

	leaderBoard, err := database.SelectLeaderBoard(limit, offset)
	if err != nil {
		response.ErrorDataBase(ctx)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK) // 200 Status OK
	buf, _ := leaderBoard.MarshalJSON()
	ctx.SetBody(buf)
}
