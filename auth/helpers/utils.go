package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
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

func Sha256hash(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	return hex.EncodeToString(hasher.Sum(nil))
}

// func init() {
// 	rand.Seed(time.Now().Unix())
// }

func RandomToken() string {
	cookieChars := []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz+_")
	result := make([]byte, 20)
	for i := 0; i < 20; {
		key := rand.Uint64()
		for j := 0; j < 10; i, j = i+1, j+1 {
			result[i] = cookieChars[key&63]
			key >>= 6
		}
	}
	return string(result)
}
