package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/xyproto/randomstring"
)

func addAccount(c echo.Context) error {
	accountData := new(Account)
	if err := c.Bind(accountData); err != nil {
		return c.String(http.StatusCreated, "Server Error")
	}
	if len(accountData.PIN) != 4 {
		return c.String(http.StatusCreated, "Bad Pin")
	}
	if len(accountData.NAME) < 3 {
		return c.String(http.StatusCreated, "Bad Name")
	}
	pin_hash, err := HashPassword(accountData.PIN)
	if err != nil {
		return c.String(http.StatusCreated, "Server Error")
	}
	acc := map[string]string{
		"balance": "0",
		"pin":     pin_hash,
		"name":    accountData.NAME,
	}
	adress := randomstring.CookieFriendlyString(14)
	if hasKey(adress, "") {
		return c.String(http.StatusCreated, "already exists")
	}
	err = addDocUnsafe(acc, adress, "")
	if err != nil {
		fmt.Println(err)
		return c.String(http.StatusCreated, "couldnt add doc")
	}
	createQr(adress)
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return c.String(http.StatusCreated, "Failed")
	}
	f.WriteString("Account Created: " + adress + "\n")
	f.Close()

	return c.String(http.StatusCreated, "success: "+adress)
}

func balanceCheck(c echo.Context) error {
	balanceData := new(Balance)
	if err := c.Bind(balanceData); err != nil {
		return err
	}
	if failedAttempts[balanceData.Acc1] > 3 {
		return c.String(http.StatusCreated, "suspended")
	}
	if hasDoc(balanceData.Acc1, "") {
		res, err := readDocUnsafe(balanceData.Acc1, "")
		if err != nil {
			return c.String(401, "Server Error")
		}
		if CheckPasswordHash(balanceData.Pin, res["pin"]) {
			f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return c.String(http.StatusCreated, "Failed")
			}
			f.WriteString("Balance Checked: " + balanceData.Acc1 + "\n")
			f.Close()
			return c.String(http.StatusCreated, res["balance"])
		}
		_, ok := failedAttempts[balanceData.Acc1]
		if !ok {
			failedAttempts[balanceData.Acc1] = 1
		}
		failedAttempts[balanceData.Acc1] += 1
		if failedAttempts[balanceData.Acc1] == 3 {
			go resetTimer(balanceData.Acc1)
			return c.String(http.StatusCreated, "suspended")
		}

		return c.String(401, "wrong pin")
	}
	return c.String(401, "Failed")

}
