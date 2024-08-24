package database

import (
    "os"
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "encoding/json"
    "sync"
)

type DB struct {
    path string
    mu sync.Mutex
}

// Struct to handle dbio
type Data struct {
    Chirps map[int]Chirp `json:"chirps"`
    Users map[int]User `json:"user"`
}

type User struct {
    Id int `json:"id"`
    Email string `json:"email"`
    Hash []byte `json:"hash"`
}

type Chirp struct {
    Id int `json:"id"`
    Body string `json:"body"`
}

// NewDB creates a new database if it doesn't exist
func NewDB(path string) DB {
    _, err := os.ReadFile(path)
    if err == os.ErrNotExist {
        // File does not exist, create one
        _, err := os.Create(path)
        if err != nil {
            panic(err)
        }
    }
    var mu sync.Mutex
    return DB{path, mu}
}

/* Read function reads and returns data from database.json as a struct */
func (db *DB) Read() (Data, error) {
    db.mu.Lock()
    defer db.mu.Unlock()

    dat, err := os.ReadFile(db.path)
    decode := Data{}

    err = json.Unmarshal(dat, &decode)
    if err != nil {
        return decode, err
    }
    return decode, err
}

/* Write function writes/appends data to database.json file */
func (db *DB) Write(d Chirp) error {
    // read thi file first
    data, error := db.Read()
    if error != nil {
        // file is empty, create new type
        data = Data{make(map[int]Chirp), make(map[int]User)}
    }

    numChirps := len(data.Chirps)
    id := numChirps
    if numChirps == 0 {
        id = 0
    }
    data.Chirps[id] = d
    
    // encode json
    dat, err := json.Marshal(data)
    if err != nil {
        panic(err)
    }

    // write dat to file
    db.mu.Lock()
    defer db.mu.Unlock()
    err = os.WriteFile(db.path, dat, 0666)
    if err != nil {
        panic(err)
    }
    return nil
}

/* createChirp converts a string into chirp */
func (db *DB) CreateChirp(chirp string) Chirp {
    data, err := db.Read()
    if err != nil {
        c := Chirp{1, chirp}
        return c
    }
    chirpId := len(data.Chirps) + 1
    var newChirp Chirp
    newChirp.Id = chirpId
    newChirp.Body = chirp
    fmt.Println(newChirp)
    return newChirp
}

/* the CreateUser function creates and stores the user with an id and the hashed password. Returns user id and username */
func (db *DB) CreateUser(email, password string) (id int, username string){
    dat, err := db.Read()
    if err != nil {
        fmt.Println("Something up wit db")
        dat = Data{make(map[int]Chirp), make(map[int]User)}
    }
    
    currentLength := len(dat.Users)
    if currentLength == 0 {
        currentLength = 1
    }
    mapId := currentLength - 1
    userId := currentLength
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
    newUser := User{userId, email, hash}

    dat.Users[mapId] = newUser
    db.mu.Lock()
    defer db.mu.Unlock()
    str, error := json.Marshal(dat)
    if error != nil {
        fmt.Println("couldn't convert json")
        return 0, ""
    }
    os.WriteFile(db.path, str, 0666)

    return newUser.Id, newUser.Email
}
