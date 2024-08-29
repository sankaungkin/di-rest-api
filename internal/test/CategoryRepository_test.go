package test

import (
	"context"
	"fmt"
	"testing"

	c "github.com/sankangkin/di-rest-api/internal/domain/category"
	"github.com/sankangkin/di-rest-api/internal/models"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	p "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CategoryRepositoryTestSuite struct {
	suite.Suite
	db        *gorm.DB
	repo      c.CategoryRepositoryInterface
	container testcontainers.Container
}

func TestCategoryRepositoryTestSuite(t *testing.T) {
	suite.Run(t, &CategoryRepositoryTestSuite{})
}

func (suite *CategoryRepositoryTestSuite) SetupSuite() {

	suite.T().Log("---------SetupSuite()--------")
	// Set up the Testcontainers PostgreSQL container
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.NoError(err)
	suite.container = postgresContainer

	// Get the container's IP address and port
	host, err := postgresContainer.Host(ctx)
	suite.NoError(err)
	port, err := postgresContainer.MappedPort(ctx, "5432")
	suite.NoError(err)

	// Create the database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpassword dbname=testdb sslmode=disable", host, port.Port())

	// Connect to the database using GORM
	db, err := gorm.Open(p.Open(dsn), &gorm.Config{})
	suite.NoError(err)
	suite.NotNil(db) // Ensure db is not nil

	// Auto migrate the Category model
	err = db.AutoMigrate(&models.Category{})
	suite.NoError(err)

	// Assign the db and repo to the suite
	suite.db = db
	suite.repo = c.NewCategoryRepository(db)
	suite.NotNil(suite.repo) // Ensure repo is not nil
}

func (suite *CategoryRepositoryTestSuite) TearDownSuite() {
	suite.T().Log("----------TearDownSuite()----------")

	err := suite.container.Terminate(context.Background())
	suite.NoError(err)
}

func (suite *CategoryRepositoryTestSuite) TestGetAll() {

	suite.T().Log("----------TestGetAll()----------")

	for _, category := range MockCategories {
		err := suite.db.Create(&category).Error
		suite.NoError(err)
	}

	categories, err := suite.repo.GetAll()
	suite.NoError(err)
	suite.NotNil(categories)
	suite.Equal(6, len(categories))

	suite.T().Log(categories)
	// Check if "Category 2" is present in the list
	found := false
	for _, category := range MockCategories {
		if category.CategoryName == "Construction Materials" {
			found = true
			break
		}
	}
	suite.True(found, `"Construction Materials" should be present in the categories list`)

}

func (suite *CategoryRepositoryTestSuite) TestCreate() {

	suite.T().Log("-----------TestCreate()---------")

	newCategory := &models.Category{
		CategoryName: "Test Category",
	}

	suite.NotNil(newCategory)

	suite.Equal("Test Category", newCategory.CategoryName)

	if _, err := suite.repo.Create(newCategory); err != nil {
		suite.T().Fatalf("failed to create category: %v", err)
	}

	// suite.T().Log("new category has been created", newCategory)
	var count int64
	err := suite.db.Model(&models.Category{}).Where("category_name = ?", "Test Category").Count(&count).Error
	suite.NoError(err)
	suite.Equal(int64(1), count, "Category should be created in the database")
}

func (suite *CategoryRepositoryTestSuite) TestGetById() {
	suite.T().Log("------------TestGetById()-------------")

	category, err := suite.repo.GetById(5)
	suite.NoError(err)
	suite.NotNil(category)
	suite.Equal("PVC Fitting", category.CategoryName)
}

func (suite *CategoryRepositoryTestSuite) TestUpdate() {
	suite.T().Log("------------TestUpdate()--------------")

	category, err := suite.repo.GetById(5)
	suite.NoError(err)
	suite.NotNil(category)

	category.CategoryName = "Updated Category"
	updatedCategory, err := suite.repo.Update(category)
	suite.NoError(err)
	suite.NotNil(updatedCategory)
	suite.Equal("Updated Category", updatedCategory.CategoryName)
}

func (suite *CategoryRepositoryTestSuite) TestZDelete() {
	suite.T().Log("--------------TestDelete()------------")

	// delete the category with ID 2
	err := suite.repo.Delete(2)
	suite.NoError(err)

	//check category with ID 2 was still exited or not
	category, err := suite.repo.GetById(2)
	suite.Error(err)
	suite.Nil(category)
}
