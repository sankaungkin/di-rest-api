//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/domain/category"
)

//TODO : cannot generate the code from wire

var CategoryWireSet = wire.NewSet(
	database.NewDB, 	
	category.NewCategoryRepository, 
	category.NewCategoryService, 
	category.NewCategoryHandler,	
)

func InitCategory() (*category.CategoryHandler, error) {
	wire.Build(CategoryWireSet)
	return &category.CategoryHandler{}, nil
}