package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"os"
	"strings"
	"time"
)

type Grafico2 struct {
	Type   string `json:"type"`
	Data   []int  `json:"data"`
	Labels []int  `json:"labels"`
}
type Grafico3 struct {
	Type   string    `json:"type"`
	Data   []float64 `json:"data"`
	Labels []int     `json:"labels"`
}

func getMonthsYears(w http.ResponseWriter, r *http.Request) {

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
	var menu1, menu2, menu3 string
	anio, mm, _ := time.Now().Date() //Fecha actual
	for key, value := range YearMonths {

		if int(mm) == key+1 {
			menu1 += fmt.Sprintf("<option label='%s' value='%02d' selected>%s</option>", value, key+1, value)
		} else {
			menu1 += fmt.Sprintf("<option label='%s' value='%02d'>%s</option>", value, key+1, value)
		}
	}
	//Generamos el select de años
	for _, value := range UpDownYears(anio) {
		if int(anio) == value {
			menu2 += fmt.Sprintf("<option label='%d' value='%d' selected>%d</option>", value, value, value)
		} else {
			menu2 += fmt.Sprintf("<option label='%d' value='%d'>%d</option>", value, value, value)
		}
	}
	menu3 = "<option label='todo' value='todo'>All</option>"
	fmt.Fprintf(w, "%s;%s;%s", menu1, menu2, menu3)
}

