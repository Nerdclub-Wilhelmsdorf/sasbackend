package main

import (
	"fmt"
	"net/http"

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
	return c.String(http.StatusCreated, "success: "+adress)
}

func balanceCheck(c echo.Context) error {
	balanceData := new(Balance)
	if err := c.Bind(balanceData); err != nil {
		return err
	}
	if hasDoc(balanceData.Acc1, "") {
		res, err := readDocUnsafe(balanceData.Acc1, "")
		if err != nil {
			return c.String(401, "Server Error")
		}
		if CheckPasswordHash(balanceData.Pin, res["pin"]) {
			return c.String(http.StatusCreated, res["balance"])
		}
		return c.String(401, "wrong pin")
	}
	return c.String(401, "Failed")

}
