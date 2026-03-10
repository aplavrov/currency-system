package server

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	mock_server "github.com/aplavrov/currency-system/gw-currency-wallet/internal/delivery/http/mocks"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWalletServer_Register_JSON_validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Parallel()

	tests := []struct {
		name          string
		body          string
		wantCode      int
		wantJSONError string
	}{
		{
			name:          "invalid json",
			body:          `{`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"invalid json"}`,
		},
		{
			name: "username missing",
			body: `{
				"password":"123",
				"email":"test@gmail.com"
			}`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"username is missing"}`,
		},
		{
			name: "password missing",
			body: `{
				"username":"test",
				"email":"test@gmail.com"
			}`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"password is missing"}`,
		},
		{
			name: "email missing",
			body: `{
				"username":"test",
				"password":"123"
			}`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"email is missing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			server := NewWalletServer(nil, nil, nil, "", slog.Default())

			server.Register(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			require.JSONEq(t, tt.wantJSONError, w.Body.String())
		})
	}
}

func TestServer_Register_logic_validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Parallel()

	tests := []struct {
		name          string
		wantCode      int
		wantJSONError string
		wantError     error
	}{
		{
			name:          "username exists",
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"Username or email already exists"}`,
			wantError:     service.ErrUsernameOrEmailExists,
		},
		{
			name:          "internal error",
			wantCode:      http.StatusInternalServerError,
			wantJSONError: `{"error":"Internal error"}`,
			wantError:     errors.New(""),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			body := `{
				"username":"test",
				"password":"123",
				"email":"test@gmail.com"
			}`

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			ctrl := gomock.NewController(t)

			mockAuth := mock_server.NewMockAuthService(ctrl)
			mockAuth.EXPECT().
				Register(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.wantError)

			server := NewWalletServer(mockAuth, nil, nil, "", slog.Default())

			server.Register(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			require.JSONEq(t, tt.wantJSONError, w.Body.String())
		})
	}
}

func TestWalletServer_Login_JSON_validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Parallel()

	tests := []struct {
		name          string
		body          string
		wantCode      int
		wantJSONError string
	}{
		{
			name:          "invalid json",
			body:          `{`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"invalid json"}`,
		},
		{
			name: "username missing",
			body: `{
				"password":"123"
			}`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"username is missing"}`,
		},
		{
			name: "password missing",
			body: `{
				"username":"test"
			}`,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"password is missing"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			ctx.Request = req

			server := NewWalletServer(nil, nil, nil, "", slog.Default())

			server.Login(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			require.JSONEq(t, tt.wantJSONError, w.Body.String())
		})
	}
}

func TestServer_Login_invalid_credentials(t *testing.T) {

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	body := `{
		"username":"test",
		"password":"123"
	}`

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	ctrl := gomock.NewController(t)

	mockAuth := mock_server.NewMockAuthService(ctrl)
	mockAuth.EXPECT().
		Login(gomock.Any(), gomock.Any(), gomock.Any()).
		Return("", service.ErrInvalidCredentials)

	server := NewWalletServer(mockAuth, nil, nil, "", slog.Default())

	server.Login(ctx)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"error":"Invalid username or password"}`, w.Body.String())
}

func TestServer_Deposit_logic_validation(t *testing.T) {

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		serviceError  error
		wantCode      int
		wantJSONError string
	}{
		{
			name:          "wallet not found",
			serviceError:  service.ErrWalletNotFound,
			wantCode:      http.StatusNotFound,
			wantJSONError: `{"error":"wallet not found"}`,
		},
		{
			name:          "invalid amount",
			serviceError:  service.ErrInvalidAmount,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"Invalid amount or currency"}`,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			ctx.Set("wallet_id", uuid.New().String())

			body := `{
				"currency":"USD",
				"amount":100
			}`

			req := httptest.NewRequest(http.MethodPost, "/wallet/deposit", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			ctrl := gomock.NewController(t)

			mockWallet := mock_server.NewMockWalletService(ctrl)

			mockWallet.EXPECT().
				Deposit(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(service.Balance{}, tt.serviceError)

			server := NewWalletServer(nil, mockWallet, nil, "", slog.Default())

			server.Deposit(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			require.JSONEq(t, tt.wantJSONError, w.Body.String())
		})
	}
}

func TestServer_Withdraw_logic_validation(t *testing.T) {

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		serviceError  error
		wantCode      int
		wantJSONError string
	}{
		{
			name:          "wallet not found",
			serviceError:  service.ErrWalletNotFound,
			wantCode:      http.StatusNotFound,
			wantJSONError: `{"error":"wallet not found"}`,
		},
		{
			name:          "insufficient funds",
			serviceError:  service.ErrInsufficientFunds,
			wantCode:      http.StatusBadRequest,
			wantJSONError: `{"error":"Insufficient funds or invalid amount"}`,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			ctx.Set("wallet_id", uuid.New().String())

			body := `{
				"currency":"USD",
				"amount":100
			}`

			req := httptest.NewRequest(http.MethodPost, "/wallet/withdraw", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			ctrl := gomock.NewController(t)

			mockWallet := mock_server.NewMockWalletService(ctrl)

			mockWallet.EXPECT().
				Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(service.Balance{}, tt.serviceError)

			server := NewWalletServer(nil, mockWallet, nil, "", slog.Default())

			server.Withdraw(ctx)

			require.Equal(t, tt.wantCode, w.Code)
			require.JSONEq(t, tt.wantJSONError, w.Body.String())
		})
	}
}

func TestServer_Exchange(t *testing.T) {

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	ctx.Set("wallet_id", uuid.New().String())

	body := `{
		"fromCurrency":"USD",
		"toCurrency":"EUR",
		"amount":100
	}`

	req := httptest.NewRequest(http.MethodPost, "/exchange", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	ctx.Request = req

	ctrl := gomock.NewController(t)

	mockExchange := mock_server.NewMockExchangeService(ctrl)

	mockExchange.EXPECT().
		Exchange(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(90.0, service.Balance{}, nil)

	server := NewWalletServer(nil, nil, mockExchange, "", slog.Default())

	server.Exchange(ctx)

	require.Equal(t, http.StatusOK, w.Code)
}
