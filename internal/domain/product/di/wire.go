//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/product"
)

var ProductWireSet = wire.NewSet(
	database.NewDB,
	product.NewProductRepository,
	product.NewProductService,
	product.NewProductHandler,
)

func InitProductDI() (*product.ProductHandler,error) {
	wire.Build(ProductWireSet)
	return &product.ProductHandler{}, nil
}