package models

type LoginResponse struct {
	User  UserDetails `json:"user"`
	Token string      `json:"token"`
}

type UserDetails struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
