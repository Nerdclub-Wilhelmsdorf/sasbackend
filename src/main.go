package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
)

const taxRate = 0.1
const taxFactor = 1.1

const DATABASE_PASSWORD = "IE76qzUk0t78JGhTz"

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
	e.POST("/getLogs", getLogs)
	e.POST("/verify", verfiy_account)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: currTime() + "method=${method}, uri=${uri}, status=${status}\n",
		Output: f,
	}))
	e.Use(middleware.CORS())
	e.Use(middleware.Secure())

	e.Use(middleware.KeyAuth(func(key string, c echo.Context) (bool, error) {
		return key == "test", nil
	}))

	if err := e.StartTLS(":8443", "fullchain.pem", "privkey.pem"); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	//e.Logger.Fatal(e.Start(":1213"))
}

func currTime() string {
	locat, error := time.LoadLocation("Europe/Berlin")
	var dt time.Time
	if error != nil {
		dt = time.Now()
		fmt.Println(error)
	} else {
		dt = time.Now().In(locat)
	}
	return dt.Format("01-02-2006 15:04:05")
}
