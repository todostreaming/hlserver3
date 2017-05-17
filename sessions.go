package main

import (
	"math/rand"
	"time"
)

// control sessions from the server side
func controlinternalsessions() {
	for {
		for k, v := range time_ {
			if time.Since(v).Seconds() > 0 { // it is negative up to expiration time
				mu_user.Lock()
				delete(user_, k)
				delete(time_, k)
				delete(type_, k)
				mu_user.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// generates a session id or Value for a random Cookie with a fixed length
func sessionid(r *rand.Rand, n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}
