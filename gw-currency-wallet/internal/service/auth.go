package service

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userStorage   storage.UserStorage
	walletStorage storage.WalletStorage
	txManager     txManager
	JWTSecret     string
	TokenTTL      time.Duration
	logger        *slog.Logger
}

func NewAuthService(userStorage storage.UserStorage, walletStorage storage.WalletStorage, txManager txManager, JWTSecret string, tokenTTL time.Duration, logger *slog.Logger) *AuthService {
	return &AuthService{userStorage: userStorage, walletStorage: walletStorage, txManager: txManager, JWTSecret: JWTSecret, TokenTTL: tokenTTL, logger: logger}
}

func (s *AuthService) Register(ctx context.Context, username string, password string, email string) error {
	log := s.logger.With("method", "Register", "username", username)
	log.Info("attempting to register")

	userID := uuid.New()
	walletID := uuid.New()

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", "error", err)
		return err
	}

	err = s.txManager.RunSerializable(ctx, func(ctx context.Context) error {
		if err := s.walletStorage.CreateWallet(ctx, walletID); err != nil {
			return err
		}

		if err := s.userStorage.CreateUser(ctx, storage.CreateUserParams{
			UserID:   userID,
			Username: username,
			Email:    email,
			PassHash: passHash,
			WalletID: walletID,
		}); err != nil {
			if errors.Is(err, storage.ErrUsernameExists) || errors.Is(err, storage.ErrEmailExists) {
				return ErrUsernameOrEmailExists
			}
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	log.Info("user registered")
	return nil
}

func (s *AuthService) generateToken(userID uuid.UUID, walletID uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["wallet_id"] = walletID
	claims["exp"] = time.Now().Add(s.TokenTTL).Unix()

	tokenString, err := token.SignedString([]byte(s.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (string, error) {
	log := s.logger.With("method", "Login", "username", username)
	log.Info("attempting to authorize")

	user, err := s.userStorage.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrUsernameNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	log.Info("user authorized")

	token, err := s.generateToken(user.ID, user.WalletID)
	if err != nil {
		log.Error("failed to generate token", "error", err)
		return "", err
	}

	return token, nil
}
