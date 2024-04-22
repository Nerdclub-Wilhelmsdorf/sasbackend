package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/surrealdb/surrealdb.go"
)

func pay(c echo.Context) error {
	paymentData := new(PaymentRoute)
	if err := c.Bind(paymentData); err != nil {
		return err
	}
	if paymentData.Acc1 == "" || paymentData.Acc2 == "" || paymentData.Amount == "" || paymentData.Pin == "" {
		return c.String(http.StatusBadRequest, "missing parameters")
	}
	err := transferMoney(Transfer{From: "user:" + paymentData.Acc1, To: "user:" + paymentData.Acc2, Amount: paymentData.Amount, Pin: paymentData.Pin})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "payment successful")
}

func transferMoney(transfer Transfer) error {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(transfer.From)
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", transfer.From, err)
	}
	acc1 := new(Account)
	err = surrealdb.Unmarshal(data, &acc1)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	if failedAttempts[transfer.From] > 3 {
		return fmt.Errorf("suspended")
	}

	if !CheckPasswordHash(transfer.Pin, acc1.Pin) {
		_, ok := failedAttempts[transfer.From]
		if !ok {
			failedAttempts[transfer.From] = 1
		}
		failedAttempts[transfer.From] += 1
		if failedAttempts[transfer.From] == 3 {
			go resetTimer(transfer.From)
			return errors.New("suspended")
		}

		return errors.New("incorrect pin")

	}
	amount, err := strconv.ParseFloat(transfer.Amount, 64)
	if err != nil {
		return fmt.Errorf("failed to parse transfer amount: %w", err)
	}
	balance, err := strconv.ParseFloat(acc1.Balance, 64)
	fmt.Print("balance: " + acc1.Balance + " amount: " + transfer.Amount)
	if err != nil {
		return err
	}
	if amount <= 0 {
		return errors.New("invalid amount")
	}
	if amount*1.1 >= balance {
		return errors.New("insufficient funds")
	}
	changes := map[string]string{"balance": fmt.Sprintf("%f", (balance - amount*1.1)), "name": acc1.Name, "pin": acc1.Pin}
	if _, err = db.Update(transfer.From, changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", transfer.From, err)
	}
	data, err = db.Select(transfer.To)
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", transfer.To, err)
	}
	acc2 := new(Account)
	err = surrealdb.Unmarshal(data, &acc2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	balance, err = strconv.ParseFloat(acc2.Balance, 64)
	if err != nil {
		return err
	}
	changes = map[string]string{"balance": fmt.Sprintf("%f", (amount + balance)), "name": acc2.Name, "pin": acc2.Pin}
	if _, err = db.Update(transfer.To, changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", transfer.To, err)
	}
	data, err = db.Select("user:zentralbank")
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", "user:zentralbank", err)
	}
	acc3 := new(Account)
	err = surrealdb.Unmarshal(data, &acc3)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	changes = map[string]string{"balance": fmt.Sprintf("%f", amount*0.1+balance), "name": acc3.Name, "pin": acc3.Pin}
	if _, err = db.Update("user:zentralbank", changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", "user:zentralbank", err)
	}
	delete(failedAttempts, transfer.Pin)
	return nil
}
