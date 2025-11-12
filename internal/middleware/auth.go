package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

// func JWTMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		tokenString := strings.Replace(r.Header.Get("Authorization"), "Bearer ", "", 1)

// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 			}
// 			return []byte("mySecretKey"), nil // Важно использовать тот же секретный ключ для проверки токена
// 		})

// 		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 			// Успешная аутентификация, продолжаем обработку запроса
// 			r.Header.Set("user_id", claims["sub"].(string))
// 			next.ServeHTTP(w, r)
// 		} else {
// 			// Аутентификация не удалась
// 			http.Error(w, "Forbidden", http.StatusForbidden)
// 		}
// 	})
// // }

func JWTMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if userID, exists := claims["user_id"]; exists {
					ctx := context.WithValue(r.Context(), "userID", userID.(string))
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					http.Error(w, "Token does not contain user ID", http.StatusUnauthorized)
					return
				}
			} else {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}
		})
	}
}
