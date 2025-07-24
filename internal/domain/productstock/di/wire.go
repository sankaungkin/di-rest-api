//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/productstock"
)

var ProductStockWireSet = wire.NewSet(
	database.NewDB,
	productstock.NewProductStockRepository,
	productstock.NewProductStockService,
	productstock.NewProductStockHandler,
)

var (
	ProductStockRepoSet = wire.NewSet(
		database.NewDB,
		productstock.NewProductStockRepository,
	)
	ProductStockServiceSet = wire.NewSet(
		database.NewDB,
		productstock.NewProductStockService,
		productstock.NewProductStockHandler,
	)
)

// For HTTP Handler
func InitProductStockDI() (*productstock.ProductStockHandler, error) {
	wire.Build(ProductStockWireSet)
	return &productstock.ProductStockHandler{}, nil
}

// For WebSocket Handler
func InitProductStockRepoDI() (productstock.ProductStockRepositoryInterface, error) {
	wire.Build(ProductStockRepoSet)
	return &productstock.ProductStockRepository{}, nil
}
