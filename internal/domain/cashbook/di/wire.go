//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/cashbook"
)

var CashbookWireSet = wire.NewSet(
	database.NewDB,
	cashbook.NewCashbookRepository,
	cashbook.NewCashbookService,
	cashbook.NewCashbookHandler,
)

func InitCashbook() (*cashbook.CashbookHandler, error) {
	wire.Build(CashbookWireSet)
	return &cashbook.CashbookHandler{}, nil
}
