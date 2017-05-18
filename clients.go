package main

import (
	"fmt"
	"net/http"
)

// Función para dar de alta un nuevo cliente en la tabla admin
func nuevoCliente(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	mu_dblive.Lock()
	_, err1 := dblive.Exec("INSERT INTO admin (`username`, `password`, `type`, `status`) VALUES (?, ?, ?, ?)",
		r.FormValue("nom_cli"), r.FormValue("passw"), r.FormValue("type"), r.FormValue("status"))
	mu_dblive.Unlock()
	if err1 != nil {
		Error.Println(err1)
	}
}

// Función que selecciona los clientes de la tabla admin
func buscarClientes(w http.ResponseWriter, r *http.Request) {
	var id int
	var nombre, selector string
	query, err := dblive.Query("SELECT id, username FROM admin WHERE type = 0")
	if err != nil {
		Error.Println(err)
	}
	for query.Next() {
		err = query.Scan(&id, &nombre)
		if err != nil {
			Warning.Println(err)
		}
		selector = fmt.Sprintf("<option value='%d'>%s</option>", id, nombre)
		fmt.Fprintf(w, "%s", selector)
	}
	query.Close()
}

// Función que borra un cliente de la tabla admin
func borrarCliente(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	mu_dblive.Lock()
	_, err1 := dblive.Exec("DELETE FROM admin WHERE id = ?", r.FormValue("clients"))
	mu_dblive.Unlock()
	if err1 != nil {
		Error.Println(err1)
	}
}