func firstMonthly(w http.ResponseWriter, r *http.Request) {

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

	var fechaAud []int
	var menu3 string
	var arrAud map[int]int = make(map[int]int)
	var arrMin map[int]int = make(map[int]int)
	var arrAVG map[int]float64 = make(map[int]float64)
	var arrMegas map[int]int = make(map[int]int)
	var arrPico map[int]int = make(map[int]int)
	var horaPico map[int]int = make(map[int]int)
	anio, mes, _ := time.Now().Date() //Fecha actual
	mesGrafico := fmt.Sprintf("%d-%02d", anio, mes)
	db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
	if err != nil {
		Error.Println(err)
		return
	}
	//Generamos el select de streams
	dbmon_mu.RLock()
	query2, err := db0.Query("SELECT  DISTINCT(streamname) FROM resumen WHERE username = ?", username)
	dbmon_mu.RUnlock()
	if err != nil {
		Error.Println(err)
	}
	if !query2.Next() {
		//No, campos vacios
		menu3 = "<option label='All' value='todo' selected>All</option>"
	} else {
		//Si, existen campos. Formamos el select
		menu3 = "<option label='All' value='todo' selected>All</option>"
		dbmon_mu.RLock()
		query3, err := db0.Query("SELECT  DISTINCT(streamname) FROM resumen WHERE username = ?", username)
		dbmon_mu.RUnlock()
		if err != nil {
			Error.Println(err)
			return
		}
		for query3.Next() {
			var stream string
			query3.Scan(&stream)
			menu3 += "<option label='" + stream + "' value='" + stream + "'>" + stream + "</option>"
		}
		query3.Close()
	}
	query2.Close()
	// Se añaden los dia del mes al grafico
	for i := 1; i <= daysIn(mes, anio); i++ {
		fechaAud = append(fechaAud, i)
	}
	dbmon_mu.RLock()
	query, err := db0.Query("SELECT sum(players), sum(hours), avg(hours), sum(gigabytes), max(peak), peaktime, date FROM resumen WHERE username = ? GROUP BY date", username)
	dbmon_mu.RUnlock()
	if err != nil {
		Error.Println(err)
	}
	for query.Next() {
		var audiencia, minutos, megas, pico int
		var promedio float64
		var fecha, horapico string
		err = query.Scan(&audiencia, &minutos, &promedio, &megas, &pico, &horapico, &fecha)
		if err != nil {
			Warning.Println(err)
		}
		hour := strings.Split(horapico, ":")
		day := strings.Split(fecha, ":")
		arrAud[toInt(day[1])] = audiencia
		arrMin[toInt(day[1])] = minutos
		arrAVG[toInt(day[1])] = promedio
		arrMegas[toInt(day[1])] = megas
		arrPico[toInt(day[1])] = pico
		horaPico[toInt(day[1])] = toInt(hour[0])
	}
	query.Close()
	g1 := grafDays(arrAud, len(fechaAud))
	g2 := grafDays(arrMin, len(fechaAud))
	g3 := grafDaysFloat(arrAVG, len(fechaAud))
	g4 := grafDays(arrMegas, len(fechaAud))
	g5 := grafDays(arrPico, len(fechaAud))
	g6 := grafDays(horaPico, len(fechaAud))
	// Aquí se crean los JSON
	grafico1, _ := json.Marshal(Grafico2{"bar", g1, fechaAud})  // Aquí se crea el JSON para el grafico de audiencia total del dia en personas
	grafico2, _ := json.Marshal(Grafico2{"bar", g2, fechaAud})  // Aquí se crea el JSON para el grafico de tiempo total visionado
	grafico3, _ := json.Marshal(Grafico3{"bar", g3, fechaAud})  // Aquí se crea el JSON para el grafico de segundos consumidos por pais
	grafico4, _ := json.Marshal(Grafico2{"bar", g4, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por pais
	grafico5, _ := json.Marshal(Grafico2{"bar", g5, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por franja horaria
	grafico6, _ := json.Marshal(Grafico2{"line", g6, fechaAud}) // Aquí se crea el JSON para el grafico de sesiones por franja horaria
	fmt.Fprintf(w, "%s;%s;%s;%s;%s;%s;%s", string(grafico1), string(grafico2), string(grafico3), string(grafico4), string(grafico5), string(grafico6), menu3)
	db0.Close()
}

func graficosMonthly(w http.ResponseWriter, r *http.Request) {

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
	var (
		arrNull, fechaAud []int
	)
	var arrAud map[int]int = make(map[int]int)
	var arrMin map[int]int = make(map[int]int)
	var arrAVG map[int]float64 = make(map[int]float64)
	var arrMegas map[int]int = make(map[int]int)
	var arrPico map[int]int = make(map[int]int)
	var horaPico map[int]int = make(map[int]int)
	mesGrafico := r.FormValue("years") + "-" + r.FormValue("months")
	// Se añaden los dias del mes al grafico
	for i := 1; i <= daysStringIn(r.FormValue("months"), toInt(r.FormValue("years"))); i++ {
		fechaAud = append(fechaAud, i)
	}
	//Se comprueba si existe la base de datos mensual
	if _, err := os.Stat(dirMonthlys + mesGrafico + "monthly.db"); os.IsNotExist(err) {
		//No hay base de datos
		Warning.Println("Database file does not exists")
		Error.Println(err)
		tipo1, _ := json.Marshal(Grafico2{"bar", arrNull, fechaAud})
		menu3 := "<option label='todo' value='todo'>All</option>"
		fmt.Fprintf(w, "%s;%s;%s;%s;%s;%s;%s", string(tipo1), string(tipo1), string(tipo1), string(tipo1), string(tipo1), string(tipo1), menu3)
	} else {
		if r.FormValue("stream") == "todo" {
			var menu3 string
			db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
			if err != nil {
				Error.Println(err)
				return
			}
			//Generamos el select de streams
			dbmon_mu.RLock()
			query2, err := db0.Query("SELECT  DISTINCT(streamname) FROM resumen WHERE username = ?", username)
			dbmon_mu.RUnlock()
			if err != nil {
				Error.Println(err)
				return
			}
			if !query2.Next() {
				//No, campos vacios
				menu3 = "<option label='All' value='todo' selected>All</option>"
			} else {
				//Si, existen campos. Formamos el select
				menu3 = "<option label='All' value='todo' selected>All</option>"
				dbmon_mu.RLock()
				query3, err := db0.Query("SELECT  DISTINCT(streamname) FROM resumen WHERE username = ?", username)
				dbmon_mu.RUnlock()
				if err != nil {
					Warning.Println(err)
					return
				}
				for query3.Next() {
					var stream string
					query3.Scan(&stream)
					menu3 += "<option label='" + stream + "' value='" + stream + "'>" + stream + "</option>"
				}
				query3.Close()
			}
			query2.Close()
			dbmon_mu.RLock()
			query, err := db0.Query("SELECT sum(players), sum(hours), avg(hours), sum(gigabytes), max(peak), peaktime, date FROM resumen WHERE username = ? GROUP BY date", username)
			dbmon_mu.RUnlock()
			if err != nil {
				Error.Println(err)
				return
			}
			for query.Next() {
				var audiencia, minutos, megas, pico int
				var promedio float64
				var horapico, fecha string
				query.Scan(&audiencia, &minutos, &promedio, &megas, &pico, &horapico, &fecha)
				hour := strings.Split(horapico, ":")
				day := strings.Split(fecha, ":")
				arrAud[toInt(day[1])] = audiencia
				arrMin[toInt(day[1])] = minutos
				arrAVG[toInt(day[1])] = promedio
				arrMegas[toInt(day[1])] = megas
				arrPico[toInt(day[1])] = pico
				horaPico[toInt(day[1])] = toInt(hour[0])
			}
			query.Close()
			//Se seneran los gŕaficos
			g1 := grafDays(arrAud, len(fechaAud))
			g2 := grafDays(arrMin, len(fechaAud))
			g3 := grafDaysFloat(arrAVG, len(fechaAud))
			g4 := grafDays(arrMegas, len(fechaAud))
			g5 := grafDays(arrPico, len(fechaAud))
			g6 := grafDays(horaPico, len(fechaAud))
			// Aquí se crean los JSON
			grafico0, _ := json.Marshal(Grafico2{"bar", g1, fechaAud})  // Aquí se crea el JSON para el grafico de audiencia total del dia en personas
			grafico1, _ := json.Marshal(Grafico2{"bar", g2, fechaAud})  // Aquí se crea el JSON para el grafico de tiempo total visionado
			grafico2, _ := json.Marshal(Grafico3{"bar", g3, fechaAud})  // Aquí se crea el JSON para el grafico de segundos consumidos por pais
			grafico3, _ := json.Marshal(Grafico2{"bar", g4, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por pais
			grafico4, _ := json.Marshal(Grafico2{"bar", g5, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por franja horaria
			grafico5, _ := json.Marshal(Grafico2{"line", g6, fechaAud}) // Aquí se crea el JSON para el grafico de sesiones por franja horaria
			fmt.Fprintf(w, "%s;%s;%s;%s;%s;%s;%s", string(grafico0), string(grafico1), string(grafico2), string(grafico3), string(grafico4), string(grafico5), menu3)
			db0.Close()
		} else {
			db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
			if err != nil {
				Error.Println(err)
				return
			}
			dbmon_mu.RLock()
			query2, err := db0.Query("SELECT  DISTINCT(streamname) FROM resumen WHERE username = ?", username)
			dbmon_mu.RUnlock()
			if err != nil {
				Error.Println(err)
				return
			}
			//Generamos el select de streams
			menu3 := "<option label='All' value='todo'>All</option>"
			//Si existen campos, formamos el select
			for query2.Next() {
				var stream string
				query2.Scan(&stream)
				if stream == r.FormValue("stream") {
					menu3 += "<option label='" + stream + "' value='" + stream + "' selected>" + stream + "</option>"
				} else {
					menu3 += "<option label='" + stream + "' value='" + stream + "'>" + stream + "</option>"
				}
			}
			query2.Close()
			dbmon_mu.RLock()
			query, err := db0.Query("SELECT sum(players), sum(hours), avg(hours), sum(gigabytes), max(peak), peaktime, date FROM resumen WHERE username = ? AND streamname = ? GROUP BY date", username, r.FormValue("stream"))
			dbmon_mu.RUnlock()
			if err != nil {
				Error.Println(err)
				return
			}
			for query.Next() {
				var audiencia, minutos, megas, pico int
				var promedio float64
				var horapico, fecha string
				query.Scan(&audiencia, &minutos, &promedio, &megas, &pico, &horapico, &fecha)
				hour := strings.Split(horapico, ":")
				day := strings.Split(fecha, ":")
				arrAud[toInt(day[1])] = audiencia
				arrMin[toInt(day[1])] = minutos
				arrAVG[toInt(day[1])] = promedio
				arrMegas[toInt(day[1])] = megas
				arrPico[toInt(day[1])] = pico
				horaPico[toInt(day[1])] = toInt(hour[0])
			}
			query.Close()
			//Se seneran los gŕaficos
			g1 := grafDays(arrAud, len(fechaAud))
			g2 := grafDays(arrMin, len(fechaAud))
			g3 := grafDaysFloat(arrAVG, len(fechaAud))
			g4 := grafDays(arrMegas, len(fechaAud))
			g5 := grafDays(arrPico, len(fechaAud))
			g6 := grafDays(horaPico, len(fechaAud))
			// Se crean los JSON
			grafico0, _ := json.Marshal(Grafico2{"bar", g1, fechaAud})  // Aquí se crea el JSON para el grafico de audiencia total del dia en personas
			grafico1, _ := json.Marshal(Grafico2{"bar", g2, fechaAud})  // Aquí se crea el JSON para el grafico de tiempo total visionado
			grafico2, _ := json.Marshal(Grafico3{"bar", g3, fechaAud})  // Aquí se crea el JSON para el grafico de segundos consumidos por pais
			grafico3, _ := json.Marshal(Grafico2{"bar", g4, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por pais
			grafico4, _ := json.Marshal(Grafico2{"bar", g5, fechaAud})  // Aquí se crea el JSON para el grafico de sesiones por franja horaria
			grafico5, _ := json.Marshal(Grafico2{"line", g6, fechaAud}) // Aquí se crea el JSON para el grafico de sesiones por franja horaria
			fmt.Fprintf(w, "%s;%s;%s;%s;%s;%s;%s", string(grafico0), string(grafico1), string(grafico2), string(grafico3), string(grafico4), string(grafico5), menu3)
			db0.Close()
		}
	}
}

// Funcion que muestra el total de horas y gigas consumidos (por primera vez)
func totalMonths(w http.ResponseWriter, r *http.Request) {

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

	var minutos, megas int
	anio, mes, _ := time.Now().Date() //Fecha actual
	mesGrafico := fmt.Sprintf("%d-%02d", anio, mes)
	db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
	if err != nil {
		Error.Println(err)
	}
	defer db0.Close()
	dbmon_mu.RLock()
	query, err := db0.Query("SELECT sum(hours), sum(gigabytes) FROM resumen WHERE username = ? GROUP BY username", username)
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
	table := fmt.Sprintf("<tr><th>Hours consumed: </th><td>&nbsp;</td><td>%d</td></tr><tr><th>GBs consumed: </th><td>&nbsp;</td><td>%d</td></tr>", minutos, megas)
	fmt.Fprintf(w, "%s", table)
}

// Funcion que muestra el total de horas y gigas consumidos (con cambio de mes)
func totalMonthsChange(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := id_[key] // De aquí podemos recoger el id del usuario logeado
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

	r.ParseForm()
	var minutos, megas int
	mesGrafico := r.FormValue("years") + "-" + r.FormValue("months")
	//Se comprueba si existe la base de datos mensual
	if _, err := os.Stat(dirMonthlys + mesGrafico + "monthly.db"); os.IsNotExist(err) {
		//No hay base de datos
		Warning.Println("Database does not exists")
		Error.Println(err)
		fmt.Fprintf(w, "%s", "NoBD")
	} else {
		db0, err := sql.Open("sqlite3", dirMonthlys+mesGrafico+"monthly.db")
		if err != nil {
			Error.Println(err)
		}
		defer db0.Close()
		dbmon_mu.RLock()
		query, err := db0.Query("SELECT sum(hours), sum(gigabytes) FROM resumen WHERE username = ? GROUP BY username", username)
		dbmon_mu.RUnlock()
		if err != nil {
			Warning.Println(err)
		}
		for query.Next() {
			err = query.Scan(&minutos, &megas)
			if err != nil {
				Warning.Println(err)
			}
		}
		query.Close()
		table := fmt.Sprintf("<tr><th>Hours consumed: </th><td>&nbsp;</td><td>%d</td></tr><tr><th>GBs consumed: </th><td>&nbsp;</td><td>%d</td></tr>", minutos, megas)
		fmt.Fprintf(w, "%s", table)
	}
}

// Se crean los canvas para colocar los gráficos
func createGraf(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	_, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	canv1 := "<label>Audience in Players</label><canvas id='graficop1'/>"
	canv2 := "<label>Hours of Playback</label><canvas id='graficop2'/>"
	canv3 := "<label>Average Time of Playback in hours</label><canvas id='graficop3'/>"
	canv4 := "<label>Traffic in Gigabytes</label><canvas id='graficop4'/>"
	canv5 := "<label>Max Audience in Players</label><canvas id='graficop5'/>"
	canv6 := "<label>Peak Times</label><canvas id='graficop6'/>"
	fmt.Fprintf(w, "%s;%s;%s;%s;%s;%s", canv1, canv2, canv3, canv4, canv5, canv6)
}

//funcion que va a colocar las datos monthly en sus correspondientes dias
func grafDaysFloat(hora map[int]float64, day int) []float64 {
	x := make([]float64, day)
	for cont, _ := range x {
		for key, value := range hora {
			if key == cont+1 {
				x[cont] = value
			}
		}
	}
	return x
}

//funcion que va a colocar las datos monthly en sus correspondientes dias
func grafDays(hora map[int]int, day int) []int {
	x := make([]int, day)
	for cont, _ := range x {
		for key, value := range hora {
			if key == cont+1 {
				x[cont] = value
			}
		}
	}
	return x
}
