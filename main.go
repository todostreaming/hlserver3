// ab -r -n 100000 -c 200 -l "http://127.0.0.1:9999/geoip.cgi"
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
)

var (
	ident    int64      // identifier for every streaming session openned by an individual player
	mu_ident sync.Mutex // exclusive mutex for the identifier

	dbplayers    *sql.DB    // db only with live players raw info
	mu_dbplayers sync.Mutex // also exclusive mutex for

	dbgeoip    *geoip2.Reader
	mu_dbgeoip sync.Mutex
)

func init() {
	var err error
	dbplayers, err = sql.Open("sqlite3", "/var/db/players.db") // on RAMdisk
	if err != nil {
		log.Fatalln("Fallo al abrir el archivo DB:", err)
	}
	dbplayers.Exec("PRAGMA journal_mode=WAL;")

	dbgeoip, err = geoip2.Open("/var/db/GeoIP2-City.mmdb")
	if err != nil {
		log.Fatalln("Fallo al abrir el GeoIP2:", err)
	}

}

func main() {
	// Handlers del Servidor HTTP
	s := &http.Server{
		Addr:           ":80",            // config http port
		Handler:        nil,              // Default Muxer for handler as usual
		ReadTimeout:    20 * time.Second, // send a segment in POST body
		WriteTimeout:   20 * time.Second, // receive a segment in GET req
		MaxHeaderBytes: 1 << 13,          // 8K as Apache and others
	}
	go func() {
		var old, num, max int
		for {
			num = runtime.NumGoroutine()
			if num > old {
				max = num
			}
			fmt.Printf("%d / %d                            \r", runtime.NumGoroutine(), max)
			time.Sleep(100 * time.Millisecond)
			old = num
		}
	}()

	http.HandleFunc("/", root)
	http.HandleFunc("/filldb.cgi", filldb)
	http.HandleFunc("/geoip.cgi", geoip)
	http.HandleFunc("/cookies.cgi", cookies)
	http.HandleFunc("/delcookies.cgi", delcookies)

	log.Fatal(s.ListenAndServe()) // Servidor HTTP multihilo
}
