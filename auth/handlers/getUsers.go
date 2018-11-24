package handlers

import (
	"auth/database"
	"auth/helpers"

	"github.com/valyala/fasthttp"
)

// GetUsers returns a list of users which ordered by wins.
func GetUsers(ctx *fasthttp.RequestCtx) {
	limit := string(ctx.FormValue("limit"))
	offset := string(ctx.FormValue("offset"))

	if len(limit) == 0 || len(offset) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "limit_or_offset_is_empty")
	}

	leaderBoard, err := database.SelectLeaderBoard(limit, offset)
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK) // 200 Status OK
	buf, _ := leaderBoard.MarshalJSON()
	ctx.SetBody(buf)
}
