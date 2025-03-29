package main

import (
	"log"
	"net/http"
	"os"
	"time"
	
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	
	"auth_backend/internal/config"
	"auth_backend/internal/handlers"
)

func main() {
	_ = godotenv.Load()
	
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}
	defer db.Close()

	r := mux.NewRouter()
	handlers.RegisterAuthHandlers(r, db)
	
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Handler:      c.Handler(r),
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server running on port %s", port)
	log.Fatal(srv.ListenAndServe())
}
