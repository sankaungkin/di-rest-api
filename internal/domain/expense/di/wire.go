//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/expense"
)

var ExpenseWireSet = wire.NewSet(
	database.NewDB,
	expense.NewExpenseRepository,
	expense.NewExpenseService,
	expense.NewExpenseHandler,
)

func InitExpense() (*expense.ExpenseHandler, error) {
	wire.Build(ExpenseWireSet)
	return &expense.ExpenseHandler{}, nil
}
