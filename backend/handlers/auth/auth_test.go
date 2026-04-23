package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/yourusername/kor-assetforge/backend/models"
	"github.com/yourusername/kor-assetforge/backend/utils"
)

type AuthTestSuite struct {
	suite.Suite
	db      *gorm.DB
	handler *AuthHandler
	router  *gin.Engine
	config  *AuthConfig
}

func (suite *AuthTestSuite) SetupTest() {
	// Setup in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate models
	err = db.AutoMigrate(&models.User{}, &models.UserSession{})
	suite.Require().NoError(err)

	// Setup auth config
	config := &AuthConfig{
		JWTSecret:           "test-secret-key",
		JWTExpirationHours:  24,
		RefreshTokenHours:   168,
		EmailTokenHours:     24,
		PasswordResetHours:  1,
		BcryptCost:          4, // Lower cost for tests
	}

	// Setup cache (mock)
	cache := utils.NewCache(nil) // nil redis client for tests

	// Setup handler
	handler := NewAuthHandler(db, config, cache)

	// Setup router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := NewAuthMiddleware(config.JWTSecret)

	// Setup routes
	v1 := router.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", handler.Register)
			authGroup.POST("/login", handler.Login)
			authGroup.POST("/refresh", handler.RefreshToken)
			authGroup.POST("/verify-email", handler.VerifyEmail)
			authGroup.POST("/forgot-password", handler.ForgotPassword)
			authGroup.POST("/reset-password", handler.ResetPassword)
		}

		protected := v1.Group("")
		protected.Use(authMiddleware.JWTAuth())
		{
			protected.GET("/profile", handler.GetProfile)
			protected.POST("/logout", handler.Logout)
		}
	}

	suite.db = db
	suite.handler = handler
	suite.router = router
	suite.config = config
}

func (suite *AuthTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *AuthTestSuite) TestRegister() {
	// Test successful registration
	req := RegisterRequest{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "test@example.com",
		Username:       "testuser",
		Password:       "password123",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "user_id")
}

func (suite *AuthTestSuite) TestRegisterDuplicateEmail() {
	// Create existing user
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "existing@example.com",
		Username:       "existinguser",
		PasswordHash:   "hashedpassword",
		EmailVerified:  true,
	}
	suite.db.Create(&user)

	// Try to register with same email
	req := RegisterRequest{
		StellarAddress: "GB7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "existing@example.com",
		Username:       "newuser",
		Password:       "password123",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

func (suite *AuthTestSuite) TestLogin() {
	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), suite.config.BcryptCost)
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "login@example.com",
		Username:       "loginuser",
		PasswordHash:   string(hashedPassword),
		EmailVerified:  true,
	}
	suite.db.Create(&user)

	// Test login
	req := LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response TokenResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(suite.T(), response.AccessToken)
	assert.NotEmpty(suite.T(), response.RefreshToken)
	assert.Equal(suite.T(), "Bearer", response.TokenType)
	assert.NotZero(suite.T(), response.ExpiresIn)
	assert.Equal(suite.T(), "loginuser", response.User.Username)
}

func (suite *AuthTestSuite) TestLoginUnverifiedEmail() {
	// Create test user with unverified email
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), suite.config.BcryptCost)
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "unverified@example.com",
		Username:       "unverifieduser",
		PasswordHash:   string(hashedPassword),
		EmailVerified:  false,
	}
	suite.db.Create(&user)

	// Test login
	req := LoginRequest{
		Email:    "unverified@example.com",
		Password: "password123",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *AuthTestSuite) TestGetProfile() {
	// Create test user
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "profile@example.com",
		Username:       "profileuser",
		PasswordHash:   "hashedpassword",
		Role:           models.RoleUser,
		EmailVerified:  true,
		KYCVerified:    true,
	}
	suite.db.Create(&user)

	// Generate token
	token, _, _ := suite.handler.generateTokens(&user)

	// Test get profile
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/profile", nil)
	r.Header.Set("Authorization", "Bearer "+token)

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response UserInfo
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(suite.T(), "profileuser", response.Username)
	assert.Equal(suite.T(), "user", response.Role)
	assert.True(suite.T(), response.EmailVerified)
	assert.True(suite.T(), response.KYCVerified)
}

func (suite *AuthTestSuite) TestVerifyEmail() {
	// Create test user with email token
	emailToken := generateSecureToken()
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "verify@example.com",
		Username:       "verifyuser",
		PasswordHash:   "hashedpassword",
		EmailVerified:  false,
		EmailToken:     emailToken,
		EmailTokenExpires: time.Now().Add(time.Hour),
	}
	suite.db.Create(&user)

	// Test email verification
	req := VerifyEmailRequest{
		Token: emailToken,
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/verify-email", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Check if user was updated
	var updatedUser models.User
	suite.db.First(&updatedUser, user.ID)
	assert.True(suite.T(), updatedUser.EmailVerified)
	assert.Empty(suite.T(), updatedUser.EmailToken)
}

func (suite *AuthTestSuite) TestForgotPassword() {
	// Create test user
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "forgot@example.com",
		Username:       "forgotuser",
		PasswordHash:   "hashedpassword",
		EmailVerified:  true,
	}
	suite.db.Create(&user)

	// Test forgot password
	req := ForgotPasswordRequest{
		Email: "forgot@example.com",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Check if reset token was generated
	var updatedUser models.User
	suite.db.First(&updatedUser, user.ID)
	assert.NotEmpty(suite.T(), updatedUser.PasswordResetToken)
	assert.NotZero(suite.T(), updatedUser.PasswordResetExpires)
}

func (suite *AuthTestSuite) TestResetPassword() {
	// Create test user with reset token
	resetToken := generateSecureToken()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), suite.config.BcryptCost)
	user := models.User{
		StellarAddress: "GA7QYNF7SOWQ3GLR2BGMZEHXAVIRZA4KVWLTJJFC7MGXUA74P7UJVSGZ",
		Email:          "reset@example.com",
		Username:       "resetuser",
		PasswordHash:   string(hashedPassword),
		EmailVerified:  true,
		PasswordResetToken: resetToken,
		PasswordResetExpires: time.Now().Add(time.Hour),
	}
	suite.db.Create(&user)

	// Test password reset
	req := ResetPasswordRequest{
		Token:    resetToken,
		Password: "newpassword123",
	}

	jsonData, _ := json.Marshal(req)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(jsonData))
	r.Header.Set("Content-Type", "application/json")

	suite.router.ServeHTTP(w, r)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Check if password was updated and tokens cleared
	var updatedUser models.User
	suite.db.First(&updatedUser, user.ID)
	assert.Empty(suite.T(), updatedUser.PasswordResetToken)
	assert.Zero(suite.T(), updatedUser.PasswordResetExpires)
	assert.NotEqual(suite.T(), string(hashedPassword), updatedUser.PasswordHash)

	// Verify new password works
	err := bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("newpassword123"))
	assert.NoError(suite.T(), err)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}