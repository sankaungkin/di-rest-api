// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/inventory"
)

// Injectors from wire.go:

func InitInventoryDI() (*inventory.InventoryHandler, error) {
	db, err := database.NewDB()
	if err != nil {
		return nil, err
	}
	inventoryRepositoryInterface := inventory.NewInventoryRepository(db)
	inventoryServiceInterface := inventory.NewInventoryService(inventoryRepositoryInterface)
	inventoryHandler := inventory.NewInventoryHandler(inventoryServiceInterface)
	return inventoryHandler, nil
}

// wire.go:

var InventoryWireSet = wire.NewSet(database.NewDB, inventory.NewInventoryRepository, inventory.NewInventoryService, inventory.NewInventoryHandler)
