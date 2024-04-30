package main

type PaymentRoute struct {
	Acc1   string `json:"acc1" xml:"acc1" form:"acc1" query:"acc1"`
	Pin    string `json:"pin" xml:"pin" form:"pin" query:"pin"`
	Amount string `json:"amount" xml:"amount" form:"amount" query:"amount"`
	Acc2   string `json:"acc2" xml:"acc2" form:"acc2" query:"acc2"`
}
type AccountRoute struct {
	PIN  string `json:"pin" xml:"pin" form:"pin" query:"pin"`
	NAME string `json:"name" xml:"name" form:"name" query:"name"`
}

type BalanceRoute struct {
	Acc1 string `json:"acc1" xml:"acc1" form:"acc1" query:"acc1"`
	Pin  string `json:"pin" xml:"pin" form:"pin" query:"pin"`
}

type Account struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Balance string `json:"balance"`
	Pin     string `json:"pin"`
}

type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Pin    string `json:"pin"`
}

type BalanceCheck struct {
	ID  string `json:"id"`
	Pin string `json:"pin"`
}

type TransactionLog struct {
	Time   string `json:"time"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type StoreTransactions struct {
	Transactions []TransactionLog
}
