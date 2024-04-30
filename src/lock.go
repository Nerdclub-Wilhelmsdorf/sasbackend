package main

import (
	"time"
)

// map of users with name and bool
var failedAttempts = make(map[string]int)

func resetTimer(pin string) {
	time.Sleep(5 * time.Minute)
	delete(failedAttempts, pin)
}
