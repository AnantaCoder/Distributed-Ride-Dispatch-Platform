package middleware

import (
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type responseRecorder struct{
	http.ResponseWriter
	body  []byte
	statusCode int
}

func(r *responseRecorder) Write(b []byte) (int, error){
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func(r *responseRecorder) WriteHeader(statusCode int){
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func IdempotencyMiddleware(redisClient *redis.Client) func(next http.Handler) http.Handler{

	return func(next http.Handler) http.Handler{

		return http.HandlerFunc(func(w http.ResponseWriter , r *http.Request){

			idemKey := r.Header.Get("Idempotency-Key")

			if idemKey == ""{
				http.Error(w, "missing Idempotency-Key", http.StatusBadRequest)
				return 
			}

			ctx := r.Context()
			res, err := redisClient.Get(ctx, "idempotency:"+idemKey).Result()

			if err == nil {
				w.Write([]byte(res))
				return
			}

			rec := &responseRecorder{
				ResponseWriter: w,
				statusCode: http.StatusOK,
			}
			next.ServeHTTP(rec,r)

			if rec.statusCode >= 200 && rec.statusCode <= 300 {
				redisClient.Set(ctx,"idempotency:"+idemKey , string(rec.body) , 24*time.Hour)
			}
		})
	}
}

/*
Summary of what this Idempotency Middleware does:
1. Extracts the "Idempotency-Key" from incoming HTTP headers.
2. If missing, it ignores the check and passes the request through.
3. Checks Redis to see if we have processed this exact key before.
4. If YES (cache hit), it returns the saved JSON response instantly (blocks duplicates).
5. If NO (first time), it creates a custom ResponseRecorder to spy on the response.
6. Let's the actual API (e.g. Trip Service) process the request.
7. Saves the resulting JSON response to Redis for 24 hours so future duplicates are blocked.
*/