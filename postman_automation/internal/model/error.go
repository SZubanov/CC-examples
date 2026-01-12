package model

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func NewErrorResponse(code int, err, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
		Code:    code,
	}
}
