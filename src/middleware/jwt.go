package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"kadane.xyz/go-backend/v2/src/config"
	"kadane.xyz/go-backend/v2/src/cookie"
)

type contextKey string

const UserContextKey contextKey = "user"

// JWTDecodeMiddleware will decode the JWT token and set the claims in the context
func JWTDecodeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("JWTDecodeMiddleware")
		cookieToken, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		tokenStr := cookieToken.Value
		claims := &cookie.CookieStruct{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return config.Ed25519PublicKey, nil
		})

		if err != nil || !token.Valid {
			next.ServeHTTP(w, r)
			return
		}

		if exp, ok := claims.MapClaims["exp"].(float64); ok {
			if time.Unix(int64(exp), 0).Before(time.Now()) {
				next.ServeHTTP(w, r)
				return
			}
		}

		//Token is valid setting context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("JWTAuthMiddleware")
		claims, ok := r.Context().Value(UserContextKey).(*cookie.CookieStruct)
		if !ok || claims.UserID == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
