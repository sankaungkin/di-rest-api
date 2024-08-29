package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/sankangkin/di-rest-api/internal/domain/category"
	"github.com/sankangkin/di-rest-api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	p "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var categories = []models.Category{
	{
		CategoryName: "Construction Materials",
	},
	{
		CategoryName: "Sanitary Ware",
	},
	{
		CategoryName: "PVC Pipe",
	},
	{
		CategoryName: "PVC Fitting",
	},
	{
		CategoryName: "GI Fitting",
	},
}

func SetUpDBContainer(ctx context.Context) (testcontainers.Container, *gorm.DB, error) {
	dbName := "testPosdb"
	dbUser := "user"
	dbPassword := "passw0rd"

	req := testcontainers.ContainerRequest(testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     dbUser,
			"POSTGRES_PASSWORD": dbPassword,
			"POSTGRES_DB":       dbName,
		},

		WaitingFor: wait.ForListeningPort("5432/tcp"),
	})
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, nil, err
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return nil, nil, err
	}

	connURL := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=disable", port.Port(), dbUser, dbPassword, dbName)
	db, err := gorm.Open(p.Open(connURL), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, nil, err
	}

	// fmt.Println("Seeding categories data ....")
	// db.Create(&categories)

	return pgContainer, db, nil

}

func TestCategoryRepo_Create(t *testing.T) {
	ctx := context.Background()
	container, gormDB, err := SetUpDBContainer(ctx)
	if err != nil {
		t.Fatalf("could not start container: %v", err)
	}
	assert.NoError(t, err)

	defer func() {
		err := container.Terminate(ctx)
		if err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	}()

	err = gormDB.AutoMigrate(&models.Category{})
	assert.NoError(t, err)
	catRepo := category.NewCategoryRepository(gormDB)

	newCategory := &models.Category{
		CategoryName: "Test Category",
	}

	assert.NoError(t, err)
	assert.NotNil(t, newCategory)

	assert.Equal(t, "Test Category", newCategory.CategoryName)

	if _, err := catRepo.Create(newCategory); err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	assert.NoError(t, err)
	t.Log("new category has been created", newCategory)
	var count int64
	err = gormDB.Model(&models.Category{}).Where("category_name = ?", "Test Category").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count, "Category should be created in the database")

	for _, category := range categories {
		err = gormDB.Create(&category).Error
		assert.NoError(t, err)
	}

	categories, err := catRepo.GetAll()
	assert.NoError(t, err)
	assert.NotNil(t, categories)
	assert.Equal(t, 6, len(categories))
	assert.Equal(t, "Test Category", categories[0].CategoryName)
	assert.Equal(t, "Construction Materials", categories[1].CategoryName)

	// Check if "Category 2" is present in the list
	found := false
	for _, category := range categories {
		if category.CategoryName == "Test Category" {
			found = true
			break
		}
	}
	assert.True(t, found, `"Construction Materials" should be present in the categories list`)

	t.Cleanup(func() {
		gormDB.Migrator().DropTable(&models.Category{})
	})
}
