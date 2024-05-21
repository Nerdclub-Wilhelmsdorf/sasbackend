package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/surrealdb/surrealdb.go"
	"golang.org/x/crypto/bcrypt"
)

func addAccount(c *gin.Context) {
	accountData := new(AccountRoute)
	if err := c.Bind(accountData); err != nil {
		c.String(http.StatusCreated, "Error getting data")
		return
	}
	if accountData.NAME == "" || accountData.PIN == "" {
		c.String(http.StatusBadRequest, "missing parameters")
		return
	}
	accountData.NAME = strings.ReplaceAll(accountData.NAME, " ", "")
	accountData.PIN = strings.ReplaceAll(accountData.PIN, " ", "")

	passrd, err := HashPassword(accountData.PIN)
	if err != nil {
		c.String(http.StatusCreated, "Error hashing password")
		return
	}

	accountCreationData := Account{Pin: passrd, Name: accountData.NAME, Balance: "0", Transactions: ""}
	id, err := createAccount(accountCreationData)
	if err != nil {
		c.String(http.StatusCreated, err.Error())
		return
	}
	id = id[len("user:"):]
	c.String(http.StatusOK, id)
}

func createAccount(user Account) (string, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Signin(map[string]interface{}{
		"user": "guffe",
		"pass": DATABASE_PASSWORD,
	}); err != nil {
		fmt.Println(err)
		return "", err
	}

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

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
