package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

var failedAttempts = make(map[string]int)

func pay(c echo.Context) error {
	paymentData := new(Payment)
	if err := c.Bind(paymentData); err != nil {
		return err
	}
	if failedAttempts[paymentData.Acc1] > 3 {
		return c.String(http.StatusCreated, "suspended")
	}

	if hasDoc(paymentData.Acc1, "") && hasDoc(paymentData.Acc2, "") {
		res, err := readDocUnsafe(paymentData.Acc1, "")
		if err != nil {
			return c.String(http.StatusCreated, "Server Error")
		}
		if res["guest"] == "true" {
			return c.String(http.StatusCreated, "guests cant recieve money")
		}
		res, err = readDocUnsafe(paymentData.Acc1, "")
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
			f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return c.String(http.StatusCreated, "Failed")
			}
			f.WriteString(paymentData.Acc1 + " -> " + paymentData.Acc2 + ": " + paymentData.Amount + "D\n")
			f.Close()
			delete(failedAttempts, paymentData.Acc1)
			return c.String(http.StatusCreated, "success")
		}
		_, ok := failedAttempts[paymentData.Acc1]
		if !ok {
			failedAttempts[paymentData.Acc1] = 1
		}
		failedAttempts[paymentData.Acc1] += 1
		if failedAttempts[paymentData.Acc1] == 3 {
			go resetTimer(paymentData.Acc1)
			return c.String(http.StatusCreated, "suspended")
		}
		return c.String(http.StatusCreated, "wrong pin")
	}
	return c.String(http.StatusCreated, "Failed")
}
func resetTimer(acc string) {
	time.Sleep(5 * time.Minute)
	delete(failedAttempts, acc)
}
