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

	table := "<table class=\"table table-hover table-condensed\"><thead class=\"bg-primary\"><tr class=\"row\"><th class=\"col-4\">Stream</th><th class=\"col-7\">Domains</th><th class=\"col-1\">&nbsp;</th></tr></thead><tbody>"
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
		table = table + fmt.Sprintf("<tr class=\"row\"><td class=\"col-4\">livestream</td><td class=\"col-7\">www.todostreaming.es</td><td class=\"col-1\"><button type=\"button\" class=\"btn btn-danger btn-xs\" title='Press to unlock' onclick='load(%d)'><span class=\"glyphicon glyphicon-lock\"></span></button></td></tr>", stream[0:11], refer[0:23], id)
	}
	query.Close()
	fmt.Fprintf(w, "%s", table+"</tbody></table></div>")
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

	r.ParseForm() // recupera campos del form tanto GET como POST

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	// revisamos antes que no existe una regla anterior y si existe pues solo actualizamos
	var id string
	dbgen_mu.RLock()
	err = dbgeneral.QueryRow("SELECT id FROM referer WHERE username = ? AND streamname = ?", username, r.FormValue("stream")).Scan(&id)
	dbgen_mu.RUnlock()
	if err == sql.ErrNoRows {
		dbgen_mu.Lock()
		_, err = dbgeneral.Exec("INSERT INTO players (`username`, `streamname`, `referrers`) VALUES (?, ?, ?)", username, r.FormValue("stream"), r.FormValue("domains"))
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
	} else if err != nil {
		Error.Println(err)
		return
	} else {
		dbgen_mu.Lock()
		_, err = dbgeneral.Exec("UPDATE players SET referrers = ? WHERE id = ?", r.FormValue("domains"), id)
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
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
	// comprobamos y extraemos el stream al que se refiere
	var streamname string
	dbgen_mu.RLock()
	err = dbgeneral.QueryRow("SELECT streamname FROM referer WHERE username = ? AND id = ?", username).Scan(&streamname)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}

	dbgen_mu.Lock()
	_, err = dbgeneral.Exec("DELETE FROM users WHERE id = ? AND username = ?", r.FormValue("load"), username)
	dbgen_mu.Unlock()
	if err != nil {
		Error.Println(err)
		return
	}
	Referer.Delete(username + "-" + streamname)
}
