package auth

import (
	"log"
	"strings"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	"gorm.io/gorm"
)

type AuthRepositoryInterface interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByName(name string) (*models.User, error)
}

type AuthRepository struct {
	db *gorm.DB
}

var (
	repoInstance *AuthRepository
	repoOnce     sync.Once
)

func NewAuthRepository(db *gorm.DB) AuthRepositoryInterface {
	log.Println(util.Red + "AuthRepository constructor is called" + util.Reset)
	repoOnce.Do(func() {
		repoInstance = &AuthRepository{db: db}
	})

	return repoInstance
}

func (r *AuthRepository) CreateUser(user *models.User) (*models.User, error) {

	newUser := models.User{
		Email:    user.Email,
		Password: user.Password,
		UserName: user.UserName,
		IsAdmin:  user.IsAdmin,
		Role:     user.Role,
	}
	result := r.db.Create(&newUser)
	if result.Error != nil {
		return nil, result.Error
	}
	return &newUser, nil
}

func (r *AuthRepository) GetUserByName(email string) (*models.User, error) {

	var user models.User
	err := r.db.Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, err
}
