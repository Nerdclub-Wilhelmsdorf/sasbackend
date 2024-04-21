package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"

	"github.com/surrealdb/surrealdb.go"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
	Pin     string `json:"pin"`
}

func main() {

	argsWithoutProg := os.Args[1:]
	arg := argsWithoutProg[0]
	switch arg {
	case "create":
		create()
	case "delete":
		delete()
	case "list":
		listAll()
	case "changepin":
		changepin()
	case "verify":
		verify()
	}

}

func create() {
	var id string
	var name string
	var balance string
	var pin string
	fmt.Println("Enter Account ID:")
	fmt.Scanln(&id)
	fmt.Println("Enter Student Number:")
	fmt.Scanln(&name)
	fmt.Println("Enter Balance:")
	fmt.Scanln(&balance)
	fmt.Println("Enter Pin:")
	fmt.Scanln(&pin)

	var user Account
	createAccountStruct(&user, name, balance, id, pin)
	id, err := createAccount(user)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("User created successfully with ID: " + id + " and Pin:" + user.Pin)
	}
}

func createAccountStruct(user *Account, name string, balance string, id string, pin string) {
	if id == "" {
		if pin == "" {
			userRef := user
			*userRef = Account{
				Name:    HashPassword(name),
				Balance: balance,
				Pin:     HashPassword(randomPin()),
			}
		} else {
			userRef := user
			*userRef = Account{
				Name:    HashPassword(name),
				Balance: balance,
				Pin:     pin,
			}
		}
	} else {
		if pin == "" {
			userRef := user
			*userRef = Account{
				ID:      id,
				Name:    HashPassword(name),
				Balance: balance,
				Pin:     HashPassword(randomPin()),
			}
		} else {
			userRef := user
			*userRef = Account{
				ID:      id,
				Name:    HashPassword(name),
				Balance: balance,
				Pin:     HashPassword(pin),
			}
		}
	}
}

func createAccount(user Account) (string, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return "", err
	}
	// Insert user
	data, err := db.Select(user.ID)
	selectedUser := new(Account)
	err = surrealdb.Unmarshal(data, &selectedUser)
	if err == nil {
		return "", errors.New("User with ID: " + user.ID + " already exists.")
	}
	if selectedUser.ID != "" {
		return "", errors.New("User with ID: " + user.ID + " already exists.")
	}
	data, err = db.Create("user", user)
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
	selectedUser = new(Account)
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

func randomPin() string {
	max := big.NewInt(10000)
	n, _ := rand.Int(rand.Reader, max)
	return fmt.Sprintf("%04d", n)
}

func listAll() {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		fmt.Println(err)
	}
	data, err := db.Query("SELECT id FROM user", map[string]interface{}{})
	if err != nil {
		panic(err)
	}
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}

	// Print the JSON data
	fmt.Println(string(jsonData))

}

func delete() {
	fmt.Println("Enter Account ID:")
	var id string
	fmt.Scanln(&id)
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		fmt.Println(err)
	}
	_, err := db.Delete("user:" + id)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("User with ID: " + id + " deleted successfully.")
	}
}

/*
	func usecsv() {
		file, err := os.Open("konten.csv")
		if err != nil {
			log.Fatal(err)
		}

}
*/
func changepin() {
	var id string
	var pin string

	var name string
	fmt.Println("Enter Account ID:")
	fmt.Scanln(&id)
	fmt.Println("Enter Student Number:")
	fmt.Scanln(&name)
	fmt.Println("Enter New Pin:")
	fmt.Scanln(&pin)
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		fmt.Println(err)
	}
	data, err := db.Select("user:" + id)
	if err != nil {
		fmt.Println(err)
	}
	selectedUser := new(Account)
	err = surrealdb.Unmarshal(data, &selectedUser)
	if err != nil {
		fmt.Println(err)
	}
	if CheckPasswordHash(name, selectedUser.Name) {
		fmt.Println("Verified")
	} else {
		fmt.Println("Error")
	}
	if _, err := db.Use("user", "user"); err != nil {
		fmt.Println(err)
	}
	changes := map[string]string{"pin": HashPassword(pin), "name": selectedUser.Name, "balance": selectedUser.Balance}
	if _, err = db.Update(selectedUser.ID, changes); err != nil {
		panic(err)
	}
}

func verify() {
	var id string
	var name string
	fmt.Println("Enter Account ID:")
	fmt.Scanln(&id)
	fmt.Println("Enter Student Number:")
	fmt.Scanln(&name)
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		fmt.Println(err)
	}
	data, err := db.Select("user:" + id)
	if err != nil {
		fmt.Println(err)
	}
	selectedUser := new(Account)
	err = surrealdb.Unmarshal(data, &selectedUser)
	if err != nil {
		fmt.Println(err)
	}
	if CheckPasswordHash(name, selectedUser.Name) {
		fmt.Println("Verified")
	} else {
		fmt.Println("Error")
	}
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func HashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}
