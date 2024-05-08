package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/surrealdb/surrealdb.go"
)

type AccountState int

const (
	Suspended AccountState = iota
	Verified
	DoesNotExist
	AccountStateError
	FailedVerification
)

func verfiy_account(c echo.Context) error {
	accData := new(AccountRoute)
	if err := c.Bind(accData); err != nil {
		return c.String(http.StatusTeapot, err.Error())
	}
	if accData.NAME == "" || accData.PIN == "" {
		return c.String(http.StatusBadRequest, "missing parameters")
	}
	accData.NAME = strings.ReplaceAll(accData.NAME, " ", "")
	accData.PIN = strings.ReplaceAll(accData.PIN, " ", "")
	acc, _ := verifyAccount("user:"+accData.NAME, accData.PIN)
	switch acc {
	case Suspended:
		return c.String(http.StatusOK, "account suspended")
	case Verified:
		return c.String(http.StatusOK, "account verified")
	case DoesNotExist:
		return c.String(http.StatusOK, "account does not exist")
	case AccountStateError:
		return c.String(http.StatusOK, "error verifying account")
	case FailedVerification:
		return c.String(http.StatusOK, "failed to verify account")
	default:
		return c.String(http.StatusOK, "server error")
	}
}

func verifyAccount(ID string, PIN string) (AccountState, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return AccountStateError, fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(ID)
	if err != nil {
		return DoesNotExist, fmt.Errorf("failed to select account with ID %s: %w", ID, err)
	}
	acc1 := new(Account)
	err = surrealdb.Unmarshal(data, &acc1)
	if err != nil {
		return AccountStateError, fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	if failedAttempts[ID] > 3 {
		return Suspended, nil
	}

	if !CheckPasswordHash(PIN, acc1.Pin) {
		_, ok := failedAttempts[ID]
		if !ok {
			failedAttempts[ID] = 1
		}
		failedAttempts[ID] += 1
		if failedAttempts[ID] == 3 {
			go resetTimer(ID)
			return Suspended, nil
		}

		return FailedVerification, errors.New("incorrect pin")

	}
	return Verified, nil
}
