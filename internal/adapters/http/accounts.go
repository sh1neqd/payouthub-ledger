package server

import (
	"context"
	"errors"
	"ledger/internal/application"
	ledger "ledger/internal/domain"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func createAccountHandler(service *application.AccountService) fiber.Handler {
	return func(c fiber.Ctx) error {
		var body struct {
			MerchantID string `json:"merchant_id"`
			Currency   string `json:"currency"`
		}

		if err := c.Bind().Body(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		merchantID, err := uuid.Parse(body.MerchantID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid merchant id")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		account, err := service.CreateAccount(ctx, application.CreateAccountInput{
			MerchantID: merchantID,
			Currency:   body.Currency,
		})

		if errors.Is(err, ledger.ErrInvalidMerchantID) ||
			errors.Is(err, ledger.ErrUnsupportedCurrency) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if err != nil {
			log.Printf("create account: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "could not create account")
		}

		return c.Status(fiber.StatusCreated).JSON(accountResponse(account))
	}
}

func getAccountByIDHandler(service *application.AccountService) fiber.Handler {
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return fiber.NewError(
				fiber.StatusBadRequest,
				"id must be a positive integer",
			)
		}

		account, err := service.GetAccount(ctx, id)

		if errors.Is(err, ledger.ErrAccountNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "account not found")
		}

		if err != nil {
			log.Printf("get account: %v", err)

			return fiber.NewError(fiber.StatusInternalServerError, "could not get account")
		}

		return c.Status(fiber.StatusOK).JSON(accountResponse(account))
	}
}

func accountResponse(account ledger.Account) fiber.Map {
	return fiber.Map{
		"id":          account.Id,
		"kind":        account.Kind,
		"merchant_id": account.MerchantID,
		"currency":    account.Currency,
		"balance":     account.Balance,
		"reserved":    account.Reserved,
		"available":   account.Available(),
	}
}
