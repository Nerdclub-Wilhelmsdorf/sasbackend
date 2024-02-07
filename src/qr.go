package main

import qrcode "github.com/skip2/go-qrcode"

func createQr(acc string) {
	qrcode.WriteFile("m"+acc, qrcode.Medium, 256, acc+".png")
}
