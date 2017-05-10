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
	go encoder()

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
					Bw_int.Store(val.Nombre, bitrate)                                                 // Save the bitrate
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
