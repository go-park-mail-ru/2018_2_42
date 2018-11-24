package handlers

import (
	"auth/database"
	"auth/helpers"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/valyala/fasthttp"
)

// const mediaRoot = "/var/www/media"
const mediaRoot = "."

func init() {
	if _, err := os.Stat(mediaRoot + "/images"); os.IsNotExist(err) {
		err = os.MkdirAll(mediaRoot+"/images", os.ModePerm)
	}
}

// SetAvatar updates user's avatar.
// func SetAvatar(ctx *fasthttp.RequestCtx) {
// 	recievedCookie := ctx.Request.Header.Cookie("SessionId")
// 	log.Println(string(recievedCookie))
// 	if len(recievedCookie) == 0 {
// 		helpers.ServerResponse(ctx, fasthttp.StatusForbidden, "403 Forbidden", "user_is_unauthorized")
// 		return
// 	}

// 	user, err := database.SelectUserBySession(string(recievedCookie))
// 	if err != nil {
// 		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
// 		return
// 	}

// 	multipartFileReader, err := ctx.FormFile("avatar")
// 	if err != nil {
// 		fmt.Println(err)
// 		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "cannot_get_file")
// 		return
// 	}

// 	// /var/www/media/images/login.jpeg
// 	fileName := user.Login + filepath.Ext(multipartFileReader.Filename)
// 	fmt.Println(fileName)
// 	path := mediaRoot + "/images/" + fileName

// 	avatarFile, _ := multipartFileReader.Open()
// 	defer avatarFile.Close()

// 	f, err := os.Create(path)
// 	if err != nil {
// 		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "invalid_request_format")
// 	}
// 	defer f.Close()

// 	err = database.UpdateUsersAvatarByLogin(user.Login, "/media/images/"+fileName)
// 	if err != nil {
// 		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
// 		return
// 	}

// 	io.Copy(f, avatarFile)
// 	helpers.ServerResponse(ctx, fasthttp.StatusOK, "200 OK", "successful_avatar_uploading")
// }

func SetAvatar(ctx *fasthttp.RequestCtx) {
	recievedCookie := ctx.Request.Header.Cookie("SessionId")
	log.Println(string(recievedCookie))
	if len(recievedCookie) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusForbidden, "403 Forbidden", "user_is_unauthorized")
		return
	}

	user, err := database.SelectUserBySession(string(recievedCookie))
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
		return
	}

	multipartFileReader, err := ctx.MultipartForm()
	if err != nil {
		fmt.Println(err)
		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "cannot_get_file")
		return
	}

	// /var/www/media/images/login.jpeg
	// fileName := user.Login + filepath.Ext(multipartFileReader.Filename)
	// fmt.Println(fileName)
	// path := mediaRoot + "/images/" + fileName

	avatarFile, _ := multipartFileReader.File["avatar"][0].Open()
	defer avatarFile.Close()

	f, err := os.Create("./images/a.jpg")
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "invalid_request_format")
	}
	defer f.Close()

	err = database.UpdateUsersAvatarByLogin(user.Login, "/media/images/"+"fileName")
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusInternalServerError, "500 Internal Server", "database_error")
		return
	}

	io.Copy(f, avatarFile)
	helpers.ServerResponse(ctx, fasthttp.StatusOK, "200 OK", "successful_avatar_uploading")
}
