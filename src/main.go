package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/xyproto/randomstring"
)

type Payment struct {
	Acc1   string `json:"acc1" xml:"acc1" form:"acc1" query:"acc1"`
	Pin    string `json:"pin" xml:"pin" form:"pin" query:"pin"`
	Amount string `json:"amount" xml:"amount" form:"amount" query:"amount"`
	Acc2   string `json:"acc2" xml:"acc2" form:"acc2" query:"acc2"`
}
type Account struct {
	PIN  string `json:"pin" xml:"pin" form:"pin" query:"pin"`
	NAME string `json:"name" xml:"name" form:"name" query:"name"`
}

type Balance struct {
	Acc1 string `json:"acc1" xml:"acc1" form:"acc1" query:"acc1"`
	Pin  string `json:"pin" xml:"pin" form:"pin" query:"pin"`
}

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
		if res["pin"] == paymentData.Pin {
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
	acc := map[string]string{
		"balance": "0",
		"pin":     accountData.PIN,
		"name":    accountData.NAME,
	}
	adress := randomstring.CookieFriendlyString(14)
	if hasKey(adress, "") {
		return c.String(http.StatusCreated, "already exists")
	}
	err := addDocUnsafe(acc, adress, "")
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
			return c.String(http.StatusCreated, "Server Error")
		}
		if res["pin"] == balanceData.Pin {
			return c.String(http.StatusCreated, res["balance"])
		}
		return c.String(http.StatusCreated, "wrong pin")
	}
	return c.String(http.StatusCreated, "Failed")

}
