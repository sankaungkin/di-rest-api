// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/itemtransactions"
)

// Injectors from wire.go:

func InitTransactionDI() (*itemtransactions.TransactionHandler, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}
	transactionRepositoryInterface := itemtransactions.NewTransactionRepository(db)
	transactionServiceInterface := itemtransactions.NewTransactionService(transactionRepositoryInterface)
	transactionHandler := itemtransactions.NewTransactionHandler(transactionServiceInterface)
	return transactionHandler, nil
}

// wire.go:

var TransactionWireSet = wire.NewSet(database.NewDB, itemtransactions.NewTransactionRepository, itemtransactions.NewTransactionService, itemtransactions.NewTransactionHandler)
