package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
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
	tipo, ok := type_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	anio, mes, _ := time.Now().Date() //Fecha actual
	table := "<tr><th>Username</th><th>Password</th><th>Hours</th><th>GBs</th><th>Status</th></tr>"
	mesGrafico := fmt.Sprintf("%d-%02d", anio, mes)

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	dbgen_mu.RLock()
	query, err := dbgeneral.Query("SELECT id, username, password, status FROM users WHERE type = ?", tipo+1)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}
	for query.Next() {
		var user, pass, estado string
		var id, status, horas, gigas int
		query.Scan(&id, &user, &pass, &status)
		if status == 1 {
			estado = "ON"
		} else {
			estado = "OFF"
		}
		horas, gigas = sumaconsumo(id, tipo+1, mesGrafico)
		table += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td><button href='#' title='Press to change the status' onclick='load(%d)'>%s</button></td></tr>", user, pass, horas, gigas, id, estado)
	}
	query.Close()
	fmt.Fprintf(w, "%s", table)
}

// Funcion que muestra los datos mensuales de los clientes
func putMonthlyAdminChange(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	_, ok := id_[key] // tipo int logeado
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

	r.ParseForm()
	mesGrafico := r.FormValue("years") + "-" + r.FormValue("months")
	tipo := toInt(r.FormValue("types"))

	table := "<tr><th>Username</th><th>Password</th><th>Hours</th><th>GBs</th><th>Status</th></tr>"

	if _, err := os.Stat(dirMonthlys + mesGrafico + "monthly.db"); os.IsNotExist(err) {
		//No hay base de datos
		Warning.Println("No existe el fichero de base de datos")
		Error.Println(err)
		fmt.Fprintf(w, "%s", "NoBD")
		return
	}
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	dbgen_mu.RLock()
	query, err := dbgeneral.Query("SELECT id, username, password, status FROM users WHERE type = ?", tipo)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}
	for query.Next() {
		var user, pass, estado string
		var id, status, horas, gigas int
		query.Scan(&id, &user, &pass, &status)
		if status == 1 {
			estado = "ON"
		} else {
			estado = "OFF"
		}
		horas, gigas = sumaconsumo(id, tipo, mesGrafico)
		table += fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td><button href='#' title='Press to change the status' onclick='load(%d)'>%s</button></td></tr>", user, pass, horas, gigas, id, estado)
	}
	query.Close()
	fmt.Fprintf(w, "%s", table)

}

// funcion que suma el consumo de todos los publishers dependientes del usuario id
// en el mes grafico yyyy-mm
func sumaconsumo(id, tipo int, mesgrafico string) (horas, gigas int) {
	publisher := []string{} // aqui metemos todos los publishers

	// SELECT sum(hours), sum(gigagabytes) FROM resumen WHERE username = ? OR username = .........
	// vamos a ir tirando del hilo y sacar todos los publishers en un array
	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	if tipo == 1 { // admin
		dbgen_mu.RLock()
		query, err := dbgeneral.Query("SELECT id FROM users WHERE id_recruiter = ?", id)
		dbgen_mu.RUnlock()
		if err != nil {
			Error.Println(err)
			return
		}
		for query.Next() {
			var iddistro int
			query.Scan(&iddistro)
			dbgen_mu.RLock()
			query2, err := dbgeneral.Query("SELECT username FROM users WHERE id_recruiter = ?", iddistro)
			dbgen_mu.RUnlock()
			if err != nil {
				Error.Println(err)
				return
			}
			for query2.Next() {
				var pub string
				query2.Scan(&pub)
				publisher = append(publisher, pub)
			}
			query2.Close()
		}
		query.Close()
	} else if tipo == 2 { // distro
		dbgen_mu.RLock()
		query, err := dbgeneral.Query("SELECT username FROM users WHERE id_recruiter = ?", id)
		dbgen_mu.RUnlock()
		if err != nil {
			Error.Println(err)
			return
		}
		for query.Next() {
			var pub string
			query.Scan(&pub)
			publisher = append(publisher, pub)
		}
		query.Close()
	} else { // publisher
		var pub string
		dbgen_mu.RLock()
		err := dbgeneral.QueryRow("SELECT username FROM users WHERE id = ?", id).Scan(&pub)
		dbgen_mu.RUnlock()
		if err != nil {
			Error.Println(err)
			return
		}
		publisher = append(publisher, pub)
	}

	qstr := "SELECT sum(hours), sum(gigabytes) FROM resumen WHERE" // query string to build ( username = 'jjj' OR username = 'pppp')
	length := len(publisher)
	for k, v := range publisher {
		if (k + 1) == length {
			qstr = qstr + fmt.Sprintf(" username = '%s'", v)
		} else {
			qstr = qstr + fmt.Sprintf(" username = '%s' OR", v)
		}
	}

	db0, err := sql.Open("sqlite3", dirMonthlys+mesgrafico+"monthly.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer db0.Close()
	dbmon_mu.RLock()
	err = db0.QueryRow(qstr).Scan(&horas, &gigas)
	dbmon_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}

	return
}

// Funcion cambia el estado ON/OFF a los clientes
func changeStatus(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	_, ok := id_[key]
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	r.ParseForm()
	var id, status int
	var user string

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db")
	if err != nil {
		Error.Println(err)
		return
	}
	defer dbgeneral.Close()

	dbgen_mu.RLock()
	err = dbgeneral.QueryRow("SELECT id, username, status FROM users WHERE id = ?", r.FormValue("load")).Scan(&id, &user, &status)
	dbgen_mu.RUnlock()
	if err != nil {
		Error.Println(err)
		return
	}

	if status == 1 { // active (must be disabled)
		dbgen_mu.Lock()
		_, err := dbgeneral.Exec("UPDATE users SET status = 0 WHERE id = ?", id)
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
			return
		}
		time.Sleep(10 * time.Millisecond)
		// Seleccionamos todos los streams pertenecientes a un usuario, para hecharlos fuera
		query, err := dblive.Query("SELECT streamname FROM encoders WHERE username = ?", user)
		if err != nil {
			Error.Println(err)
			return
		}
		for query.Next() {
			var streams string
			query.Scan(&streams)
			nombre := fmt.Sprintf("%s-%s", user, streams)
			//Sacamos uno a uno los streams
			peticion := fmt.Sprintf("http://127.0.0.1:8080/control/drop/publisher?app=live&name=%s", nombre)
			http.Get(peticion)
			time.Sleep(10 * time.Millisecond)
		}
		query.Close()
	} else {
		dbgen_mu.Lock()
		_, err := dbgeneral.Exec("UPDATE users SET status = 1 WHERE id= ?", id)
		dbgen_mu.Unlock()
		if err != nil {
			Error.Println(err)
		}
	}
}
