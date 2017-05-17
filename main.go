// ab -r -n 100000 -c 200 -l "http://127.0.0.1:9999/geoip.cgi"
package main

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/sync/syncmap"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func init() {
	var err error
	// Logging errors machanism
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Fallo al abrir el archivo de error:", err)
	}
	Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(io.MultiWriter(file, os.Stderr), "ERROR :", log.Ldate|log.Ltime|log.Lshortfile)
	// Live DB
	dblive, err = sql.Open("sqlite3", "/var/db/live.db") // on RAMdisk
	if err != nil {
		log.Fatalln("Fallo al abrir el archivo DB:", err)
	}
	dblive.Exec("PRAGMA journal_mode=WAL;")
	// GeoIP2 DB
	dbgeoip, err = geoip2.Open("/var/db/GeoIP2-City.mmdb")
	if err != nil {
		log.Fatalln("Fallo al abrir el GeoIP2:", err)
	}
	// empty the bitrates map
	Bw_int = new(syncmap.Map)
	// empty the referer map
	Referer = new(syncmap.Map)
	// empty the forecaster map
	Forecaster = new(syncmap.Map)
}

func main() {
	// Handlers del Servidor HTTP
	s := &http.Server{
		Addr:           ":9999",          // config http port
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
	go encoder()
	go maintenance()
	go diskforecastmechanism()

	http.HandleFunc("/", root)
	// testing functions
	http.HandleFunc("/filldb.cgi", filldb)
	http.HandleFunc("/geoip.cgi", geoip)
	http.HandleFunc("/cookies.cgi", cookies)
	http.HandleFunc("/delcookies.cgi", delcookies)

	log.Fatal(s.ListenAndServe()) // Servidor HTTP multihilo
}

// every 3 seconds we explore xml stats of RTMP streams published
func encoder() {
	var username, streamname string
	var count int
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for {
		// xml tree to follow
		type Client struct {
			Ip      string `xml:"address"`
			Time    string `xml:"time"`
			Publish int    `xml:"publishing"`
		}
		type Stream struct {
			Nombre     string   `xml:"name"`
			Bw_in      string   `xml:"bw_in"`
			Width      string   `xml:"meta>video>width"`
			Height     string   `xml:"meta>video>height"`
			Frame      string   `xml:"meta>video>frame_rate"`
			Vcodec     string   `xml:"meta>video>codec"`
			Acodec     string   `xml:"meta>audio>codec"`
			ClientList []Client `xml:"client"`
		}
		type Result struct {
			Stream []Stream `xml:"server>application>live>stream"`
		}

		// load stats.xml and start the parsing and DB update
		resp, err := client.Get("http://127.0.0.1:8080/stats")
		if err != nil {
			Warning.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			Warning.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		v := Result{}
		err = xml.Unmarshal([]byte(body), &v)
		if err != nil {
			Error.Printf("xml read error: %s", err)
			time.Sleep(2 * time.Second)
			continue
		}
		for _, val := range v.Stream {
			for _, val2 := range val.ClientList {
				if val2.Publish == 1 {
					userstream := strings.Split(val.Nombre, "-")
					if len(userstream) > 1 {
						username = userstream[0]
						streamname = userstream[1]
					}
					tiempo := toInt(val2.Time) / 1000 // convert msec to sec
					tiempo_now := time.Now().Unix()
					bitrate := toInt(val.Bw_in)                                                       // bps
					Bw_int.Store(val.Nombre, bitrate)                                                 // ["luztv-livestream"] = 3780000
					info := fmt.Sprintf("%sx%s %s/%s", val.Width, val.Height, val.Vcodec, val.Acodec) // 1280x720 H264/AAC
					err := dblive.QueryRow("SELECT count(*) FROM encoders WHERE username = ? AND streamname = ? AND ip= ?", username, streamname, val2.Ip).Scan(&count)
					if err != nil {
						Error.Println(err)
					}
					if count == 0 { // not record of user, stream, ip
						country, isocode, city := geoIP(val2.Ip) // Datos de geolocalización
						if isocode == "" {
							isocode = "OT" //cuando el isocode esta vacio, lo establecemos a OT (other)
						}
						if country == "" {
							country = "Unknown" //cuando el country esta vacio, lo establecemos a Unknown (desconocido)
						}
						mu_dblive.Lock()
						_, err := dblive.Exec("INSERT INTO encoders (`username`, `streamname`, `time`, `bitrate`, `ip`, `info`, `isocode`, `country`, `city`, `timestamp`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
							username, streamname, tiempo, bitrate, val2.Ip, info, isocode, country, city, tiempo_now)
						mu_dblive.Unlock()
						if err != nil {
							Error.Println(err)
						}
					} else { // pre-existing record, just update
						mu_dblive.Lock()
						_, err := dblive.Exec("UPDATE encoders SET username=?, streamname=?, time=?, bitrate=?, info=?, timestamp=? WHERE username = ? AND streamname = ? AND ip = ?",
							username, streamname, tiempo, bitrate, info, tiempo_now, username, streamname, val2.Ip)
						mu_dblive.Unlock()
						if err != nil {
							Error.Println(err)
						}
					}
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
}

// MAINTENACE TASKS
func maintenance() {
	var fecha_actual, fecha_antigua string
	var mes_actual, mes_antiguo string
	for {
		cambio_de_fecha := false
		cambio_de_mes := false
		hh, mm, _ := time.Now().Clock()
		anio, mes, dia := time.Now().Date() //Fecha actual
		// Se saca la hora y los minutos
		fecha_actual = fmt.Sprintf("%04d-%02d-%02d", anio, mes, dia) // Calculo de fecha actual
		// Se comprueba si hay cambio de dia
		if fecha_actual != fecha_antigua { // dayly.db
			cambio_de_fecha = true
			if _, err := os.Stat(dirDaylys + fecha_actual + "dayly.db"); err == nil {
				cambio_de_fecha = false // se debe a un reinicio del hlserver
			}
		}
		// Se comprueba si hay cambio de mes
		mes_actual = fecha_actual[0:7] // year-month
		if mes_actual != mes_antiguo { // monthly.db
			cambio_de_mes = true
			if _, err := os.Stat(dirMonthlys + mes_actual + "monthly.db"); err == nil {
				cambio_de_mes = false // se debe a un reinicio del hlserver
			}
		}
		if cambio_de_mes {
			// Aqui hago la copia de monthly.db en mes_actual + monthly.db
			exec.Command("/bin/sh", "-c", "cp "+monthlyDB+" "+dirMonthlys+mes_actual+"monthly.db").Run()
		}
		if cambio_de_fecha {
			//Comprobamos si existe el fichero con fecha antigua
			if _, err := os.Stat(dirDaylys + fecha_antigua + "dayly.db"); os.IsNotExist(err) {
				// Aqui hago la copia de dayly.db en fecha_actual + dayly.db
				exec.Command("/bin/sh", "-c", "cp "+daylyDB+" "+dirDaylys+fecha_actual+"dayly.db").Run()
			} else {
				exec.Command("/bin/sh", "-c", "cp "+daylyDB+" "+dirDaylys+fecha_actual+"dayly.db").Run()
				limit_time := time.Now().Unix() - 86400
				//Sacamos los datos de la fecha
				datos_antiguos := strings.Split(fecha_antigua, "-")
				fechaMonth := fmt.Sprintf("%s:%s", datos_antiguos[1], datos_antiguos[2])
				// Antes de nada borramos los players con timestamp a más de 1 día
				mu_dblive.Lock()
				dblive.Exec("DELETE FROM players WHERE timestamp < ?", limit_time)
				mu_dblive.Unlock()
				// Se seleccionan el total de Ips, las horas totales y el total de Gigabytes
				query, err := dblive.Query("SELECT count(id), sum(total_time)/3600, sum(kilobytes)/1000000, username, streamname FROM players GROUP BY username, streamname")
				if err != nil {
					Error.Println(err)
				}
				db1, err := sql.Open("sqlite3", dirDaylys+fecha_antigua+"dayly.db") // Apertura de la dateDayly.db antigua para lectura del pico/hora
				if err != nil {
					Error.Println(err)
				}
				db2, err := sql.Open("sqlite3", dirMonthlys+mes_antiguo+"monthly.db") // Apertura de mes actual + Monthly.db para escritura del resumen del pasado dia
				if err != nil {
					Error.Println(err)
				}
				//Declaracion de variables
				var ips, horas, gigas, pico, horapico, minpico int
				var userName, streamName string
				for query.Next() {
					err = query.Scan(&ips, &horas, &gigas, &userName, &streamName)
					if err != nil {
						Error.Println(err)
					}
					// Se seleccionan el máximo de usuarios conectados, y la hora:min de la dayly antigua
					// SELECT sum(count) AS cuenta, username, streamname, hour, minutes FROM resumen WHERE username = ? AND streamname = ? GROUP BY username, streamname, hour, minutes ORDER BY cuenta DESC
					dbday_mu.RLock()
					err := db1.QueryRow("SELECT sum(players) AS cuenta, username, streamname, hour, minutes FROM resumen WHERE username = ? AND streamname = ? GROUP BY username, streamname, hour, minutes ORDER BY cuenta DESC", userName, streamName).Scan(&pico, &userName, &streamName, &horapico, &minpico)
					dbday_mu.RUnlock()
					if err != nil {
						Error.Println(err)
					}
					hourMin := fmt.Sprintf("%02d:%02d", horapico, minpico) //hour:min para monthly.db
					dbmon_mu.Lock()
					// Inserto los datos de resumen mensual
					_, err1 := db2.Exec("INSERT INTO resumen (`username`,`streamname`, `players`, `minutes`, `peak`, `peaktime`, `gigabytes`, `date`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
						userName, streamName, ips, horas, pico, hourMin, gigas, fechaMonth)
					dbmon_mu.Unlock()
					if err1 != nil {
						Error.Println(err1)
					}
				}
				query.Close()
				db2.Close()
				db1.Close()
				// Ponemos kilobytes, total_time a CERO de live.db xq empezamos un nuevo dia con trafico y horas acumuladas a CERO
				mu_dblive.Lock()
				dblive.Exec("UPDATE players SET kilobytes=0 , total_time=0")
				mu_dblive.Unlock()
			}
		}
		// Solo grabaremos en este minuto en dayly.db los q estan activos ahora mismo
		tiempo_limite := time.Now().Unix() - 30
		var user, stream, so, isocode string
		var num_filas, total_time, total_kb, proxies int
		db3, err := sql.Open("sqlite3", dirDaylys+fecha_actual+"dayly.db") // Apertura de dateDayly.db
		if err != nil {
			Error.Println(err)
		}
		query, err := dblive.Query("SELECT count(id), username, streamname, os, isocode, sum(total_time), sum(kilobytes), count(distinct(ipproxy)) FROM players WHERE timestamp > ? AND time > 0 GROUP BY username, streamname, os, isocode", tiempo_limite)
		if err != nil {
			Error.Println(err)
		}
		for query.Next() {
			err = query.Scan(&num_filas, &user, &stream, &so, &isocode, &total_time, &total_kb, &proxies)
			if err != nil {
				Error.Println(err)
			}
			dbday_mu.Lock()
			// inserto los datos de resumen
			_, err1 := db3.Exec("INSERT INTO resumen (`username`, `streamname`, `os`, `isocode`, `time`, `kilobytes`, `players`, `proxies`, `hour`, `minutes`, `date`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				user, stream, so, isocode, total_time, total_kb, num_filas, proxies, hh, mm, fecha_actual)
			dbday_mu.Unlock()
			if err1 != nil {
				Error.Println(err1)
			}
		}
		query.Close()
		db3.Close()

		fecha_antigua = fecha_actual
		mes_antiguo = mes_actual
		time.Sleep(1 * time.Minute)
	}
}
