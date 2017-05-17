package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"math/rand"
	"net/http"
	"time"
)

// login control code
func login(w http.ResponseWriter, r *http.Request) {
	// get user,pass from the login form at index.html
	r.ParseForm() // recupera campos del form tanto GET como POST
	user := r.FormValue(name_username)
	pass := r.FormValue(name_password)

	var username, password string
	var tipo, status int
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db") // Apertura de la dateDayly.db antigua para lectura del pico/hora
	if err != nil {
		Error.Println(err)
		// go back to the login form page
		http.Redirect(w, r, "/"+first_page+".html?err", http.StatusFound)
		return
	}
	defer dbgeneral.Close()
	dbgen_mu.RLock()
	err = dbgeneral.QueryRow("SELECT username, password, type, status FROM users WHERE username = ? AND password = ?", user, pass).Scan(&username, &password, &tipo, &status)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		// go back to the login form page
		http.Redirect(w, r, "/"+first_page+".html?err", http.StatusFound)
		return
	}
	if (username == user) && (password == pass) {
		// generate the cookie with the session value
		aleat := rand.New(rand.NewSource(time.Now().UnixNano()))
		sid := sessionid(aleat, session_value_len)
		expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
		cookie := http.Cookie{Name: CookieName, Value: sid, Expires: expiration}
		http.SetCookie(w, &cookie)

		if tipo < 2 { // superadmin or admin user
			mu_user.Lock()
			user_[sid] = username
			time_[sid] = expiration
			type_[sid] = tipo
			mu_user.Unlock()
			// Send you to the 1st admin's page
			http.Redirect(w, r, "/"+enter_page_admin, http.StatusFound)
			return
		} else { // publisher user
			mu_user.Lock()
			user_[sid] = username
			time_[sid] = expiration
			type_[sid] = tipo
			mu_user.Unlock()
			// Send you to the 1st publisher's page
			http.Redirect(w, r, "/"+enter_page, http.StatusFound)
			return
		}
	} else {
		// go back to the login form page
		http.Redirect(w, r, "/"+first_page+".html?err", http.StatusFound)
		return
	}
}

// logout control code
func logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(CookieName)

	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
	} else {
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
		mu_user.Lock()
		delete(user_, cookie.Value)
		delete(time_, cookie.Value)
		delete(type_, cookie.Value)
		mu_user.Unlock()

		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
	}

}
