//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	itemtransactions "github.com/sankangkin/di-rest-api/internal/domain/itemtransactions"
)

var TransactionWireSet = wire.NewSet(
	database.NewDB,
	itemtransactions.NewTransactionRepository,
	itemtransactions.NewTransactionService,
	itemtransactions.NewTransactionHandler,
)

func InitTransactionDI() (*itemtransactions.TransactionHandler, error) {
	wire.Build(TransactionWireSet)
	return &itemtransactions.TransactionHandler{}, nil
}
