package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/shopspring/decimal"
	"github.com/surrealdb/surrealdb.go"
)

func transferMoney_2(transfer Transfer) error {
	from, err := loadUser(transfer.From)
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	to, err := loadUser(transfer.To)
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	bank, err := loadUser("zentralbank")
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	if validateTransaction(from, transfer) != nil {
		return fmt.Errorf("failed to validate transaction: %w", err)
	}
	transferDecimal, err := decimal.NewFromString(transfer.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert transfer amount to decimal: %w", err)
	}

	//handle from
	fromDecimal, err := decimal.NewFromString(from.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance to decimal: %w", err)
	}
	from.Balance = fromDecimal.Sub(transferDecimal.Mul(decimal.NewFromFloat(taxFactor))).String()
	updatedTransaction, err := appendToLog(from, transfer)
	if err != nil {
		return fmt.Errorf("failed to append transaction to log: %w", err)
	}
	from.Transactions = updatedTransaction.Transactions
	if updateUser(from.ID, from) != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	//handle to
	toDecimal, err := decimal.NewFromString(to.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance to decimal: %w", err)
	}
	to.Balance = toDecimal.Add(transferDecimal).String()
	updatedTransaction, err = appendToLog(to, transfer)
	if err != nil {
		return fmt.Errorf("failed to append transaction to log: %w", err)
	}
	to.Transactions = updatedTransaction.Transactions
	if updateUser(to.ID, to) != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	bankDecimal, err := decimal.NewFromString(bank.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance to decimal: %w", err)
	}
	bank.Balance = bankDecimal.Add(transferDecimal.Mul(decimal.NewFromFloat(taxRate))).String()
	updatedTransaction, err = appendToLog(bank, transfer)
	if err != nil {
		return fmt.Errorf("failed to append transaction to log: %w", err)
	}
	bank.Transactions = updatedTransaction.Transactions
	if updateUser(bank.ID, bank) != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	logfile(TransactionLog{Time: currTime(), From: transfer.From, To: transfer.To, Amount: transfer.Amount})
	return nil
}

func loadUser(id string) (Account, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return Account{}, fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(id)
	if err != nil {
		return Account{}, fmt.Errorf("failed to select account with ID %s: %w", id, err)
	}
	acc := new(Account)
	err = surrealdb.Unmarshal(data, &acc)
	if err != nil {
		return Account{}, fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	return *acc, nil
}

func validateTransaction(payer Account, transfer Transfer) error {
	if !CheckPasswordHash(transfer.Pin, payer.Pin) {
		_, ok := failedAttempts[transfer.From]
		if !ok {
			failedAttempts[transfer.From] = 1
		}
		failedAttempts[transfer.From] += 1
		if failedAttempts[transfer.From] > 3 {
			if !slices.Contains(failedAttemtsCurrentlyLocking, transfer.From) {
				go resetTimer(transfer.From)
			}
			return fmt.Errorf("suspended")
		}
		return fmt.Errorf("wrong pin")
	}
	if failedAttempts[transfer.From] > 3 {
		return fmt.Errorf("suspended")
	}
	payerDecimal, err := decimal.NewFromString(payer.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance to decimal: %w", err)
	}
	transferDecimal, err := decimal.NewFromString(transfer.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert transfer amount to decimal: %w", err)
	}
	if transferDecimal.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("bad amount")
	}
	if payerDecimal.Mul(decimal.NewFromFloat(taxFactor)).LessThan(transferDecimal) {
		return fmt.Errorf("insufficient funds")
	}
	return nil
}

func updateUser(id string, acc Account) error {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(id)
	if err != nil {
		return fmt.Errorf("failed to select account with ID %s: %w", id, err)
	}
	acc2 := new(Account)
	err = surrealdb.Unmarshal(data, &acc2)
	if err != nil {
		return fmt.Errorf("failed to unmarshal account data: %w", err)
	}

	changes := map[string]string{"balance": acc.Balance, "name": acc.Name, "pin": acc.Pin, "transactions": acc.Transactions}
	if _, err = db.Update(id, changes); err != nil {
		return fmt.Errorf("failed to update account with ID %s: %w", id, err)
	}
	return nil
}

func appendToLog(acc1 Account, transfer Transfer) (Account, error) {
	var transactions []TransactionLog
	if acc1.Transactions == "" {
		transactions = []TransactionLog{}
	} else {
		err := json.Unmarshal([]byte(acc1.Transactions), &transactions)
		if err != nil {
			return Account{}, fmt.Errorf("failed to unmarshal transactions: %w", err)
		}
	}
	transactions = append(transactions, TransactionLog{Time: currTime(), From: transfer.From, To: transfer.To, Amount: transfer.Amount})
	transactionsJSON, err := json.Marshal(transactions)
	if err != nil {
		return Account{}, fmt.Errorf("failed to marshal transactions: %w", err)
	}
	transactionsString := string(transactionsJSON)
	acc1.Transactions = transactionsString
	return acc1, nil
}
