package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/honesea/go-chirpy/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("could not load environment varibales")
		return
	}

	r := chi.NewRouter()
	admin := chi.NewRouter()
	api := chi.NewRouter()
	cfg := apiConfig{
		db:        database.NewDB(),
		jwtSecret: os.Getenv("JWT_SECRET"),
	}

	fileServer := cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))

	r.Handle("/app", http.StripPrefix("/app", fileServer))
	r.Handle("/app/*", http.StripPrefix("/app", fileServer))

	api.Get("/healthz", cfg.healthz)
	api.Get("/metrics", cfg.metrics)
	api.Get("/reset", cfg.reset)
	api.Get("/chirps", cfg.listChirps)
	api.Post("/chirps", cfg.createChirp)
	api.Get("/chirps/{chirp_id}", cfg.readChirp)
	api.Post("/users", cfg.createUser)
	api.Put("/users", cfg.updateUser)
	api.Post("/login", cfg.login)
	api.Post("/refresh", cfg.refresh)
	api.Post("/revoke", cfg.revoke)

	admin.Get("/metrics", cfg.adminMetrics)

	r.Mount("/api", api)
	r.Mount("/admin", admin)

	server := &http.Server{
		Addr:    ":3000",
		Handler: middlewareCors(r),
	}

	log.Println("server starting")
	err = server.ListenAndServe()

	if err != nil {
		log.Printf("error: %v\n", err)
		log.Println("server closing")
	}
}
