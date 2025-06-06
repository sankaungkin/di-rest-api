package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/sankangkin/di-rest-api/internal/domain/util"
	"github.com/sankangkin/di-rest-api/internal/models"
	mylog "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	Signup(user *models.User) (*models.User, error)
	Signin(username, password string) (string, string, string, string, error)
	FindUserByEmail(email string) (*models.User, error)
	Refresh(refreshToken string) (string, error)
	Signout(accessToken string) error
}

type AuthService struct {
	repo AuthRepositoryInterface
}

var (
	svcInstance *AuthService
	svcOnce     sync.Once
)

func init() {
	mylog.SetReportCaller(true)
	Formatter := new(mylog.JSONFormatter)
	Formatter.TimestampFormat = "2006-01-02 15:04:05"
	mylog.SetFormatter(Formatter)
}

func NewAuthService(repo AuthRepositoryInterface) AuthServiceInterface {
	mylog.Info("AuthService is called.")
	// log.Println(util.Red + "AuthService constructor is called" + util.Reset)

	svcOnce.Do(func() {
		svcInstance = &AuthService{repo: repo}
	})

	return svcInstance
}

func (s *AuthService) Signup(user *models.User) (*models.User, error) {

	password := hashAndSalt([]byte(user.Password))
	newUser := models.User{
		Email:    user.Email,
		UserName: user.UserName,
		Password: password,
		IsAdmin:  user.IsAdmin,
		Role:     user.Role,
	}

	result, err := s.repo.CreateUser(&newUser)
	if err != nil {
		return nil, err
	}

	return result, nil
	//   return nil,  "", "", nil
}

func (s *AuthService) FindUserByEmail(email string) (*models.User, error) {
	return s.repo.GetUserByName(email)
}

func (s *AuthService) Signin(email, password string) (string, string, string, string, error) {

	found, err := s.repo.GetUserByName(email)

	if err != nil {
		return "", "", "", "", err
	}

	if !comparePasswords(found.Password, []byte(password)) {
		return "", "", "", "", errors.New("invalid credentials")
	}

	at, rt, userName, role, err := generateTokens(found, util.SecreteKey)
	if err != nil {
		return "", "", "", "", errors.New("authentication faileds")
	}
	return at, rt, userName, role, nil
}

func (s *AuthService) Refresh(refreshToken string) (string, error) {

	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(refreshToken, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(util.SecreteKey), nil
	})
	for key, val := range claims {
		fmt.Printf("key: %v, value: %v\n", key, val)
	}
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		query := models.User{Email: claims["email"].(string)}
		found, err := s.repo.GetUserByName(query.Email)
		if err != nil {
			return "", err
		}
		rtClaims := jwt.MapClaims{
			"id":    found.ID,
			"email": found.Email,
			"admin": found.IsAdmin,
			"role":  found.Role,
			"exp":   time.Now().Add(time.Hour * 1).Unix(),
		}
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
		rt, err := refreshToken.SignedString([]byte(util.SecreteKey))
		if err != nil {
			return "", err
		}
		return rt, nil
	}
	return "", err

}

func (s *AuthService) Signout(accessToken string) error {
	return nil
}

func hashAndSalt(pwd []byte) string {
	hash, _ := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	return string(hash)
}

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	return err == nil
}

func SessionExpires() time.Time {
	return time.Now().Add(5 * 24 * time.Hour)
}

func generateTokens(user *models.User, secretKey string) (string, string, string, string, error) {
	// Define signing method and create claims
	claims := &jwt.MapClaims{
		"id":       user.ID,
		"email":    user.Email,
		"userName": user.UserName,
		"admin":    true,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Minute * 3).Unix(),
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString([]byte(secretKey))
	if err != nil {
		return "", "", "", "", err
	}

	// Define refresh token claims with longer expiry
	refreshTokenClaims := &jwt.MapClaims{

		"id":       user.ID,
		"email":    user.Email,
		"userName": user.UserName,
		"admin":    true,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 3).Unix(),
		//   "user_id": userID,
		//   "exp":    time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secretKey))
	userName := user.UserName
	role := string(user.Role)
	if err != nil {
		return "", "", "", "", err
	}
	return accessTokenString, refreshTokenString, userName, role, nil
}
