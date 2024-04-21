package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	e := echo.New()
	e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	e.Use(middleware.Recover())
	e.POST("/pay", pay)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "0")
	})
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	}
	defer f.Close()
	e.POST("/addAccount", addAccount)
	e.POST("/balanceCheck", checkBalance)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
		Output: f,
	}))
	e.Use(middleware.CORS())
	//e.Use(middleware.Secure())

	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == "test", nil
	}))
	if err := e.StartTLS(":8443", "fullchain.pem", "privkey.pem"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	// e.Logger.Fatal(e.Start(":1213"))
}

func pay(c echo.Context) error {
	paymentData := new(PaymentRoute)
	if err := c.Bind(paymentData); err != nil {
		return err
	}
	err := transferMoney(Transfer{From: "user:" + paymentData.Acc1, To: "user:" + paymentData.Acc2, Amount: paymentData.Amount, Pin: paymentData.Pin})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, "payment successful")
}
func addAccount(c echo.Context) error {
	accountData := new(AccountRoute)
	if err := c.Bind(accountData); err != nil {
		return c.String(http.StatusInternalServerError, "Error getting data")
	}
	passrd, err := HashPassword(accountData.PIN)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Error hashing password")
	}
	accountCreationData := Account{Pin: passrd, Name: accountData.NAME, Balance: "0"}
	id, err := createAccount(accountCreationData)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	id = id[len("user:"):]
	return c.String(http.StatusOK, id)
}
func checkBalance(c echo.Context) error {
	balanceData := new(BalanceRoute)
	if err := c.Bind(balanceData); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	balance, err := balanceCheck(BalanceCheck{ID: "user:" + balanceData.Acc1, Pin: balanceData.Pin})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, balance)
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
