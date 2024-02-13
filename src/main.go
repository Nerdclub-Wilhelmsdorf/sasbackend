package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.POST("/pay", pay)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "0")
	})
	e.POST("/addAccount", addAccount)
	e.GET("/balanceCheck", balanceCheck)
	e.Logger.Fatal(e.Start(":1323"))
}
