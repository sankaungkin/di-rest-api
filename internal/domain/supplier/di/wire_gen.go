// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/supplier"
)

// Injectors from wire.go:

func InitSupplier() (*supplier.SupplierHandler, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}
	supplierRepositoryInterface := supplier.NewSupplierRepository(db)
	supplierServiceInterface := supplier.NewSupplierService(supplierRepositoryInterface)
	supplierHandler := supplier.NewSupplierHandler(supplierServiceInterface)
	return supplierHandler, nil
}

// wire.go:

var SupplierWireSet = wire.NewSet(database.NewDB, supplier.NewSupplierRepository, supplier.NewSupplierService, supplier.NewSupplierHandler)
