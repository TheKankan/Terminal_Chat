package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/TheKankan/TerminalSecuredChat/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	db        database.Store
	jwtSecret string
}

func main() {
	const port = "8080"

	// Getting .env variables
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	// Setting up connection to postgres database
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(dbConn)

	// Saving variables in config
	apiCfg := apiConfig{
		db:        dbQueries,
		jwtSecret: jwtSecret,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", apiCfg.handlerRegister)
	mux.HandleFunc("POST /login", apiCfg.handlerLogin)
	mux.HandleFunc("/ws", apiCfg.handlerWebSocket)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
