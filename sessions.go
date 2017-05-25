package main

import (
	"math/rand"
	"net/http"
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

// call this request on every html from javascript setInterval(func, 20000)
func autologout(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	_, ok := id_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------
}
