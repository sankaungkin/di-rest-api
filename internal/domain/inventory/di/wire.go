//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/inventory"
)

var InventoryWireSet = wire.NewSet(
	database.NewDB,
	inventory.NewInventoryRepository,
	inventory.NewInventoryService,
	inventory.NewInventoryHandler,
)

func InitInventoryDI() (*inventory.InventoryHandler, error) {
	wire.Build(InventoryWireSet)
	return &inventory.InventoryHandler{}, nil
}