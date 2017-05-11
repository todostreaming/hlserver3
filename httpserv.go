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
	path := r.URL.Path[1:] // live/luztv-livestream.m3u8
	resp := ""
	if strings.Contains(path, ".m3u8") { // .m3u8 playlists
		if strings.Contains(path, "-playlist.m3u8") { // 1st identifying playlist
			// path = live/luztv-livestream-playlist.m3u8
			// recover the player cookie "rawstream" => ident
			var id int64
			var bps int
			var rawstream, key string
			a := strings.Split(path, "/")
			if len(a) == 2 {
				rawstream = strings.Replace(a[1], "-playlist.m3u8", "", -1) // luztv-livestream
			} else {
				http.NotFound(w, r)
				return
			}
			// get the bandwidth
			val, ok := Bw_int.Load(rawstream)
			if ok {
				bps = val.(int)
			} else {
				bps = 1000000 // 1 Mbps by default to avoid empty playlist.m3u8
			}
			// get the player cookie
			cookie, err := r.Cookie(rawstream)
			if err != nil {
				// not player cookie, let's create one
				mu_ident.Lock()
				ident++
				id = ident
				mu_ident.Unlock()
				key = fmt.Sprintf("%s", id)
			} else {
				// we have the cookie so we have the ident of this player
				key = cookie.Value // this is the id in string form
			}
			resp = fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%d\n%s.wid%d.m3u8", bps, rawstream, key)
			expiration := time.Now().Add(24 * time.Hour)
			newcookie := http.Cookie{Name: rawstream, Value: key, Expires: expiration}
			http.SetCookie(w, &newcookie)
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
				// we have to watch the referer, if allowed for this rawstream  ["rawstream"] = "domain1.com;domain2.com"
				val, ok := Referer.Load(spl[0])
				if ok {
					// get the referrer from request and compare to the one in the map (url: http://www.w3.org/hypertext/DataSources/Overview.html)
					if !strings.Contains(val.(string), getdomain(r.Referer())) {
						http.NotFound(w, r)
						return
					}
				}
				fmt.Sscanf(spl[1], "wid%d", &id) // wid9876
				file := spl[0] + ".m3u8"         // rawstream = spl[0]
				fileinfo, err := os.Stat(rootdir + "live/old/" + file)
				if err != nil {
					http.NotFound(w, r)
					return
				} else {
					// open the old/file.m3u8 (forecast pre-caching mechanism for CDNs)
					fr, errn := os.Open(rootdir + "live/old/" + file)
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
		fileinfo, err := os.Stat(rootdir + path)
		if err != nil {
			http.NotFound(w, r)
			return
		} else {
			fr, errn := os.Open(rootdir + path)
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

// just record in live.db @ table players (insert or update)
func createstats(r *http.Request, rawstream string, id int64) {
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
