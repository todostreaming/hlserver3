package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
)

// Funcion para dar de alta clientes
func nuevoCliente(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	id, ok := id_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	//mu_user.RLock()
	//tipo := type_[key]
	//user := user_[key]
	//mu_user.RUnlock()
	// ---- end of session identification -------------------------------

	resp := "BAD"
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		fmt.Fprintf(w, "%s", resp)
		return
	}
	defer dbgeneral.Close()
	var count int
	dbgen_mu.RLock()
	err = dbgeneral.QueryRow("SELECT count(id) FROM users WHERE username = ", r.FormValue("nom_cli")).Scan(&count)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		fmt.Fprintf(w, "%s", resp)
		return
	}
	if count > 0 { // username already exists
		resp = "DUP"
		Error.Println(err)
		fmt.Fprintf(w, "%s", resp)
		return
	}

	dbgen_mu.Lock()
	_, err1 := dblive.Exec("INSERT INTO users (`username`, `password`, `pubpass`, `type`, `status`, `id_recruiter`) VALUES (?, ?, ?, ?, ?)",
		r.FormValue("nom_cli"), r.FormValue("passw"), r.FormValue("passw"), r.FormValue("type"), r.FormValue("status"), id)
	dbgen_mu.Unlock()
	if err1 != nil {
		Error.Println(err)
		fmt.Fprintf(w, "%s", resp)
		return
	}
	resp = "OK"
	fmt.Fprintf(w, "%s", resp)
}

// Función para crear las opciones de tipos
func types(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	tipo, ok := type_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	options := ""
	switch tipo {
	case 0: // superadmin
		options = "<option value=\"1\">Admin</option><option value=\"2\">Distributor</option><option value=\"3\">Publisher</option>"
	case 1: // admin
		options = "<option value=\"2\">Distributor</option><option value=\"3\">Publisher</option>"
	case 2: // distro
		options = "<option value=\"3\">Publisher</option>"
	}
	fmt.Fprintf(w, "%s", options)
}

// Función que selecciona los clientes de la tabla admin
func buscarClientes(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	tipo, ok := type_[key]
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	mu_user.RLock()
	id := type_[key]
	//user := user_[key]
	mu_user.RUnlock()
	// ---- end of session identification -------------------------------

	var idn int
	var nombre, selector string
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		fmt.Fprintf(w, "%s", selector)
		return
	}
	defer dbgeneral.Close()
	dbgen_mu.RLock()
	query, err := dblive.Query("SELECT id, username FROM users WHERE type = ? AND id_recruiter = ?", tipo+1, id)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		fmt.Fprintf(w, "%s", selector)
		return
	}
	for query.Next() {
		err = query.Scan(&idn, &nombre)
		if err != nil {
			Warning.Println(err)
			fmt.Fprintf(w, "%s", selector)
			return
		}
		selector = selector + fmt.Sprintf("<option value='%d'>%s</option>", idn, nombre)
	}
	query.Close()
	fmt.Fprintf(w, "%s", selector)
}

// Función que borra un cliente de la tabla admin
func borrarCliente(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	tipo, ok := type_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	//mu_user.RLock()
	//tipo := type_[key]
	//user := user_[key]
	//mu_user.RUnlock()
	// ---- end of session identification -------------------------------

	r.ParseForm()

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	// debemos revisar si quien queremos borrar tiene usuarios a su vez (excepto si es un distro que borra directamente al publisher)
	if tipo == 2 { // distro
		dbgen_mu.Lock()
		_, err = dbgeneral.Exec("DELETE FROM users WHERE id = ?", r.FormValue("clients"))
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
	} else { // admin o superadmin
		var count int
		dbgen_mu.RLock()
		err = dbgeneral.QueryRow("SELECT count(id) FROM users WHERE id_recruiter = ?", r.FormValue("clients")).Scan(&count)
		dbgen_mu.RUnlock()
		if err != nil {
			Error.Println(err)
			return
		}
		if count > 0 { // tiene clientes dependientes (no borrar)
			// do nothing
		} else { // no tiene clientes, se le borra
			dbgen_mu.Lock()
			_, err = dbgeneral.Exec("DELETE FROM users WHERE id = ?", r.FormValue("clients"))
			dbgen_mu.Unlock()
			if err != nil {
				Error.Println(err)
				return
			}
		}
	}
}
