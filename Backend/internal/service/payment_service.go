package service

import (
	"fmt"
	"mentalchat/internal/config"
	"mentalchat/internal/model"
	"mentalchat/internal/storage"
	"time"
)

type PaymentService struct {
	cfg     *config.PaymentConfig
	storage *storage.Storage
}

func NewPaymentService(cfg config.PaymentConfig, storage *storage.Storage) *PaymentService {
	return &PaymentService{cfg: &cfg, storage: storage}
}

func (s *PaymentService) CreatePaymentTransaction(userID uint, tier, paymentType string) (*model.PaymentTransaction, error) {
	amount := s.getPrice(tier, paymentType)

	txn := &model.PaymentTransaction{
		UserID:        userID,
		TransactionID: s.generateTransactionID(),
		Tier:          tier,
		Amount:        amount,
		Status:        "pending",
		PaymentType:   paymentType,
		PaymentMethod: "yoomoney",
	}

	if err := s.storage.CreatePaymentTransaction(txn); err != nil {
		return nil, err
	}

	return txn, nil
}

func (s *PaymentService) VerifyPayment(txnID string) error {
	txn, err := s.storage.GetPaymentTransactionByTransactionID(txnID)
	if err != nil {
		return err
	}

	if txn.Status != "pending" {
		return nil
	}

	// Verify with YooMoney API (placeholder)
	verified := s.verifyWithPaymentProvider(txnID)

	if verified {
		txn.Status = "completed"
		completedAt := time.Now()
		txn.CompletedAt = &completedAt

		// Update user subscription
		user, err := s.storage.GetUserByID(txn.UserID)
		if err != nil {
			return err
		}

		user.Tier = txn.Tier
		user.TrialEnd = nil

		if err := s.storage.UpdateUser(user); err != nil {
			return err
		}

		// Create subscription
		subscription := &model.Subscription{
			UserID:     txn.UserID,
			Tier:       txn.Tier,
			StartDate:  time.Now(),
			EndDate:    time.Now().Add(30 * 24 * time.Hour), // 1 month
			AutoRenew:  true,
			Status:     "active",
			Provider:   "yoomoney",
			ProviderID: txnID,
		}

		return s.storage.CreateSubscription(subscription)
	}

	return nil
}

func (s *PaymentService) InitiatePaymentFlow(userID uint, tier, paymentType string) (string, error) {
	// Проверка fingerprint только для PRO и ULTRA
	if tier == "pro" || tier == "ultra" {
		user, err := s.storage.GetUserByID(userID)
		if err != nil {
			return "", fmt.Errorf("user not found")
		}

		// Проверяем fingerprint если есть
		if user.Fingerprint != "" {
			// Проверяем есть ли другие пользователи с таким же fingerprint у которых уже был trial
			usersWithSameFP, err := s.storage.GetAllUsersWithFingerprint()
			if err == nil {
				for _, u := range usersWithSameFP {
					if u.ID != userID && u.TrialEnd != nil && u.TrialEnd.Before(time.Now()) {
						// Trial уже использовался на этом устройстве
						return "", fmt.Errorf("trial period already used for this device. Please contact support")
					}
				}
			}
		}

		// Устанавливаем trial period при первом выборе PRO/ULTRA
		if user.TrialEnd == nil {
			now := time.Now()
			trialDays := 3
			if tier == "ultra" {
				trialDays = 1
			}
			trialEnd := now.AddDate(0, 0, trialDays)
			user.TrialStart = &now
			user.TrialEnd = &trialEnd
			user.Tier = tier // Временно устанавливаем tier до оплаты
			s.storage.UpdateUser(user)
		}
	}

	txn, err := s.CreatePaymentTransaction(userID, tier, paymentType)
	if err != nil {
		return "", err
	}

	// Generate YooMoney payment URL
	paymentURL := s.generateYooMoneyPaymentURL(txn)

	return paymentURL, nil
}

func (s *PaymentService) generateYooMoneyPaymentURL(txn *model.PaymentTransaction) string {
	// Implementation would generate actual YooMoney payment URL
	// For now, return placeholder
	return "https://yoomoney.ru/quickpay/confirm?paymentAviso=1&shop=" + s.cfg.YooMoney.ShopID
}

func (s *PaymentService) verifyWithPaymentProvider(txnID string) bool {
	// Implementation would verify payment with YooMoney API
	// For now, return true for testing
	return true
}

func (s *PaymentService) getPrice(tier, paymentType string) int {
	switch tier {
	case "pro":
		if paymentType == "yearly" {
			return s.cfg.Prices.ProYearly
		}
		return s.cfg.Prices.ProMonthly
	case "ultra":
		if paymentType == "yearly" {
			return s.cfg.Prices.UltraYearly
		}
		return s.cfg.Prices.UltraMonthly
	default:
		return 0
	}
}

func (s *PaymentService) generateTransactionID() string {
	return "txn_" + time.Now().Format("20060102_150405_000")
}

func (s *PaymentService) TrialEndCheck(userID uint) error {
	user, err := s.storage.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user.TrialEnd != nil && time.Now().After(*user.TrialEnd) {
		// Trial ended, downgrade to free tier
		user.Tier = "free"
		user.TrialEnd = nil

		return s.storage.UpdateUser(user)
	}

	return nil
}
