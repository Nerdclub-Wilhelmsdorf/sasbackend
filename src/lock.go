package main

import (
	"slices"
	"time"
)

// map of users with name and bool
var failedAttempts = make(map[string]int)
var failedAttemtsCurrentlyLocking = []string{}

func resetTimer(name string) {
	failedAttemtsCurrentlyLocking = append(failedAttemtsCurrentlyLocking, name)
	time.Sleep(5 * time.Minute)
	delete(failedAttempts, name)
	failedAttemtsCurrentlyLocking = deleteElement(failedAttemtsCurrentlyLocking, slices.Index(failedAttemtsCurrentlyLocking, name))

}

func deleteElement(slice []string, index int) []string {
	return append(slice[:index], slice[index+1:]...)
}
