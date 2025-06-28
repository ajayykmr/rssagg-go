package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"rssagg-go/initializers"
	"rssagg-go/internal/database"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func init() {
	// LoadEnvVariables()
	initializers.LoadEnvVariables()
}

var startTime = time.Now() //to track uptime

func main() {

	// Import Port number
	portString := os.Getenv("PORT")
	if portString == "" {
		log.Fatal("PORT is not found in the environment")
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not found in the environment")
	}

	// Connect to the database
	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Cannot connect to database")
	}

	db := database.New(conn)
	apiCfg := apiConfig{
		DB: db,
	}

	concurrency := 2
	go startScraping(apiCfg.DB, concurrency, time.Hour*24) //scraping every 5 minutes

	// Set up router
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	v1Router := chi.NewRouter()

	v1Router.Get("/healthz", handlerReadiness)
	v1Router.Get("/err", handlerErr)

	v1Router.Post("/signup", apiCfg.handlerSignUp)
	v1Router.Post("/login", apiCfg.handlerLogin)
	v1Router.Get("/user", apiCfg.middlewareAuth(apiCfg.handlerGetUser))

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.handlerCreateFeed))
	v1Router.Get("/feeds", apiCfg.handlerGetAllFeeds)
	v1Router.Delete("/feeds/{feedID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteFeed))
	v1Router.Get("/feeds/user", apiCfg.middlewareAuth(apiCfg.handlerGetUserCreatedFeeds))

	v1Router.Get("/feed-follows", apiCfg.middlewareAuth(apiCfg.handlerGetUserFeedFollows))
	v1Router.Post("/feed-follows", apiCfg.middlewareAuth(apiCfg.handlerCreateUserFeedFollow))
	v1Router.Delete("/feed-follows/{feedFollowID}", apiCfg.middlewareAuth(apiCfg.handlerDeleteUserFeedFollow))

	v1Router.Get("/posts/user", apiCfg.middlewareAuth(apiCfg.handlerGetPostsForUser))

	router.Mount("/v1", v1Router)
	router.Get("/", handlerReadiness)

	srv := &http.Server{
		Handler: router,
		Addr:    ":" + portString,
	}

	log.Printf("Server starting on port: http://localhost:%v", portString)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
