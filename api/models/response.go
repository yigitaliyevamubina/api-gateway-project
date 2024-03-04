package models

type ResponseError struct {
	Error interface{} `json:"error"`
}

type ServerError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ValidationError struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	UserMessage string `json:"user_message"`
}

type StandardErrorModel struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
