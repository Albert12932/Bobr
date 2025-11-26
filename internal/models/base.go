package models

type Payload struct {
	Sub       int64 `json:"sub"`
	RoleLevel int64 `json:"role_level"`
	Exp       int64 `json:"exp"`
	Iat       int64 `json:"iat"`
}
type ErrorResponse struct {
	Error   string
	Message string
}
type SuccessResponse struct {
	Successful bool   `json:"successful"`
	Message    string `json:"message"`
}
