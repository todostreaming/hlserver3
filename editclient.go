package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
)

//Funcion para editar los datos del admin
func editar_cliente(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	user, ok := user_[key]
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// actualizamos la cookie actual
	expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
	newcookie := http.Cookie{Name: CookieName, Value: key, Expires: expiration}
	http.SetCookie(w, &newcookie)
	mu_user.Lock()
	time_[key] = expiration
	mu_user.Unlock()
	// ---- end of session identification -------------------------------

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	r.ParseForm() // recupera campos del form tanto GET como POST
	//Solo si las contraseñas son iguales modificamos
	if r.FormValue("password") == r.FormValue("repeat-password") {
		good := "Password correctly changed"
		dbgen_mu.Lock()
		_, err := dbgeneral.Exec("UPDATE users SET password = ? WHERE username = ?", r.FormValue("password"), user)
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
		fmt.Fprintf(w, "<div class=\"text-success\"><strong>%s</strong></div>", good)
	} else {
		bad := "Passwords do not coincide"
		fmt.Fprintf(w, "<div class=\"text-danger\"><strong>%s</strong></div>", bad)
	}
}

//Funcion para editar los datos del admin
func editar_publish(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	user, ok := user_[key]
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// actualizamos la cookie actual
	expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
	newcookie := http.Cookie{Name: CookieName, Value: key, Expires: expiration}
	http.SetCookie(w, &newcookie)
	mu_user.Lock()
	time_[key] = expiration
	mu_user.Unlock()
	// ---- end of session identification -------------------------------

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	r.ParseForm() // recupera campos del form tanto GET como POST
	//Solo si las contraseñas son iguales modificamos
	if r.FormValue("password") == r.FormValue("repeat-password") {
		good := "Password correctly changed"
		dbgen_mu.Lock()
		_, err := dbgeneral.Exec("UPDATE users SET pubpass = ? WHERE username = ?", r.FormValue("password"), user)
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
		fmt.Fprintf(w, "<div class=\"text-success\"><strong>%s</strong></div>", good)
	} else {
		bad := "Passwords do not coincide"
		fmt.Fprintf(w, "<div class=\"text-danger\"><strong>%s</strong></div>", bad)
	}
}

//Función que muestra el usuario en activo
func user_admin(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	user, ok := user_[key]
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// actualizamos la cookie actual
	expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
	newcookie := http.Cookie{Name: CookieName, Value: key, Expires: expiration}
	http.SetCookie(w, &newcookie)
	mu_user.Lock()
	time_[key] = expiration
	mu_user.Unlock()
	// ---- end of session identification -------------------------------

	fmt.Fprintf(w, "<span class=\"input-group-addon\"><i class=\"glyphicon glyphicon-user\"></i></span><input class='form-control' placeholder='Usuario' readonly='readonly' name='username' type='username' value='%s' autofocus>", user)
}

//Funcion para editar los datos del admin
func username(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return
	}
	key := cookie.Value
	mu_user.RLock()
	user, ok := user_[key]
	mu_user.RUnlock()
	if !ok {
		return
	}
	// ---- end of session identification -------------------------------

	fmt.Fprintf(w, user)
}
