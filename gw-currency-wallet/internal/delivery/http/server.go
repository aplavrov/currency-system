//go:generate mockgen -source ./server.go -destination=./mocks/server.go -package=mock_server

package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/aplavrov/currency-system/gw-currency-wallet/docs"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/delivery"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type AuthService interface {
	Register(ctx context.Context, username string, password string, email string) error
	Login(ctx context.Context, username string, password string) (string, error)
}

type WalletService interface {
	Deposit(ctx context.Context, walletID uuid.UUID, currency string, amount float64) (service.Balance, error)
	Withdraw(ctx context.Context, walletID uuid.UUID, currency string, amount float64) (service.Balance, error)
	GetBalance(ctx context.Context, walletID uuid.UUID) (service.Balance, error)
}

type ExchangeService interface {
	Exchange(ctx context.Context, walletID uuid.UUID, FromCurrency string, ToCurrency string, Amount float64) (float64, service.Balance, error)
	GetRates(ctx context.Context) (service.Rates, error)
}

type WalletServer struct {
	authService     AuthService
	walletService   WalletService
	exchangeService ExchangeService
	jwtSecret       string
	logger          *slog.Logger
}

func NewWalletServer(authService AuthService, walletService WalletService, exchangeService ExchangeService, jwtSecret string, logger *slog.Logger) *WalletServer {
	return &WalletServer{authService: authService, walletService: walletService, exchangeService: exchangeService, jwtSecret: jwtSecret, logger: logger}
}

func (s *WalletServer) Start(port string) error {
	s.logger.Info("gin server started", "port", port)
	gin.SetMode(gin.ReleaseMode)
	server := gin.Default()
	docs.SwaggerInfo.BasePath = "/api/v1"

	server.POST("/api/v1/register", s.Register)
	server.POST("/api/v1/login", s.Login)

	authGroup := server.Group("/api/v1")
	authGroup.Use(AuthenticationMiddleware(s.jwtSecret))

	authGroup.GET("/balance", s.GetBalance)
	authGroup.POST("/wallet/deposit", s.Deposit)
	authGroup.POST("/wallet/withdraw", s.Withdraw)
	authGroup.GET("/exchange/rates", s.GetRates)
	authGroup.POST("/exchange", s.Exchange)

	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return server.Run(fmt.Sprintf(":%v", port))
}

