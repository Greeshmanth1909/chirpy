package main

import (
    "net/http"
    "fmt"
)

func main() {
    serveMux := http.NewServeMux()
    var server http.Server
    server.Addr = "localhost:8080"
    server.Handler = serveMux


    // add handler to root
    serveMux.Handle("/app/", http.StripPrefix("/app/", hits.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
    serveMux.HandleFunc("GET /api/healthz", readyness)
    serveMux.HandleFunc("GET /admin/metrics", metricsHandler)
    serveMux.HandleFunc("/api/reset", resetHandler)
    serveMux.HandleFunc("POST /api/validate_chirp", valChirpHandler)

    server.ListenAndServe()
}

func readyness(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

// Type to assess metrics
type apiConfig struct {
	fileserverHits int
}

var hits apiConfig

/* middlewareMetricsInc is a middle ware function that runs before the root handler is called, it counts the number of times the
site was visited */
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits += 1
        next.ServeHTTP(w, r)
    })
}

/* metricsHandler writes the fileserverHits from apiConfig to the response body*/
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %v times!</p></body></html>", hits.fileserverHits)))
}

/* Reset function handler resets the hits count to zero*/
func resetHandler(w http.ResponseWriter, r *http.Request) {
    hits.fileserverHits = 0
    w.WriteHeader(http.StatusOK)
}

/* valChirpHandler validates the post request from the user */
func valChirpHandler(w http.ResponseWriter, r *http.Request) {
    return
}
