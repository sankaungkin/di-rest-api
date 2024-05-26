//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/purchase"
)

var PurchaseWireSet = wire.NewSet(
	database.NewDB,
	purchase.NewSaleRepository,
	purchase.NewSaleService,
	purchase.NewSaleHandler,
)

func InitPurchaseDI() (*purchase.PurchaseHandler, error){
	wire.Build(PurchaseWireSet)
	return &purchase.PurchaseHandler{}, nil
}