// @Summary Register a new user
// @Description Create a new user account with username, password, and email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body delivery.RegisterRequest true "Register Request"
// @Success 201 {object} map[string]string "User registered successfully"
// @Failure 400 {object} map[string]string "Bad request / Username or email exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /register [post]
func (s *WalletServer) Register(ctx *gin.Context) {
	var req delivery.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		s.logger.Error("invalid JSON", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if req.Username == "" {
		s.logger.Error("username is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username is missing"})
		return
	}

	if req.Password == "" {
		s.logger.Error("password is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "password is missing"})
		return
	}

	if req.Email == "" {
		s.logger.Error("email is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is missing"})
		return
	}

	err := s.authService.Register(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		if errors.Is(err, service.ErrUsernameOrEmailExists) {
			s.logger.Error("username or email already exists")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
			return
		}
		s.logger.Error("internal error", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body delivery.LoginRequest true "Login Request"
// @Success 200 {object} map[string]string "JWT token returned"
// @Failure 400 {object} map[string]string "Invalid JSON / missing fields"
// @Failure 401 {object} map[string]string "Invalid username or password"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /login [post]
func (s *WalletServer) Login(ctx *gin.Context) {
	var req delivery.LoginRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		s.logger.Error("invalid JSON", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if req.Username == "" {
		s.logger.Error("username is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username is missing"})
		return
	}

	if req.Password == "" {
		s.logger.Error("password is missing")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "password is missing"})
		return
	}

	token, err := s.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			s.logger.Error("invalid username or password")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		s.logger.Error("internal error", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

func (s *WalletServer) getWalletID(ctx *gin.Context) (uuid.UUID, error) {
	logger := s.logger.With("method", "getWalletID")
	ctxWalletID, ok := ctx.Get("wallet_id")
	if !ok {
		logger.Error("no walletID in a JWT token")
		return [16]byte{}, fmt.Errorf("provided token has no walletID")
	}

	walletID, err := uuid.Parse(ctxWalletID.(string))
	if err != nil {
		logger.Error("invalid walletID", "error", err)
		return [16]byte{}, fmt.Errorf("invalid walletID")
	}
	return walletID, nil
}

// @Summary Get wallet balance
// @Description Returns balances for all currencies of the authenticated user's wallet
// @Tags Wallet
// @Produce json
// @Success 200 {object} delivery.BalanceResponse "Wallet balance"
// @Failure 500 {object} map[string]string "Internal server error"
// @Failure 404 {object} map[string]string "Wallet not found"
// @Security ApiKeyAuth
// @Router /balance [get]
func (s *WalletServer) GetBalance(ctx *gin.Context) {
	logger := s.logger.With("method", "GetBalance")
	walletID, err := s.getWalletID(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	balance, err := s.walletService.GetBalance(ctx.Request.Context(), walletID)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			logger.Error("wallet not found", "error", err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, delivery.BalanceResponse{Balance: balance})
}

// @Summary Deposit funds to wallet
// @Description Deposit amount to the specified currency in the authenticated user's wallet
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body delivery.DepositRequest true "Deposit Request"
// @Success 200 {object} delivery.WalletOperationResponse "Deposit successful, new balances returned"
// @Failure 400 {object} map[string]string "Invalid amount or currency"
// @Failure 404 {object} map[string]string "Wallet not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /wallet/deposit [post]
func (s *WalletServer) Deposit(ctx *gin.Context) {
	logger := s.logger.With("method", "GetBalance")
	var req delivery.DepositRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("Error: invalid JSON: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	walletID, err := s.getWalletID(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("deposit", "walletID", walletID, "currency", req.Currency, "amount", req.Amount)

	balance, err := s.walletService.Deposit(ctx.Request.Context(), walletID, req.Currency, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidAmount) || errors.Is(err, service.ErrInvalidCurrency) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount or currency"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, delivery.WalletOperationResponse{
		Message:    "Account topped up successfully",
		NewBalance: balance,
	})
}

// @Summary Withdraw funds from wallet
// @Description Withdraw amount from the specified currency in the authenticated user's wallet
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body delivery.WithdrawRequest true "Withdraw Request"
// @Success 200 {object} delivery.WalletOperationResponse "Withdrawal successful, new balances returned"
// @Failure 400 {object} map[string]string "Insufficient funds or invalid amount"
// @Failure 404 {object} map[string]string "Wallet not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /wallet/withdraw [post]
func (s *WalletServer) Withdraw(ctx *gin.Context) {
	logger := s.logger.With("method", "Withdraw")
	var req delivery.WithdrawRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("Error: invalid JSON: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	walletID, err := s.getWalletID(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("withdraw", "walletID", walletID, "currency", req.Currency, "amount", req.Amount)

	balance, err := s.walletService.Withdraw(ctx.Request.Context(), walletID, req.Currency, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidAmount) || errors.Is(err, service.ErrInvalidCurrency) || errors.Is(err, service.ErrInsufficientFunds) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds or invalid amount"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	ctx.JSON(http.StatusOK, delivery.WalletOperationResponse{
		Message:    "Withdrawal successful",
		NewBalance: balance,
	})
}

// @Summary Get exchange rates
// @Description Get current exchange rates for supported currencies
// @Tags Exchange
// @Produce json
// @Success 200 {object} delivery.RatesResponse "Exchange rates returned"
// @Failure 500 {object} map[string]string "Failed to retrieve exchange rates"
// @Security ApiKeyAuth
// @Router /exchange/rates [get]
func (s *WalletServer) GetRates(ctx *gin.Context) {
	logger := s.logger.With("method", "GetRates")
	logger.Info("requesting rates")

	rates, err := s.exchangeService.GetRates(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve exchange rates"})
		return
	}

	logger.Info("retrieved rates")
	ctx.JSON(http.StatusOK, delivery.RatesResponse{Rates: rates})
}

// @Summary Exchange currencies
// @Description Convert amount from one currency to another for the authenticated user's wallet
// @Tags Exchange
// @Accept json
// @Produce json
// @Param request body delivery.ExchangeRequest true "Exchange Request"
// @Success 200 {object} delivery.ExchangeResponse "Exchange successful, new balances returned"
// @Failure 400 {object} map[string]string "Insufficient funds or invalid currencies"
// @Failure 404 {object} map[string]string "Wallet not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Security ApiKeyAuth
// @Router /exchange [post]
func (s *WalletServer) Exchange(ctx *gin.Context) {
	logger := s.logger.With("method", "Exchange")
	var req delivery.ExchangeRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		log.Printf("Error: invalid JSON: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	walletID, err := s.getWalletID(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	logger.Info("exchange", "walletID", walletID, "from_currency", req.FromCurrency, "currency", req.ToCurrency, "amount", req.Amount)

	exchanged, balance, err := s.exchangeService.Exchange(ctx.Request.Context(), walletID, req.FromCurrency, req.ToCurrency, req.Amount)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "wallet not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidAmount) || errors.Is(err, service.ErrInvalidCurrency) || errors.Is(err, service.ErrInsufficientFunds) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds or invalid currencies"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}

	ctx.JSON(http.StatusOK, delivery.ExchangeResponse{
		Message:         "Exchange successful",
		ExchangedAmount: exchanged,
		NewBalance:      balance,
	})
}
