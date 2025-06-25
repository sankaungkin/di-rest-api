//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/unitofmeasurement"
)

var UnitOfMeasurementWireSet = wire.NewSet(
	database.NewDB,
	unitofmeasurement.NewUnitOfMeasurementRepository,
	unitofmeasurement.NewUnitOfMeasurementService,
	unitofmeasurement.NewUnitOfMeasurementHandler,
)

func InitUnitOfMeasurementDI() (*unitofmeasurement.UnitOfMeasurementHandler, error) {
	wire.Build(UnitOfMeasurementWireSet)
	return &unitofmeasurement.UnitOfMeasurementHandler{}, nil
}
