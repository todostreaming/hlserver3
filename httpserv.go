package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// benchmarks: ab -r -k -l -n 30000 -c 5000 [uri]
func root(w http.ResponseWriter, r *http.Request) {
	// request uri = "http://localhost/live/luztv-livestream.w8889.m3u8?id=0x449484abb&wid=0xbc677870"
	// r.URL.Path[1:] = "live/luztv-livestream.w8889.m3u8" <=> r.URL.RawQuery = "id=0x449484abb&wid=0xbc677870"
	path := r.URL.Path[1:]
	resp := ""
	if strings.Contains(path, ".m3u8") { // .m3u8 playlists
		if strings.Contains(path, "-playlist.m3u8") { // 1st identifying playlist
			// live/luztv-livestream-playlist.m3u8
			// recover the player cookie "rawstream" => ident

			// if no cookie found, lets create a new one with a new ident number for this player and stream

			var id int64
			mu_ident.Lock()
			ident++
			id = ident
			mu_ident.Unlock()
			path = strings.Replace(path, "-playlist.m3u8", "", -1)
			tr := strings.Split(path, "/")
			resp = fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=3177936\n%s.wid%d.m3u8", tr[len(tr)-1], id)
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
					go func() {
						time.Sleep(1 * time.Millisecond) // this can be a MySQL writer INSERT ON DUPLICATE UPDATE (create very few variables inside to avoid filling the RAM)
					}()
					//createstats(r, spl[0], id) //evaluate not to use goroutines here that could overload the system and panic
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
	fmt.Printf("id=%d, rawstream=%s, ipproxy=%s, ipclient=%s, agent=%s, referer=%s\r", id, rawstream, tr[0], spl[0], r.UserAgent(), r.Referer())
	// maxmind geoip2 from (github.com/oschwald/geoip2-golang) loaded on RAM, only once openned and exclusive mutex locked at every read

	return
}
