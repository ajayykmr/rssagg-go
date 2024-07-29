package main

import (
	"fmt"
	"net/http"
	"strconv"

	"rssagg-go/internal/database"
)

func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, 200, databaseUserToUser(user))
}

type postsResponse struct {
	Posts []Post `json:"posts"`
}

func (apiCfg *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	//get count from url parameters

	countStr := r.URL.Query().Get("count")

	count, err := strconv.ParseInt(countStr, 10, 32)
	if err != nil {
		count = 100 // Default value if count is not provided or is not a valid integer
	}

	posts, err := apiCfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(count),
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't get posts: %v", err))
		return
	}

	respondWithJSON(w, 200, postsResponse{Posts: databasePostsToPosts(posts)})
}
