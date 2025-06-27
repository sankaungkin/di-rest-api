//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/productprice"
)

var ProductPriceWireSet = wire.NewSet(
	database.NewDB,
	productprice.NewProductPriceRepository,
	productprice.NewProductPriceService,
	productprice.NewProductPriceHandler,
)

func InitProductPriceDI() (*productprice.ProductPriceHandler, error) {
	wire.Build(ProductPriceWireSet)
	return &productprice.ProductPriceHandler{}, nil
}
