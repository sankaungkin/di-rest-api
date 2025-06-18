package auth

import "github.com/sankangkin/di-rest-api/internal/models"

type SignUpDTO struct {
	UserName string `json:"userName" validate:"required, min=3"`
	Email    string `json:"email" validate:"required" gorm:"unique"`
	Password string `json:"password" validate:"required, min=3"`
	Role     string `json:"role" validate:"required"`
}

type SignUpResponseDTO struct {
	User models.User
	At   string `json:"AccessToken"`
	Rt   string `json:"RefreshToken"`
}

type SignInRequestDTO struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type SignInResponseDTO struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserName     string `json:"userName"`
	Role         string `json:"role"`
}

type RefreshResponseDTO struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshRequestDTO struct {
	RefreshToken string `json:"refreshToken"`
}
