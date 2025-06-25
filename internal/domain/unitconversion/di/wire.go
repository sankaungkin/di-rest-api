//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/unitconversion"
)

var UnitConversionWireSet = wire.NewSet(
	database.NewDB,
	unitconversion.NewUnitConversionRepository,
	unitconversion.NewUnitConversionService,
	unitconversion.NewUnitConversionHandler,
)

func InitUnitConversionDI() (*unitconversion.UnitConversionHandler, error) {
	wire.Build(UnitConversionWireSet)
	return &unitconversion.UnitConversionHandler{}, nil
}
