# chirpy
A webserver built with Golang and a custom "database" that stores information in a `.json` file.

### Requirements
`go 1.22.2` or greater

### Installation and Setup

### Running the server
### Endpoints
(temporary referance only)
```
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
```

### Technicals
- jwt
- CRUD operations
