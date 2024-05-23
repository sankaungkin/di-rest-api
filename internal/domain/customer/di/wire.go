//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/customer"
)

var CustomerWireSet = wire.NewSet(
	database.NewDB,
	customer.NewCustomerRepository,
	customer.NewCustomerService,
	customer.NewCustomerHandler,
)

func InitCustomer() (*customer.CustomerHandler, error) {
	wire.Build(CustomerWireSet)
	return &customer.CustomerHandler{}, nil
}
