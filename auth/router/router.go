package router

import (
	"auth/handlers"

	"github.com/buaazp/fasthttprouter"
)

func NewRouter() *fasthttprouter.Router {
	router := fasthttprouter.New()

	router.GET("/api/v1/user", handlers.GetUser)     // done
	router.POST("/api/v1/user", handlers.CreateUser) // done
	router.POST("/api/v1/users", handlers.GetUsers)

	router.POST("/api/v1/avatar", handlers.SetAvatar)

	router.POST("/api/v1/session", handlers.CreateSession)
	router.DELETE("/api/v1/session", handlers.DeleteSession)

	return router
}
