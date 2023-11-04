package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/honesea/go-chirpy/internal/database"
)

func main() {
	r := chi.NewRouter()
	admin := chi.NewRouter()
	api := chi.NewRouter()
	cfg := apiConfig{
		db: database.NewDB(),
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

	admin.Get("/metrics", cfg.adminMetrics)

	r.Mount("/api", api)
	r.Mount("/admin", admin)

	server := &http.Server{
		Addr:    ":3000",
		Handler: middlewareCors(r),
	}

	log.Println("server starting")
	err := server.ListenAndServe()

	if err != nil {
		log.Printf("error: %v\n", err)
		log.Println("server closing")
	}
}
