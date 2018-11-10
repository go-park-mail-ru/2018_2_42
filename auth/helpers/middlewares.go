package helpers

import (
	"log"
	"time"

	"github.com/valyala/fasthttp"
)

func CommonHandlerMiddleware(handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered in CommonMiddlware: %s", r)
			}
		}()
		start := time.Now()
		ctx.SetContentType("application/json")
		handler(ctx)
		log.Printf("[%s] %s, %s\n", string(ctx.Method()), ctx.URI(), time.Since(start))
	})
}
