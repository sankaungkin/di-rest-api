swagger:
	@swag init --parseDependency --parseInternal -g ./cmd/main.go -o ./cmd/docs

compile:
	@go build -o bin/min cmd/main.go

run:
	@go run cmd/main.go 

unitconversion:
	@wire ./internal/domain/unitconversion/di/wire.go

uom:
	@wire ./internal/domain/unitofmeasurement/di/wire.go

product:
	@wire ./internal/domain/product/di/wire.go

productprice:	
	@wire ./internal/domain/productprice/di/wire.go

productstock:
	@wire ./internal/domain/productstock/di/wire.go

itemtransactions:
	@wire ./internal/domain/itemtransactions/di/wire.go

customer:
	@wire ./internal/domain/customer/di/wire.go

supplier:
	@wire ./internal/domain/supplier/di/wire.go

category:
	@wire ./internal/domain/category/di/wire.go

inventory:
	@wire ./internal/domain/inventory/di/wire.go

sale:
	@wire ./internal/domain/sale/di/wire.go

purchase:
	@wire ./internal/domain/purchase/di/wire.go

auth:
	@wire ./internal/auth/di/wire.go


	