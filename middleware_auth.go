package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"rssagg-go/internal/database"

	"github.com/golang-jwt/jwt"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

// func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		apiKey, err := auth.GetAPIKey(r.Header)
// 		if err != nil {
// 			respondWithError(w, 403, fmt.Sprintf("Auth error: %v", err))
// 			return
// 		}

// 		user, err := apiCfg.DB.GetUserByAPIKey(r.Context(), apiKey)
// 		if err != nil {
// 			respondWithError(w, 400, fmt.Sprintf("Couldn't get user: %v", err))
// 			return
// 		}

// 		handler(w, r, user)
// 	}
// }

func (apiCfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//get the token
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			respondWithError(w, 403, "no token found")
			return
		}

		//decode and validate it

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil {
			respondWithError(w, http.StatusForbidden, "token is expired or not valid")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			//check the expiry time
			if float64((time.Now().Unix())) > claims["exp"].(float64) {
				respondWithError(w, http.StatusForbidden, "token expired")
				return
			}

			//find the user
			user, err := apiCfg.DB.GetUserByEmail(r.Context(), claims["email"].(string))
			if err != nil {
				if err == sql.ErrNoRows {
					respondWithError(w, http.StatusForbidden, "user not found")
					return
				}
				respondWithError(w, http.StatusInternalServerError, "Server error")
				return
			}

			//pass the user to the handler
			handler(w, r, user)
		} else {
			respondWithError(w, http.StatusForbidden, "token is not valid")
		}

	}
}
