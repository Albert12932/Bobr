package models

type AuthBookRequest struct {
	BookId int64 `json:"book_id"`
}
type AuthStatus struct {
	Status          string `json:"status"`
	DisplayName     string `json:"display_name"`
	StudentGroup    string `json:"student_group"`
	LinkToken       string `json:"link_token"`
	LinkTokenTtlSec int64  `json:"link_token_ttl_sec"`
}

type RegisterRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type RegisterResponse struct {
	UserSubstructure `json:"user"`
	AuthTokens       `json:"auth"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type LoginResponse struct {
	UserSubstructure `json:"user"`
	AuthTokens       `json:"auth"`
}

type GetTokensRequest struct {
	UserId    int64
	RoleLevel int64
}

type GetTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpUnix      int64  `json:"expires_at"`
}
type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpUnix      int64  `json:"expires_at"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RefreshTokenResponse struct {
	UserID     int64 `json:"user_id"`
	AuthTokens `json:"auth"`
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}
type SetNewPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type EmailAuth struct {
	EmailFrom string
	EmailPass string
}
