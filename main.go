package main

import (
    "net/http"
    "strconv"
    "fmt"
    "os"
    "encoding/json"
    "encoding/hex"
    "strings"
    "time"
    "crypto/rand"
    "golang.org/x/crypto/bcrypt"
    "github.com/Greeshmanth1909/chirpy/database"
    "github.com/golang-jwt/jwt/v5"
    "github.com/joho/godotenv"
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
    serveMux.HandleFunc("PUT /api/users", updateUsers)
    serveMux.HandleFunc("POST /api/login", loginUsers)
    serveMux.HandleFunc("POST /api/refresh", refresh)
    serveMux.HandleFunc("POST /api/revoke", revoke)
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
        Token string `json:"token"`
        Refresh_token string `json:"refresh_token"`
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
    user := User{id, e, "", ""}
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
                // Generate refresh token
                c := 32
                b := make([]byte, c)
                rand.Read(b)
                ref_token := hex.EncodeToString(b)
                w.WriteHeader(200)
                var user User
                user.Id = val.Id
                user.Email = val.Email
                user.Refresh_token = ref_token
                var claims jwt.RegisteredClaims
                claims.Issuer = "chirpy"
                claims.IssuedAt = jwt.NewNumericDate(time.Now())
                if req.Expires_in_seconds != 0 {
                    claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(req.Expires_in_seconds) * time.Second))
                }
                // add refresh token to the database
                db.AddRefToken(val.Id, ref_token)
                claims.Subject = string(val.Id)
                token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
                jwt, _ := token.SignedString([]byte(hits.jwtSecret))
                user.Token = jwt
                dat, _ := json.Marshal(user)
                w.Write(dat)
                return
            }
        }
    }
    w.WriteHeader(401)
}

/* updateUsers function verifies the jwt and updates a user's email */
func updateUsers(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        w.WriteHeader(404)
        return
    }
    token = strings.TrimPrefix(token, "Bearer ")
    var claims jwt.RegisteredClaims
    _ , err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(hits.jwtSecret), nil
    })
    if err != nil {
        w.WriteHeader(401)
        return
    }

    // read and update the database
    i, _ := claims.GetSubject()
    id, _ := strconv.Atoi(i)
    // get fields that need to be updated
    var user Email
    decoder := json.NewDecoder(r.Body)
    decoder.Decode(&user)
    // user now has email and password
    // Update username
    db.UpdateUser(user.Email, user.Password, id)
    var updatedUser User
    updatedUser.Id = id + 1
    updatedUser.Email = user.Email

    data, _ := json.Marshal(updatedUser)
    w.Write(data)
}

/* refresh function verifies the refresh token and generates a new jwt */
func refresh(w http.ResponseWriter, r *http.Request) {
    auth := r.Header.Get("Authorization")
    auth = strings.TrimPrefix(auth, "Bearer ")
    dat, _ := db.Read()
    for _, val := range dat.Users {
        if val.Refresh_token == auth {
            // Generate new jwt
                w.WriteHeader(200)
                var user User
                user.Id = val.Id
                user.Email = val.Email
                user.Refresh_token = auth
                var claims jwt.RegisteredClaims
                claims.Issuer = "chirpy"
                claims.IssuedAt = jwt.NewNumericDate(time.Now())
                claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(60) * time.Minute))
                // add refresh token to the database
                claims.Subject = string(val.Id)
                token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
                jwt, _ := token.SignedString([]byte(hits.jwtSecret))
                user.Token = jwt
                dat, _ := json.Marshal(user)
                w.Write(dat)
                return
        }
    }
    w.WriteHeader(401)
}

/* revoke function removes the token */
func revoke(w http.ResponseWriter, r *http.Request) {
    auth := r.Header.Get("Authorization")
    auth = strings.TrimPrefix(auth, "Bearer ")
    
    var user database.User
    dat, _ := db.Read()
    for _, val := range dat.Users {
        if val.Refresh_token == auth {
            user.Id = val.Id
            user.Email = val.Email
            user.Hash = val.Hash
            user.Refresh_token = ""
        }
    }
    mapId := user.Id - 1
    dat.Users[mapId] = user
    db.WriteDb(dat)
    w.WriteHeader(204)
}
