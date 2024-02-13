package main

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
