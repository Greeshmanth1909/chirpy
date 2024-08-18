package database

import (
    "os"
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

    return DB{path}
}

/* Read function reads and returns data from database.json as a struct */
func (db *DB) Read() Data {
    db.mu.Lock()
    defer dm.mu.Unlock()

    dat, err := os.ReadFile(db.path)
    decode := Data{}

    err := json.Unmarshal(dat, &decode)
    if err != nil {
        panic(err)
    }
    return decode
}

/* Write function writes/appends data to database.json file */
func (db *DB) Write(d Chirp) error {
    // read thi file first
    data := db.Read()

    numChirps := len(data.Chirps)
    id := numChirps + 1
    data.Chirps[id] = d
    
    // encode json
    dat, err := json.Marshal(data)
    if err != nil {
        panic(err)
    }

    // write dat to file
    db.mu.Lock()
    defer db.mu.Unlock()
    err := os.WriteFile(db.path, dat, 0666)
    if err != nil {
        panic(err)
    }
    return nil
}

/* createChirp converts a string into chirp */
func (db *DB) CreateChirp(chirp string) Chirp {
    db.mu.Lock()
    defer db.mu.Unlock()

    data := db.Read()
    chirpId := len(data.Chirps) + 1
    var newChirp Chirp
    newChirp.Id = chirpId
    newChirp.Body = chirp
    return newChirp
}
