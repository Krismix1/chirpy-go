package main

import (
	"chirpy/internal/database"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	tokenSecret := os.Getenv("TOKEN_SECRET")
	if tokenSecret == "" {
		log.Fatal("TOKEN_SECRET envvar must be set")
		return
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL envvar must be set")
		return
	}
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Failed to parse DB URL: %s\n", err)
		return
	}
	db := stdlib.OpenDB(*config)
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %s\n", err)
		return
	}

	dbQueries := database.New(db)

	apiCfg := apiConfig{dbQueries: dbQueries, platform: os.Getenv("PLATFORM"), tokenSecret: tokenSecret}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", handlerHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerListAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.GetChirpById)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
