package main

import (
    "net/http"
    "strconv"
    "fmt"
    "os"
    "github.com/joho/godotenv"
    "encoding/json"
    "strings"
    "golang.org/x/crypto/bcrypt"
    "github.com/Greeshmanth1909/chirpy/database"
)

func main() {
    // Load environment variables from dot file
    godotenv.Load()
    jwtSecret := os.Getenv("JWT_SECRET")
    hits.jwtSecret = jwtSecret

    serveMux := http.NewServeMux()
    var server http.Server
    server.Addr = "localhost:8080"
    server.Handler = serveMux


    // add handler to root
    serveMux.Handle("/app/", http.StripPrefix("/app/", hits.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
    serveMux.HandleFunc("GET /api/healthz", readyness)
    serveMux.HandleFunc("GET /admin/metrics", metricsHandler)
    serveMux.HandleFunc("/api/reset", resetHandler)
    serveMux.HandleFunc("POST /api/chirps", valChirpHandler)
    serveMux.HandleFunc("GET /api/chirps", getChirps)
    serveMux.HandleFunc("GET /api/chirps/{chirpid}", getChirpById)
    serveMux.HandleFunc("POST /api/users", postUsers)
    serveMux.HandleFunc("POST /api/login", loginUsers)
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
    jwtSecret string
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

var db = database.NewDB("/Users/geechu/Desktop/Programing/Projects/chirpy/database.json")

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

    chirp := db.CreateChirp(rmProfane(bodyStr))
    dat, err := json.Marshal(chirp)
    w.WriteHeader(201)
    w.Write([]byte(dat))
    db.Write(chirp)
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

/* getChirps function reads json from the database and returns it*/
func getChirps(w http.ResponseWriter, r *http.Request) {
    dat, err := db.Read()
    if err != nil {
        w.WriteHeader(200)
        w.Write([]byte("No chirps in the database yet"))
        return
    }
    var s []database.Chirp
    for _, val := range dat.Chirps {
        s = append(s, val)
    }
    data, err := json.Marshal(s)
    w.Write(data)
    return
}

/* getChirpById function returns the requested chirp with a status code of 200 */
func getChirpById(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.Atoi(r.PathValue("chirpid"))
    data, err := db.Read()
    if err != nil {
        fmt.Println("llll")
        return
    }
    if id > len(data.Chirps) {
        w.WriteHeader(404)
        return
    }
    id -= 1
    dat, _ := json.Marshal(data.Chirps[id])
    w.Write(dat)
}

type Email struct {
        Email string `json:"email"`
        Password string `json:"password"`
        Expires_in_seconds int `json:"expires_in_seconds"`
}

type User struct {
        Id int `json:"id"`
        Email string `json:"email"`
}

/* postUsers function adds a user to the database and responds with the username and its corresponding id */
func postUsers(w http.ResponseWriter, r *http.Request) {
    var email Email
    decoder := json.NewDecoder(r.Body)
    err := decoder.Decode(&email)
    if err != nil {
        fmt.Println("couldn't unmarshal request body")
    }

    id, e := db.CreateUser(email.Email, email.Password)
    user := User{id, e}
    dat, _ := json.Marshal(user)
    w.WriteHeader(201)
    w.Write(dat)
}

/* loginUsers logs a user in if they exist in the db and password mathes stored hash */
func loginUsers(w http.ResponseWriter, r *http.Request) {
    var req Email
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&req)

    dat, err := db.Read()
    if err != nil {
        w.WriteHeader(401)
        return
    }
    
    for _, val := range dat.Users {
        if val.Email == req.Email {
            err = bcrypt.CompareHashAndPassword(val.Hash, []byte(req.Password))
            if err != nil {
                w.WriteHeader(401)
                return
            } else {
                w.WriteHeader(200)
                user := User{val.Id, val.Email}
                dat, _ := json.Marshal(user)
                w.Write(dat)
                return
            }
        }
    }
    w.WriteHeader(401)
}
