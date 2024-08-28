package main

type IndexRequest struct {
	Say string `json:"say"`
}

type IndexResponse struct {
	Back string `json:"back"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
