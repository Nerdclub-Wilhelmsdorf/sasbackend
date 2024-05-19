package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
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
	paymentData.Acc1 = strings.ReplaceAll(paymentData.Acc1, " ", "")
	paymentData.Acc2 = strings.ReplaceAll(paymentData.Acc2, " ", "")
	paymentData.Amount = strings.ReplaceAll(paymentData.Amount, " ", "")
	paymentData.Amount = strings.ReplaceAll(paymentData.Amount, ",", ".")
	paymentData.Pin = strings.ReplaceAll(paymentData.Pin, " ", "")

	err := transferMoney(Transfer{From: "user:" + paymentData.Acc1, To: "user:" + paymentData.Acc2, Amount: paymentData.Amount, Pin: paymentData.Pin})
	if err != nil {
		return c.String(http.StatusCreated, err.Error())
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
	_, err = db.Select(transfer.To)
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
			if !slices.Contains(failedAttemtsCurrentlyLocking, transfer.From) {
				go resetTimer(transfer.From)
			}
			return errors.New("suspended")
		}

		return errors.New("incorrect pin")

	}
	amount, err := decimal.NewFromString(transfer.Amount)
	if err != nil {
		return fmt.Errorf("failed to parse transfer amount: %w", err)
	}
	balance, err := decimal.NewFromString(acc1.Balance)
	fmt.Print("balance: " + acc1.Balance + " amount: " + transfer.Amount)
	if err != nil {
		return err
	}
	if decimal.NewFromInt(0).GreaterThanOrEqual(amount) {
		return errors.New("invalid amount")
	}
	// 10% tax
	if amount.Mul(decimal.NewFromFloat(1.1)).GreaterThan(balance) {
		return errors.New("insufficient funds")
	}
	var transactions []TransactionLog
	if acc1.Transactions == "" {
		transactions = []TransactionLog{}
	} else {
		err = json.Unmarshal([]byte(acc1.Transactions), &transactions)
		if err != nil {
			return fmt.Errorf("failed to unmarshal transactions: %w", err)
		}
	}
	err = logfile(TransactionLog{Time: currTime(), From: transfer.From, To: transfer.To, Amount: transfer.Amount})
	if err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}
	transactions = append(transactions, TransactionLog{Time: currTime(), From: transfer.From, To: transfer.To, Amount: transfer.Amount})
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}
	transactionsString := string(transactionsJSON)
	fmt.Println("transactions: " + transactionsString)
	changes := map[string]string{"balance": balance.Sub(amount.Mul(decimal.NewFromFloat(1.1))).String(), "name": acc1.Name, "pin": acc1.Pin, "transactions": transactionsString}
	if _, err = db.Update(transfer.From, changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", transfer.From, err)
	}
	data, err = db.Select(transfer.To)
	amount = amount.Div(decimal.NewFromFloat(1.1))
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", transfer.To, err)
	}
	acc2 := new(Account)
	err = surrealdb.Unmarshal(data, &acc2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	if acc2.Transactions == "" {
		transactions = []TransactionLog{}
	} else {
		err = json.Unmarshal([]byte(acc2.Transactions), &transactions)
		if err != nil {
			return fmt.Errorf("failed to unmarshal transactions: %w", err)
		}
	}
	transactionsReciever := append(transactions, TransactionLog{Time: currTime(), From: transfer.From, To: transfer.To, Amount: amount.Div(decimal.NewFromFloat(1.1)).String()})
	transactionsRecieverJSON, err2 := json.Marshal(transactionsReciever)
	if err2 != nil {
		return fmt.Errorf("failed to unmarshal transaction")
	}
	transactionsRecieverString := string(transactionsRecieverJSON)

	balance, err = decimal.NewFromString(acc2.Balance)
	if err != nil {
		return err
	}
	changes = map[string]string{"balance": amount.Add(balance).String(), "name": acc2.Name, "pin": acc2.Pin, "transactions": transactionsRecieverString}
	if _, err = db.Update(transfer.To, changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", transfer.To, err)
	}
	data, err = db.Select("user:zentralbank")
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", "user:zentralbank", err)
	}
	acc3 := new(Account)
	err = surrealdb.Unmarshal(data, &acc3)
	if acc3.Transactions == "" {
		transactions = []TransactionLog{}
	} else {
		err = json.Unmarshal([]byte(acc3.Transactions), &transactions)
		if err != nil {
			return fmt.Errorf("failed to unmarshal transactions: %w", err)
		}
	}
	transactionsBank := append(transactions, TransactionLog{Time: currTime(), From: transfer.From, To: "zentralbank", Amount: amount.Mul(decimal.NewFromFloat(0.1)).String()})
	transactionsBankJSON, err2 := json.Marshal(transactionsBank)
	if err2 != nil {
		return fmt.Errorf("failed to unmarshal transaction")
	}
	transactionsBankString := string(transactionsBankJSON)
	balance, err = decimal.NewFromString(acc3.Balance)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	changes = map[string]string{"balance": amount.Mul(decimal.NewFromFloat(0.1)).Add(balance).String(), "name": acc3.Name, "pin": acc3.Pin, "transactions": transactionsBankString}
	if _, err = db.Update("user:zentralbank", changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", "user:zentralbank", err)
	}
	delete(failedAttempts, transfer.Pin)
	return nil
}
