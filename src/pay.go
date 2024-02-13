package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func pay(c echo.Context) error {
	paymentData := new(Payment)
	if err := c.Bind(paymentData); err != nil {
		return err
	}
	if hasDoc(paymentData.Acc1, "") && hasDoc(paymentData.Acc2, "") {
		res, err := readDocUnsafe(paymentData.Acc1, "")
		if err != nil {
			return c.String(http.StatusCreated, "Server Error")
		}
		if CheckPasswordHash(paymentData.Pin, res["pin"]) {
			amount, err := strconv.ParseFloat(paymentData.Amount, 64)
			if err != nil {
				return c.String(http.StatusCreated, "Server Error")
			}
			balance, err := strconv.ParseFloat(res["balance"], 64)
			if err != nil {
				return c.String(http.StatusCreated, "Server Error")
			}
			if balance >= amount*1.1 {
				balance -= amount * 1.1
			} else {
				return c.String(http.StatusCreated, "Not enough money")
			}
			addDocUnsafe(map[string]string{"balance": fmt.Sprintf("%f", balance), "pin": res["pin"]}, paymentData.Acc1, "")
			res, err = readDocUnsafe(paymentData.Acc2, "")
			if err != nil {
				return c.String(http.StatusCreated, "Server Error")
			}
			balance, err = strconv.ParseFloat(res["balance"], 64)
			if err != nil {
				return c.String(http.StatusCreated, "Server Error")
			}

			balance += amount
			addDocUnsafe(map[string]string{"balance": fmt.Sprintf("%f", balance), "pin": res["pin"]}, paymentData.Acc2, "")
			//add 10& to the bank
			res, _ = readDocUnsafe("zentralbank", "")
			balance, _ = strconv.ParseFloat(res["balance"], 64)
			balance += amount * 0.1
			addDocUnsafe(map[string]string{"balance": fmt.Sprintf("%f", balance), "pin": res["pin"]}, "zentralbank", "")
			return c.String(http.StatusCreated, "success")
		}
		return c.String(http.StatusCreated, "wrong pin")
	}
	return c.String(http.StatusCreated, "Failed")
}
