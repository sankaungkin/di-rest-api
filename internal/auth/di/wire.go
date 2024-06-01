//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/auth"
	"github.com/sankangkin/di-rest-api/internal/database"
)

var AuthWireSet = wire.NewSet(
	database.NewDB,
	auth.NewAuthRepository,
	auth.NewAuthService,
	auth.NewAuthHandler,
)

func InitAuth() (*auth.AuthHandler, error) {
	wire.Build(AuthWireSet)
	return &auth.AuthHandler{}, nil
}