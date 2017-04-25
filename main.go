package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ident    int64      // identifier for every streaming session openned by an individual player
	mu_ident sync.Mutex // exclusive mutex for the identifier

	dbplayers    *sql.DB    // db only with live players raw info
	mu_dbplayers sync.Mutex // also exclusive mutex for
)

func init() {
	var err error
	dbplayers, err = sql.Open("sqlite3", "./players.db") // on RAMdisk
	if err != nil {
		log.Fatalln("Fallo al abrir el archivo DB:", err)
	}
	dbplayers.Exec("PRAGMA journal_mode=WAL;")
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

	http.HandleFunc("/", root)
	http.HandleFunc("/filldb.cgi", filldb)

	log.Fatal(s.ListenAndServe()) // Servidor HTTP multihilo
}

// sirve todos los ficheros estáticos de la web html,css,js,graficos,etc
// benchmarks: ab -r -k -n 30000 -c 5000 [uri]
func root(w http.ResponseWriter, r *http.Request) {
	// request uri = "http://localhost/live/luztv-livestream.w8889.m3u8?id=0x449484abb&wid=0xbc677870"
	// r.URL.Path[1:] = "live/luztv-livestream.w8889.m3u8" <=> r.URL.RawQuery = "id=0x449484abb&wid=0xbc677870"
	path := r.URL.Path[1:]
	resp := ""
	if strings.Contains(path, ".m3u8") { // .m3u8 playlists
		if strings.Contains(path, "-playlist.m3u8") { // 1st identifying playlist
			// live/luztv-livestream-playlist.m3u8
			var id int64
			mu_ident.Lock()
			ident++
			id = ident
			mu_ident.Unlock()
			path = strings.Replace(path, "-playlist.m3u8", "", -1)
			tr := strings.Split(path, "/")
			resp = fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=3177936,CODECS=\"avc1.100.41, mp4a.40.2\",RESOLUTION=1920x1080\n%s.wid%d.m3u8", tr[len(tr)-1], id)
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Accept-Ranges", "bytes")
			fmt.Fprintf(w, "%s\n", resp)
			return
		} else if strings.Contains(path, ".wid") { // recursive identified playlist
			// live/luztv-livestream.wid45006.m3u8
			var id int64
			tr := strings.Split(path, "/")
			spl := strings.Split(tr[len(tr)-1], ".")
			if len(spl) == 3 { // we response the content of the original .m3u8 playlist and record on database the stats info
				fmt.Sscanf(spl[1], "wid%d", &id)
				file := spl[0] + ".m3u8"
				fileinfo, err := os.Stat(file)
				if err != nil {
					http.NotFound(w, r)
					return
				} else {
					fr, errn := os.Open(file)
					if errn != nil {
						http.Error(w, "Internal Server Error", 500)
						return
					}
					defer fr.Close()
					go createstats(r, spl[0], id)
					w.Header().Set("Cache-Control", "no-cache")
					w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
					w.Header().Set("Access-Control-Allow-Headers", "*")
					w.Header().Set("Access-Control-Expose-Headers", "*")
					w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					w.Header().Set("Access-Control-Allow-Origin", "*")
					w.Header().Set("Content-Length", fmt.Sprintf("%d", fileinfo.Size()))
					w.Header().Set("Accept-Ranges", "bytes")
					io.Copy(w, fr)
					return
				}
			} else {
				http.NotFound(w, r)
				return
			}
		} else {
			http.NotFound(w, r)
			return
		}
	} else if strings.Contains(path, ".ts") { // .TS segments
		// live/segment-56.ts
		spl := strings.Split(path, "/")
		file := spl[len(spl)-1]
		fileinfo, err := os.Stat(file)
		if err != nil {
			http.NotFound(w, r)
			return
		} else {
			fr, errn := os.Open(file)
			if errn != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}
			defer fr.Close()
			w.Header().Set("Cache-Control", "max-age=300")
			w.Header().Set("Content-Type", "video/MP2T")
			w.Header().Set("Access-Control-Allow-Headers", "*")
			w.Header().Set("Access-Control-Expose-Headers", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileinfo.Size()))
			w.Header().Set("Accept-Ranges", "bytes")
			io.Copy(w, fr)
			return
		}
	} else { // regular web content

	}
}

func createstats(r *http.Request, rawstream string, id int64) { // function to record on DB insert on duplicate update "INSERT INTO table (a,b,c) VALUES (1,2,3) ON DUPLICATE KEY UPDATE c=c+1;"
	/*
		Remote-Ip => [79.109.178.183]
		X-Remote-Ip => [79.109.178.183]
	*/
	var remoteip string
	value, ok := r.Header["Remote-Ip"]
	if !ok {
		remoteip = r.RemoteAddr
	} else {
		remoteip = value[0]
	}
	tr := strings.Split(r.RemoteAddr, ":")
	spl := strings.Split(remoteip, ":")
	fmt.Printf("id=%d, rawstream=%s, ipproxy=%s, ipclient=%s, agent=%s, referer=%s\n\n", id, rawstream, tr[0], spl[0], r.UserAgent(), r.Referer())

	return
}

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

	mu_dbplayers.Lock()
	_, err := dbplayers.Exec("INSERT INTO players (`id`, `rawstream`, `ipproxy`, `ipclient`, `timestamp`, `time`, `kilobytes`, `total_time`, `agent`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, "livestream", "46.0.34.7", "192.168.4.90", 14909928, 0, 0, 0, "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.81 Safari/537.36")
	mu_dbplayers.Unlock()
	if err != nil {
		if strings.Contains(err.Error(), "constraint") { // UNIQUE constraint failed: players.id
			mu_dbplayers.Lock()
			dbplayers.Exec("UPDATE players SET time = time +10, total_time = total_time + 10 WHERE id = ?", id)
			mu_dbplayers.Unlock()
		} else {
			fmt.Println("DB error:", err)
		}
	}
}
