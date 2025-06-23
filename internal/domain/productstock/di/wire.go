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

func InitProductStockDI() (*productstock.ProductStockHandler, error) {
	wire.Build(ProductStockWireSet)
	return &productstock.ProductStockHandler{}, nil
}
