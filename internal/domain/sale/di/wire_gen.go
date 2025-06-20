// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/sale"
)

// Injectors from wire.go:

func InitSaleDI() (*sale.SaleHandler, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}
	saleRepositoryInterface := sale.NewSaleRepository(db)
	saleServiceInterface := sale.NewSaleService(saleRepositoryInterface)
	saleHandler := sale.NewSaleHandler(saleServiceInterface)
	return saleHandler, nil
}

// wire.go:

var SaleWireSet = wire.NewSet(database.NewDB, sale.NewSaleRepository, sale.NewSaleService, sale.NewSaleHandler)
