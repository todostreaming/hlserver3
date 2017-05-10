package main

import (
	"fmt"
	"mime"
	"net"
	"net/http"
	"strings"
	"time"
)

/*
CREATE TABLE "players" (
	"id"  INTEGER PRIMARY KEY NOT NULL,
	"rawstream"  TEXT(255),
	"ipproxy"  TEXT(255),
	"ipclient"  TEXT(255),
	"timestamp"  INTEGER,
	"time"  INTEGER,
	"kilobytes"  INTEGER,
	"total_time"  INTEGER,
	"agent" TEXT(255)
);
*/
func filldb(w http.ResponseWriter, r *http.Request) {
	mu_ident.Lock()
	ident++
	id := ident
	mu_ident.Unlock()

	go func() {
		mu_dblive.Lock()
		_, err := dblive.Exec("INSERT INTO players (`id`, `rawstream`, `ipproxy`, `ipclient`, `timestamp`, `time`, `kilobytes`, `total_time`, `agent`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			id, "livestream", "46.0.34.7", "192.168.4.90", 14909928, 0, 0, 0, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.81 Safari/537.36")
		mu_dblive.Unlock()
		if err != nil {
			if strings.Contains(err.Error(), "constraint") { // UNIQUE constraint failed: players.id
				mu_dblive.Lock()
				dblive.Exec("UPDATE players SET time = time +10, total_time = total_time + 10 WHERE id = ?", id)
				mu_dblive.Unlock()
			} else {
				fmt.Println("DB error:", err)
			}
		}
	}()
	fmt.Fprintf(w, "record id: %d", id)
}

func geoip(w http.ResponseWriter, r *http.Request) {
	var city, country, isocode string

	ipstring := fmt.Sprintf("%d.%d.%d.%d", random(1, 255), random(1, 256), random(1, 256), random(1, 256))
	ip := net.ParseIP(ipstring)
	mu_dbgeoip.Lock()
	record, err := dbgeoip.City(ip)
	mu_dbgeoip.Unlock()
	if err != nil {
		return
	}
	city = record.City.Names["en"]
	country = record.Country.Names["en"]
	isocode = record.Country.IsoCode

	fmt.Fprintf(w, "%s [%s] (%s)\n", country, isocode, city) // avoid console printing of plenty logs

}

func cookies(w http.ResponseWriter, r *http.Request) {
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: "username", Value: "antonio", Expires: expiration}
	http.SetCookie(w, &cookie)

	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	fmt.Fprintf(w, "<h2>Cookies list:</h2><br>")
	for _, cookie := range r.Cookies() {
		fmt.Fprintf(w, "<p>%s : %s</p>", cookie.Name, cookie.Value)
	}
	fmt.Fprintf(w, "<a href=\"/delcookies.cgi\">Delete Cookies</a>")
}

func delcookies(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		cookie.MaxAge = -1
		http.SetCookie(w, cookie)
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	fmt.Fprintf(w, "<h2>Cookies list:</h2><br>")
	for _, cookie := range r.Cookies() {
		fmt.Fprintf(w, "<p>%s : %s (deleted)</p>", cookie.Name, cookie.Value)
	}

}
