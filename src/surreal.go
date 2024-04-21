package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/surrealdb/surrealdb.go"
	"golang.org/x/crypto/bcrypt"
)

func test() {
	// Create a new account
	user := Account{
		Name:    "203876",
		Balance: "1000",
		Pin:     "1234",
	}
	_, err := createAccount(user)
	if err != nil {
		panic(err)
	}

	// Check balance
	balance, err := balanceCheck(BalanceCheck{
		ID:  "user:36xxj2a4al7ybid1j7pd",
		Pin: "1234",
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Balance:", balance)

	// Transfer money
	transfer := Transfer{
		From:   "user:36xxj2a4al7ybid1j7pd",
		To:     "user:36xxj2a4al7ybid1j7pd",
		Amount: "100",
		Pin:    "1234",
	}
	err = transferMoney(transfer)
	if err != nil {
		panic(err)
	}

	// Check balance
	balance, err = balanceCheck(BalanceCheck{
		ID:  "user:36xxj2a4al7ybid1j7pd",
		Pin: user.Pin,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Balance:", balance)
}

func createAccount(user Account) (string, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return "", err
	}

	// Insert user
	data, err := db.Create("user", user)
	if err != nil {
		return "", err
	}
	createdUser := make([]Account, 1)
	err = surrealdb.Unmarshal(data, &createdUser)
	if err != nil {
		return "", err
	}

	// Get user by ID
	data, err = db.Select(createdUser[0].ID)
	if err != nil {
		return "", err
	}
	selectedUser := new(Account)
	err = surrealdb.Unmarshal(data, &selectedUser)
	if err != nil {
		return "", err
	}
	if user.Name == selectedUser.Name {
		fmt.Println("User with ID: " + selectedUser.ID + " created successfully.")
		return selectedUser.ID, nil
	} else {
		return "", errors.New("failed to create user")
	}

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
	if !CheckPasswordHash(transfer.Pin, acc1.Pin) {
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

	return nil
}
func balanceCheck(account BalanceCheck) (string, error) {
	db, err := surrealdb.New("ws://localhost:8000/rpc")
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

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
	if !CheckPasswordHash(account.Pin, acc.Pin) {
		return "", errors.New("incorrect pin")
	}
	return acc.Balance, nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
