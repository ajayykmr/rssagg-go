package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"rssagg-go/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type feedListResponse struct {
	Feeds []Feed `json:"feeds"`
}

func (apiCfg *apiConfig) handlerGetAllFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := apiCfg.DB.GetAllFeeds(r.Context())
	if err != nil {
		respondWithError(w, 404, fmt.Sprintln("Could not get feeds"))
		return
	}

	respondWithJSON(w, 200, feedListResponse{Feeds: databaseFeedsToFeeds(feeds)})
}

func (apiConfig *apiConfig) handlerGetUserCreatedFeeds(w http.ResponseWriter, r *http.Request, user database.User) {
	feeds, err := apiConfig.DB.GetFeedsByUserID(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("Could not get feeds: %v", err))
		return
	}

	respondWithJSON(w, 200, feedListResponse{Feeds: databaseFeedsToFeeds(feeds)})
}

func (apiCfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	feed, err := apiCfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.URL,
		UserID:    user.ID,
	})

	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Couldn't create feed: %v", err))
		return
	}

	respondWithJSON(w, 201, databaseFeedToFeed(feed))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go scrapeFeed(apiCfg.DB, wg, feed)
}

func (apiCfg *apiConfig) handlerDeleteFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	// Get the feed ID from the request URL

	///feeds/{feedID}
	// feedID := r.URL.Query().Get("feedID") //try this once,probably wont work
	feedID := chi.URLParam(r, "feedID") //this works
	if feedID == "" {
		respondWithError(w, 400, "Missing feed ID")
		return
	}

	parsedFeedID, err := uuid.Parse(feedID)
	if err != nil {
		respondWithError(w, 400, "Invalid feed ID")
		return
	}
	// Delete the feed from the database
	err = apiCfg.DB.DeleteFeed(r.Context(), database.DeleteFeedParams{
		ID:     parsedFeedID,
		UserID: user.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w, 200, "Feed not found/already deleted")
			return
		}
		respondWithError(w, 500, fmt.Sprintf("Failed to delete feed: %v", err))
		return
	}

	respondWithJSON(w, http.StatusAccepted, "Feed deleted successfully")
}
