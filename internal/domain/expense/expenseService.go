package expense

import (
	"log"
	"sync"

	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
)

type ExpenseServiceInterface interface {
	CreateExpenseService(expense *models.Expense) (*models.Expense, error)
	GetAllExpensesService() ([]models.Expense, error)
}

type ExpenseService struct {
	repo ExpenseRepositoryInterface
}

var (
	svcInstance *ExpenseService
	svcOnce     sync.Once
)

func NewExpenseService(repo ExpenseRepositoryInterface) ExpenseServiceInterface {
	log.Println(util.Cyan + "ExpenseService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &ExpenseService{repo: repo}
	})
	return svcInstance
}
func (s *ExpenseService) CreateExpenseService(expense *models.Expense) (*models.Expense, error) {
	return s.repo.Create(expense)
}

func (s *ExpenseService) GetAllExpensesService() ([]models.Expense, error) {
	return s.repo.Getall()
}
