package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/surrealdb/surrealdb.go"
)

func checkBalance(c echo.Context) error {
	balanceData := new(BalanceRoute)
	if err := c.Bind(balanceData); err != nil {
		return c.String(http.StatusCreated, err.Error())
	}
	if balanceData.Acc1 == "" || balanceData.Pin == "" {
		return c.String(http.StatusBadRequest, "missing parameters")
	}
	balanceData.Acc1 = strings.ReplaceAll(balanceData.Acc1, " ", "")
	balanceData.Pin = strings.ReplaceAll(balanceData.Pin, " ", "")

	fmt.Println(balanceData.Acc1, balanceData.Pin)
	balance, err := balanceCheck(BalanceCheck{ID: "user:" + balanceData.Acc1, Pin: balanceData.Pin})
	if err != nil {
		return c.String(http.StatusCreated, err.Error())
	}
	return c.String(http.StatusOK, balance)
}

func balanceCheck(account BalanceCheck) (string, error) {
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	if _, err := db.Signin(map[string]interface{}{
		"user": "guffe",
		"pass": DATABASE_PASSWORD,
	}); err != nil {
		fmt.Println(err)
		return "", err
	}

	if _, err := db.Use("user", "user"); err != nil {
		return "", fmt.Errorf("failed to use database: %w", err)
	}
	fmt.Print(account.ID)
	data, err := db.Select(account.ID)
	if err != nil {
		return "", fmt.Errorf("failed to select account with ID %s: %w", account.ID, err)
	}

	acc := new(Account)
	err = surrealdb.Unmarshal(data, &acc)
	if err != nil {
		fmt.Println(data)
		return "", fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	if failedAttempts[account.ID] > 3 {
		return "", errors.New("suspended")
	}
	if !CheckPasswordHash(account.Pin, acc.Pin) {
		_, ok := failedAttempts[account.ID]
		if !ok {
			failedAttempts[account.ID] = 1
		}
		failedAttempts[account.ID] += 1
		if failedAttempts[account.ID] == 3 {
			go resetTimer(account.ID)
			return "", errors.New("suspended")
		}

		return "", errors.New("incorrect pin")
	}
	return acc.Balance, nil
}
