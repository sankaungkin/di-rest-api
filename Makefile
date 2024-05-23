compile:
	@go build -o bin/min cmd/main.go

run:
	@go run cmd/main.go

product:
	@wire ./internal/domain/product/di/wire.go

customer:
	@wire ./internal/domain/customer/di/wire.go

category:
	@wire ./internal/domain/category/di/wire.go
	