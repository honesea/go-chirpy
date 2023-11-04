package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	admin := chi.NewRouter()
	api := chi.NewRouter()
	cfg := apiConfig{}

	fileServer := cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))

	r.Handle("/app", http.StripPrefix("/app", fileServer))
	r.Handle("/app/*", http.StripPrefix("/app", fileServer))

	api.Get("/healthz", healthz)
	api.Get("/metrics", cfg.metrics)
	api.Get("/reset", cfg.reset)
	api.Post("/validate_chirp", validateChrip)

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
