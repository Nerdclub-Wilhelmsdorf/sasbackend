package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/surrealdb/surrealdb.go"
)

func getLogs(c echo.Context) error {
	logs := new(GetLogs)
	if err := c.Bind(logs); err != nil {
		return err
	}

	fmt.Println(logs)
	getLogs, err := readLogs("user:"+logs.Acc, logs.Pin)
	if err != nil {
		return c.String(http.StatusTeapot, err.Error())
	}
	if err != nil {
		return c.String(http.StatusTeapot, err.Error())
	}
	return c.JSON(http.StatusOK, getLogs)
}

func readLogs(ID string, PIN string) (string, error) {
	fmt.Println(ID, PIN)
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return "", fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(ID)
	if err != nil {
		return "", fmt.Errorf("failed to select account with ID %s: %w", ID, err)
	}
	acc1 := new(Account)
	err = surrealdb.Unmarshal(data, &acc1)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	if !CheckPasswordHash(PIN, acc1.Pin) {
		return "", fmt.Errorf("incorrect pin")
	}

	if failedAttempts[ID] > 3 {
		return "", fmt.Errorf("suspended")
	}
	fmt.Println(acc1.Transactions)
	if acc1.Transactions == "" {
		return "", fmt.Errorf("no transactions found for account with ID %s", ID)
	} else {
		return acc1.Transactions, nil
	}

	return acc1.Transactions, nil
}