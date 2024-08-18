package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "strings"
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
    type getData struct {
        Body string `json:"body"`
    }
    type sendData struct {
        Error string `json:"error"`
        Valid bool `json:"valid"`
        Cleaned_body string `json:"cleaned_body"`
    }

    get := getData{}
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&get)
    
    if err != nil {
        fmt.Println("error")
    }

    bodyStr := get.Body
    if len(bodyStr) > 140 {
       // encode response body
       response := sendData{"Chirp is too long", false, ""}
       dat, err := json.Marshal(response)

       if err != nil {
            fmt.Println("error encoding json")
            return
       }
       w.WriteHeader(400)
       w.Write([]byte(dat))
       return
    } 

    response := sendData{}
    response.Valid = true
    response.Cleaned_body = rmProfane(bodyStr)

    dat, err := json.Marshal(response)
    if err != nil {
        fmt.Println("error encoding json 2")
        return
    }
    w.Write([]byte(dat))
    return
}

// rmProfane replaces certain words with ****
func rmProfane(opinion string) string {
    forbiddenWords := []string{"kerfuffle", "sharbert", "fornax"}
    lowerCaseOpinion := strings.ToLower(opinion)
    opinionSplit := strings.Split(opinion, " ")

    for _, val := range forbiddenWords {
        splitstr := strings.Split(lowerCaseOpinion, " ")
        for i, k := range splitstr {
            if k == val {
                opinionSplit[i] = "****"
            }
        }
    }
    return strings.Join(opinionSplit, " ")
}
