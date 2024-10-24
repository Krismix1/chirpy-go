package main

import (
	"chirpy/internal/database"
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL envvar must be set")
		return
	}

	db, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %s\n", err)
		return
	}
	defer db.Close(context.Background())

	dbQueries := database.New(db)

	apiCfg := apiConfig{dbQueries: dbQueries}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.reset)
	mux.HandleFunc("POST /api/validate_chirp", handlerChirpsValidate)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
