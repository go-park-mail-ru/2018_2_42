package handlers

import (
	"auth/database"
	"auth/helpers"
	"log"
	"os"

	"github.com/valyala/fasthttp"
)

const mediaRoot = "/var/www/media"

func init() {
	if _, err := os.Stat(mediaRoot + "/images"); os.IsNotExist(err) {
		err = os.MkdirAll(mediaRoot+"/images", os.ModePerm)
	}
}

// SetAvatar updates user's avatar.
func SetAvatar(ctx *fasthttp.RequestCtx) {
	recievedCookie := ctx.Request.Header.Cookie("SessionId")
	log.Println(string(recievedCookie))
	if len(recievedCookie) == 0 {
		helpers.ServerResponse(ctx, fasthttp.StatusForbidden, "403 Forbidden", "user_is_unauthorized")
		return
	}

	_, err := database.SelectUserBySession(string(recievedCookie))
	if err != nil {
		helpers.ServerResponse(ctx, fasthttp.StatusNotFound, "404 Not Found", "user_not_found")
		return
	}

	// file, err := ctx.FormFile("avatar")
	// if err != nil {
	// 	fmt.Println(err)
	// 	helpers.ServerResponse(ctx, fasthttp.StatusBadRequest, "400 Bad Request", "cannot_get_file")
	// 	return
	// }

	// fmt.Println("hellp")
	// /var/www/media/images/login.jpeg
	// fileName := user.Login + filepath.Ext(mForm.File["avatar"])
	// fmt.Println(fileName)
	// path := mediaRoot + "/images/" + fileName
	// err = fasthttp.SaveMultipartFile(file, path)
	// if err != nil {
	// 	response.ErrorInvalidRequestFormat(ctx)
	// }

	// err = database.UpdateUsersAvatarByLogin(user.Login, "/media/images/" + fileName)
	// if err != nil {
	// 	log.Error(err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	json.NewEncoder(w).Encode(types.ServerResponse{
	// 		Status:  http.StatusText(http.StatusInternalServerError),
	// 		Message: "database_error",
	// 	})
	// 	return
	// }
	// io.Copy(f, file)
	// w.WriteHeader(http.StatusCreated)
	// json.NewEncoder(w).Encode(types.ServerResponse{
	// 	Status:  http.StatusText(http.StatusCreated),
	// 	Message: "successful_avatar_uploading",
	// })
	// return

}
