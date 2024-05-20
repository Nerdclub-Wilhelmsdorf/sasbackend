package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/surrealdb/surrealdb.go"
)

func getLogs(c echo.Context) error {
	logs := new(GetLogs)
	if err := c.Bind(logs); err != nil {
		return err
	}
	if logs.Acc == "" || logs.Pin == "" {
		return c.String(http.StatusBadRequest, "missing parameters")
	}
	logs.Acc = strings.ReplaceAll(logs.Acc, " ", "")
	logs.Pin = strings.ReplaceAll(logs.Pin, " ", "")
	fmt.Println(logs)
	getLogs, err := readLogs("user:"+logs.Acc, logs.Pin)
	if err != nil {
		return c.String(http.StatusCreated, err.Error())
	}
	return c.JSON(http.StatusOK, getLogs)
}

func readLogs(ID string, PIN string) (string, error) {
	fmt.Println(ID, PIN)
	db, _ := surrealdb.New("https://localhost:8000/rpc")
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
		return "", nil
	} else {
		return acc1.Transactions, nil
	}

}

func logfile(transaction TransactionLog) error {
	//create file if it doesn't exist
	if _, err := os.Stat("transactions.csv"); errors.Is(err, os.ErrNotExist) {
		os.Create("transactions.csv")
		//create header
		file, err := os.OpenFile("transactions.csv", os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		data := []string{"Time", "From", "To", "Amount"}
		err = writer.Write(data)
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile("transactions.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	data := []string{transaction.Time, strings.TrimPrefix(transaction.From, "user:"), strings.TrimPrefix(transaction.From, "user:"), transaction.Amount}
	err = writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}
