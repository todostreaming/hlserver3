package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"time"
)

func getMonthsYearsAdmin(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	tipo, ok := type_[key] // tipo int logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	//mu_user.RLock()
	//tipo := type_[key]
	//user := user_[key]
	//mu_user.RUnlock()
	// actualizamos la cookie actual
	expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
	newcookie := http.Cookie{Name: CookieName, Value: key, Expires: expiration}
	http.SetCookie(w, &newcookie)
	mu_user.Lock()
	time_[key] = expiration
	mu_user.Unlock()
	// ---- end of session identification -------------------------------

	var menu1, menu2, menu3 string

	yy, mm, _ := time.Now().Date() //Fecha actual
	for key, value := range YearMonths {
		if int(mm) == key+1 {
			menu1 += fmt.Sprintf("<option label='%s' value='%02d' selected>%s</option>", value, key+1, value)
		} else {
			menu1 += fmt.Sprintf("<option label='%s' value='%02d'>%s</option>", value, key+1, value)
		}
	}
	//Generamos el select de años
	for _, value := range UpDownYears(yy) {
		if int(yy) == value {
			menu2 += fmt.Sprintf("<option label='%d' value='%d' selected>%d</option>", value, value, value)
		} else {
			menu2 += fmt.Sprintf("<option label='%d' value='%d'>%d</option>", value, value, value)
		}
	}
	// Generamos los tipos clientes del logeado
	switch tipo {
	case 0:
		menu3 = "<option value='1'>Administrators</option><option value='2'>Distributors</option><option value='3'>Publishers</option>"
	case 1:
		menu3 = "<option value='2'>Distributors</option><option value='3'>Publishers</option>"
	case 2:
		menu3 = "<option value='3'>Publishers</option>"
	}

	fmt.Fprintf(w, "%s;%s", menu1, menu2, menu3)
}

// Funcion que muestra los datos mensuales de los clientes
func putMonthlyAdmin(w http.ResponseWriter, r *http.Request) {

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
	mu_user.RLock()
	tipo := type_[key]
	//user := user_[key]
	mu_user.RUnlock()
	// ---- end of session identification -------------------------------

	anio, mes, _ := time.Now().Date() //Fecha actual
	table := "<tr><th>Username</th><th>Password</th><th>Hours</th><th>GBs</th><th>Status</th></tr>"
	mesGrafico := fmt.Sprintf("%d-%02d", anio, mes)
	db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer db0.Close()
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	query2, err := dbgeneral.Query("SELECT id, username, password, status FROM users WHERE type = ?", tipo+1)
	if err != nil {
		Error.Println(err)
		return
	}
	for query2.Next() {
		var user, pass, estado string
		var id, status, minutos, megas int
		err = query2.Scan(&id, &user, &pass, &status)
		if status == 1 {
			estado = "ON"
		} else {
			estado = "OFF"
		}
		dbmon_mu.RLock()
		query, err := db0.Query("SELECT sum(hours), sum(gigabytes) FROM resumen WHERE username = ? GROUP BY username", user)
		dbmon_mu.RUnlock()
		if err != nil {
			Error.Println(err)
		}
		for query.Next() {
			err = query.Scan(&minutos, &megas)
			if err != nil {
				Warning.Println(err)
			}
		}
		query.Close()
		table += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td><button href='#' title='Press to change the status' onclick='load(%d)'>%s</button></td></tr>", user, pass, minutos, megas, id, estado)
	}
	query2.Close()
	fmt.Fprintf(w, "%s", table)
}
