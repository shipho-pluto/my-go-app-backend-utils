package main

import (
	"context"
	"my-go-app/redis/cahce/handlers"
	"my-go-app/redis/cahce/storage"
	"net/http"
	"time"

	"github.com/go-chi/chi"
)

func main() {
	cfg := storage.Config{
		Addr:        "localhost:6379",
		Password:    "",
		User:        "",
		DB:          0,
		MaxRetries:  5,
		DialTimeout: 10 * time.Second,
		Timeout:     5 * time.Second,
	}

	db, err := storage.NewClient(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	router := chi.NewRouter()
	router.Route("/card", handlers.NewCardHandler(context.Background(), db))

	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
