package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"strings"
)

func publish(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	stream := strings.Split(r.FormValue("name"), "-")
	nom_user := stream[0]

	dbgeneral, err := sql.Open("sqlite3", DirDB+"general.db") // Apertura de la dateDayly.db antigua para lectura del pico/hora
	if err != nil {
		Error.Println(err)
	}
	defer dbgeneral.Close()
	dbgen_mu.RLock()
	query, err := dbgeneral.Query("SELECT username, pubpass, type, status FROM users WHERE username = ?", nom_user)
	dbgen_mu.RUnlock()
	if err != nil {
		Warning.Println(err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer query.Close()
	for query.Next() {
		var user, pass string
		var status, type_ int
		err = query.Scan(&user, &pass, &type_, &status)
		if err != nil {
			Warning.Println(err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		if user == r.FormValue("username") && pass == r.FormValue("password") && r.FormValue("call") == "publish" && status == 1 && type_ == 2 {
			fmt.Fprintf(w, "Server OK")
			return
		} else {
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
	http.Error(w, "Internal Server Error", 500)
}

func onplay(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Internal Server Error", 500)
}
