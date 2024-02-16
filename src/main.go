package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
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
	e.POST("/balanceCheck", balanceCheck)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
		Output: f,
	}))
	e.Use(middleware.CORS())
	//e.Use(middleware.Secure())
	e.Logger.Fatal(e.StartAutoTLS(":443"))
}
