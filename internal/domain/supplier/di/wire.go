//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/supplier"
)

var SupplierWireSet = wire.NewSet(
	database.NewDB,
	supplier.NewSupplierRepository,
	supplier.NewSupplierService,
	supplier.NewSupplierHandler,
)

func InitSupplier() (*supplier.SupplierHandler, error) {
	wire.Build(SupplierWireSet)
	return &supplier.SupplierHandler{}, nil
}
