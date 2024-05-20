package main

import (
	"github.com/google/wire"
	"github.com/sankangkin/di-rest-api/internal/database"
	"github.com/sankangkin/di-rest-api/internal/handler"
	"github.com/sankangkin/di-rest-api/internal/repository"
	"github.com/sankangkin/di-rest-api/internal/service"
)

//TODO : cannot generate the code from wire

var CategoryWireSet = wire.NewSet(
	database.NewDB, 
	repository.NewCategoryRepository, 
	service.NewCategoryService, 
	handler.NewCategoryHandler,
	
)

func InitCategory() (*handler.CategoryHandler, error) {
	wire.Build(CategoryWireSet)
	return &handler.CategoryHandler{}, nil
}