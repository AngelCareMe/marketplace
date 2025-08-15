package bcrypt

import (
	"marketplace/pkg/errors"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type BcryptManager struct {
	logger *logrus.Logger
	cost   int
}

func NewBcryptManager(logger *logrus.Logger, cost int) *BcryptManager {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		logger.Warnf("Invalid bcrypt cost %d, using default %d", cost, bcrypt.DefaultCost)
		cost = bcrypt.DefaultCost
	}
	return &BcryptManager{logger: logger, cost: cost}
}

func (b *BcryptManager) GenerateHashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		b.logger.WithFields(logrus.Fields{
			"method": "GenerateHashPassword",
			"error":  err,
		}).Error("failed to generate bcrypt hash")

		return "", errors.NewAppError("HASHING", "failed to generate password hash", err)
	}

	return string(hashedBytes), nil
}

func (b *BcryptManager) CompareHashPassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			b.logger.Info("password mismatch")
			return errors.NewAppError("AUTH", "password mismatch", err)
		}

		b.logger.WithFields(logrus.Fields{
			"method": "CompareHashPassword",
			"error":  err,
		}).Error("failed to compare password hash")

		return errors.NewAppError("HASHING", "failed to compare password hash", err)
	}

	return nil
}
