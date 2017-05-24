package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
)

// referer map ( ["rawstream"] = "domain1.com;domain2.com" )
func listlocks(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
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

	table := "<tr><th>Stream</th><th>Domains</th><th>&nbsp;</th></tr>"
	dbgen_mu.RLock()
	query, err := dbgeneral.Query("SELECT id, streamname, referrers FROM referer WHERE username = ?", username)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}
	for query.Next() {
		var stream, refer string
		var id int
		query.Scan(&id, &stream, &refer)
		table += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td><button href='#' title='Press to change the status' onclick='load(%d)'>delete</button></td></tr>", stream, refer[0:80], id)
	}
	query.Close()
	fmt.Fprintf(w, "%s", table)
}

func add_referrer(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	r.ParseForm() // recupera campos del form tanto GET como POST
	dbgen_mu.Lock()
	_, err = dbgeneral.Exec("INSERT INTO players (`username`, `streamname`, `referrers`) VALUES (?, ?, ?)", username, r.FormValue("stream"), r.FormValue("domains"))
	dbgen_mu.Unlock()
	if err != nil {
		Error.Println(err)
		return
	}
	Referer.Store(username+"-"+r.FormValue("stream"), r.FormValue("domains"))
}

func delreferer(w http.ResponseWriter, r *http.Request) {
	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	r.ParseForm() // recupera campos del form tanto GET como POST
	dbgen_mu.Lock()
	_, err = dbgeneral.Exec("DELETE FROM users WHERE id = ? AND username = ?", r.FormValue("load"), username)
	dbgen_mu.Unlock()
	if err != nil {
		Error.Println(err)
		return
	}
	rawstream := "" // username + "-" + streamname
	Referer.Delete(rawstream)

}
