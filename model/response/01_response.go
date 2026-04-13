package response

type Default struct {
	Result interface{} `json:"result,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
