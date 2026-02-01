//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/cashbook"
	"github.com/sankangkin/di-rest-api/internal/domain/sale"
)

// var SaleWireSet = wire.NewSet(
// 	database.NewDB,
// 	sale.NewSaleRepository,
// 	sale.NewSaleService,
// 	sale.NewSaleHandler,
// )

var SaleWireSet = wire.NewSet(
	database.NewDB,
	// 1. Add the Cashbook repository constructor
	cashbook.NewCashbookRepository,

	// 2. Bind the concrete struct to the interface so Wire knows they are the same
	// Replace *cashbook.CashbookRepository with whatever your actual struct name is
	// wire.Bind(new(cashbook.CashbookRepositoryInterface), new(*cashbook.CashbookRepository)),
	// cashbook.NewCashbookRepository,
	sale.NewSaleRepository,
	sale.NewSaleService,
	sale.NewSaleHandler,
)

func InitSaleDI() (*sale.SaleHandler, error) {
	wire.Build(SaleWireSet)
	return &sale.SaleHandler{}, nil
}
