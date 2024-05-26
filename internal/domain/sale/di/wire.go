//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/sale"
)

var SaleWireSet = wire.NewSet(
	database.NewDB,
	sale.NewSaleRepository,
	sale.NewSaleService,
	sale.NewSaleHandler,
)

func InitSaleDI() (*sale.SaleHandler, error) {
	wire.Build(SaleWireSet)
	return &sale.SaleHandler{},nil
}