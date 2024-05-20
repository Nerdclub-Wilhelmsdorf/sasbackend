package main

import (
	"fmt"
	"slices"

	"github.com/shopspring/decimal"
	"github.com/surrealdb/surrealdb.go"
)

// TODO:
/*
Schrittfolge:
1. Konto 1 (Sender) aus der Datenbank in den Account Struct laden DONE
2. Konto 2 (Empfänger) aus der Datenbank in den Account Struct laden DONE
3. Zentralbank laden DONE
(vielleicht für 4-6 eine Validationsfunktion) was soll die machen? einfach 4 bis 6 ja i guess macht sinn
4. Checken, ob die PIN von K1 richtig ist	DONE
5. Checken, ob K1 gesperrt ist DONE
6. Checken, ob K1 genug Geld hat DONE
7. Checken, ob K2 existiert
8. Wenn alles passt, Bezahlte summe * 1.1 von K1 abziehen
9. Bezahlte Summe K2 aufaddieren
10. Bezahlte Summe * 0.1 auf die Zentralbank schreiben

Daten an Logging funktion übergeben:
	11. Die Transaktionen in den jeweiligen Accounts loggen
	12. alles in CSV loggen
*/
func transferMoney_2(transfer Transfer) error {
	from, err := loadUser(transfer.From)
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	//to, err := loadUser(transfer.To)
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	//bank, err := loadUser("zentralbank")
	if err != nil {
		return fmt.Errorf("failed to load user with ID %s: %w", transfer.From, err)
	}
	validateTransaction(from, transfer)
	if !CheckPasswordHash(transfer.Pin, from.Pin) {
		return fmt.Errorf("wrong pin")
	}
	return nil
}

func logger() {

}
func loadUser(id string) (Account, error) {
	db, _ := surrealdb.New("ws://localhost:8000/rpc")
	defer db.Close()
	if _, err := db.Use("user", "user"); err != nil {
		return Account{}, fmt.Errorf("failed to use database: %w", err)
	}
	data, err := db.Select(id)
	if err != nil {
		return Account{}, fmt.Errorf("failed to select account with ID %s: %w", id, err)
	}
	acc := new(Account)
	err = surrealdb.Unmarshal(data, &acc)
	if err != nil {
		return Account{}, fmt.Errorf("failed to unmarshal account data: %w", err)
	}
	return *acc, nil
}

func validateTransaction(payer Account, transfer Transfer) error {
	if !CheckPasswordHash(transfer.Pin, payer.Pin) {
		_, ok := failedAttempts[transfer.From]
		if !ok {
			failedAttempts[transfer.From] = 1
		}
		failedAttempts[transfer.From] += 1
		if failedAttempts[transfer.From] > 3 {
			if !slices.Contains(failedAttemtsCurrentlyLocking, transfer.From) {
				go resetTimer(transfer.From)
			}
			return fmt.Errorf("suspended")
		}
		return fmt.Errorf("wrong pin")
	}
	if failedAttempts[transfer.From] > 3 {
		return fmt.Errorf("suspended")
	}
	payerDecimal, err := decimal.NewFromString(payer.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance to decimal: %w", err)
	}
	transferDecimal, err := decimal.NewFromString(transfer.Amount)
	if err != nil {
		return fmt.Errorf("failed to convert transfer amount to decimal: %w", err)
	}
	if transferDecimal.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("bad amount")
	}
	if payerDecimal.Mul(decimal.NewFromFloat(taxFactor)).LessThan(transferDecimal) {
		return fmt.Errorf("insufficient funds")
	}
	return nil
}
