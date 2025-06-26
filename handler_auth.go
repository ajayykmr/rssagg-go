package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"rssagg-go/internal/database"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (apiCfg *apiConfig) handlerSignUp(w http.ResponseWriter, r *http.Request) {
	//get email/password From body
	type parameters struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	//Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error hashing password: %v", err))
	}

	//check if user exists
	existingUser, err := apiCfg.DB.GetUserByEmail(r.Context(), params.Email)

	if err == nil && existingUser.Email == params.Email {
		respondWithError(w, http.StatusConflict, "Email already exists")
		return
	}

	//Create the user
	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		Email:     params.Email,
		Password:  sql.NullString{String: string(hash), Valid: true},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error creating user: %v", err))
		return
	}

	log.Println("User Created: ", user.Email)
	//Respond
	respondWithJSON(w, http.StatusCreated, databaseUserToUser(user))
}

func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	//Get email and password from body
	type parameters struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	validate := validator.New()
	if err := validate.Struct(params); err != nil {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%v", err))
		return
	}

	//Look up  requested User
	user, err := apiCfg.DB.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "User not found, Please sign up first")
			return
		}
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error getting user: %v", err))
		return
	}

	//Compare sent in pass with saved pass hash
	if !user.Password.Valid {
		//if user signed up via google or other providers
		respondWithError(w, http.StatusUnauthorized, "Password not set. Please reset your password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(params.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid password")
		return
	}

	//Create JWT
	secretString := os.Getenv("JWT_SECRET")
	if secretString == "" {
		log.Fatal("JWT_SECRET is not found in the environment")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Minute * 60 * 24 * 30).Unix(), // Token expiration time
	})

	tokenString, err := token.SignedString([]byte(secretString)) //FINAL JWT TOKEN STRING
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error signing token: %v", err))
		return
	}

	// Set cookie or send token in response
	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		MaxAge:   3600 * 24 * 60, //60 days
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	//send it back
	type response struct {
		User
		Token string `json:"token"`
	}
	res := response{
		User:  databaseUserToUser(user),
		Token: tokenString,
	}
	respondWithJSON(w, http.StatusOK, res)
}
