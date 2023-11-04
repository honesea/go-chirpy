package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) adminMetrics(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
		<html>

		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>

		</html>
	`, cfg.fileserverHits)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(html))
}
