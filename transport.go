package main

// Requests and Responses

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V string `json:"v"`
	Err string `json:"err,omitempty"`
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int64 `json:"v"`
}

type healthRequest struct {}

type healthResponse struct {
	S bool `json:"status"`
}

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token,omitempty"`
	Err   string `json:"err,omitempty"`
}
