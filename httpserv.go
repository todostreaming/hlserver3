package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
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
				key = fmt.Sprintf("%d", id)
			} else {
				// we have the cookie so we have the ident of this player
				key = cookie.Value // this is the id in string form
			}
			resp = fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%d\n%s.wid%s.m3u8\n", bps, rawstream, key)
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
			fmt.Fprintf(w, "%s", resp)
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
				fileinfo, err := os.Stat(rootdir + "old/" + file)
				if err != nil {
					http.NotFound(w, r)
					return
				} else {
					// open the old/file.m3u8 (forecast pre-caching mechanism for CDNs)
					fr, errn := os.Open(rootdir + "old/" + file)
					if errn != nil {
						http.Error(w, "Internal Server Error", 500)
						return
					}
					defer fr.Close()
					if numgo < 1000000 { // if there are more than 1M goroutines working, live stats will stop for a while
						go createstats(r, spl[0], id)
					}
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
			fr, err := os.Open(rootdir + path)
			if err != nil {
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
	} else if strings.Contains(path, "-precache.js") {
		// this code will send the javascript code for the forecaster mechanism of pre-caching
		// http://hlserver/rawstream-precache.js (path = rawstream-precache.js)
		file := rootdir + "precache.js"
		_, err := os.Stat(file)
		if err != nil { // does not exist the path (file nor dir)
			http.NotFound(w, r)
			return
		}
		fr, err := os.Open(file)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		defer fr.Close()
		buf, err := ioutil.ReadAll(fr)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		jscode := string(buf)
		mu_cloud.RLock()
		domain := cloud["cloudserver"]
		mu_cloud.RUnlock()
		rawstream := strings.Replace(path, "-precache.js", "", -1)
		jscode = strings.Replace(jscode, "mydomain", domain, -1)
		jscode = strings.Replace(jscode, "rawstream", rawstream, -1)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/javascript; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Accept-Ranges", "bytes")
		fmt.Fprintf(w, "%s", jscode)
	} else if strings.Contains(path, ".lst") { // latest segment to pre-cache (forecaster map)
		// http://hlserver3/rawstream.lst
		// tail -1 /var/segments/new/luztv-livestream.m3u8
		preload := false
		tr := strings.Split(path, "/")
		spl := strings.Split(tr[len(tr)-1], ".") // spl[0] = rawstream
		ip := getip(r.RemoteAddr)                // take near_proxy ip or r.Header["X-Cdn-Pop"] = [gsw] ???
		key := ip + "=" + spl[0]
		val, ok := Forecaster.Load(key)
		t := time.Now().Unix()
		if ok {
			if t-val.(int64) >= 4 { // 4 secs
				preload = true
				Forecaster.Store(key, t)
			}
		} else {
			preload = true
			Forecaster.Store(key, t)
		}
		if preload {
			resp = getlatestseg(spl[0])
		}
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Accept-Ranges", "bytes")
		fmt.Fprintf(w, "%s", resp) // here the response is the .ts file or just a blank string
	} else { // regular web content: http://hlserver.com/path_to_file
		file := rootdir + path
		fileinfo, err := os.Stat(file)
		if err != nil { // does not exist the path (file nor dir)
			http.NotFound(w, r)
			return
		} else if fileinfo.IsDir() { // it is a dir
			if strings.HasSuffix(file, "/") { // add /index.html to the end
				file = file + first_page + ".html"
			} else {
				file = file + "/" + first_page + ".html"
			}
		}
		fr, err := os.Open(file)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
		defer fr.Close()
		if session {
			if strings.Contains(r.URL.String(), "?err") {
				// replace <span id="loginerr"></span> with an error text to show
				buf, _ := ioutil.ReadAll(fr)
				html := string(buf)
				html = strings.Replace(html, spanHTMLlogerr, ErrorText, -1)
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
				fmt.Fprint(w, html)
			} else {
				// Get the cookies
				filepart := strings.Split(file, ".")
				if (filepart[1] != "html") || (filepart[0] == (rootdir + first_page)) {
					w.Header().Set("Cache-Control", "no-cache")
					http.ServeContent(w, r, file, fileinfo.ModTime(), fr)
				} else {
					cookie, err := r.Cookie(CookieName)
					if err != nil {
						Error.Println("Cookie not found in the browser")
						http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
					} else {
						key := cookie.Value
						mu_user.RLock()
						_, ok := user_[key]
						mu_user.RUnlock()
						if ok {
							cookie.Expires = time.Now().Add(time.Duration(session_timeout) * time.Second)
							http.SetCookie(w, cookie)
							mu_user.Lock()
							time_[cookie.Value] = cookie.Expires
							mu_user.Unlock()
							w.Header().Set("Cache-Control", "no-cache")
							http.ServeContent(w, r, file, fileinfo.ModTime(), fr)
						} else {
							Error.Println("Cookie not found in the server")
							http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
						}
					}
				}
			}
		} else {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeContent(w, r, file, fileinfo.ModTime(), fr)
		}
	}
}

// just record in live.db @ table players (insert or update)
func createstats(r *http.Request, rawstream string, id int64) {
	// lets collect all the info to save in live.db
	var ipclient, ipproxy, username, streamname, os string
	var country, isocode, city string

	tr := strings.Split(rawstream, "-")
	if len(tr) != 2 {
		return
	}
	username, streamname = tr[0], tr[1]
	value, ok := r.Header["Remote-Ip"]
	if !ok {
		ipclient = getip(r.RemoteAddr)
	} else {
		ipclient = value[0]
	}
	ipproxy = getip(r.RemoteAddr)
	os = getos(r.UserAgent())
	country, isocode, city = geoIP(ipclient)
	timestamp := time.Now().Unix()

	// store everything in live.db
	mu_dblive.Lock()
	_, err := dblive.Exec("INSERT INTO players (`id`, `username`, `streamname`, `os`, `ipproxy`, `ipclient`, `isocode`, `country`, `city`, `timestamp`, `time`, `kilobytes`, `total_time`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		id, username, streamname, os, ipproxy, ipclient, isocode, country, city, timestamp, 0, 0, 0)
	mu_dblive.Unlock()
	if err != nil {
		if strings.Contains(err.Error(), "constraint") { // UNIQUE constraint failed: players.id // this id existed
			// let's get: timestamp, time, kilobytes, total_time and bandwith(rawstream) // ["luztv-livestream"] = 3780000 bps
			bw := 0
			bandwidth, ok := Bw_int.Load(rawstream)
			if ok {
				bw = bandwidth.(int)
			}
			var timestamp_db, time_db, kilobytes_db, total_time_db int64
			err := dblive.QueryRow("SELECT timestamp, time, kilobytes, total_time FROM players WHERE id = ?", id).Scan(&timestamp_db, &time_db, &kilobytes_db, &total_time_db)
			if err != nil {
				Error.Println(err)
				return
			}
			seconds := timestamp - int64(timestamp_db)
			if seconds > 30 { // reconnected from a previous disconn
				mu_dblive.Lock()
				dblive.Exec("UPDATE players SET time = 0, timestamp = ? WHERE id = ?", timestamp, id)
				mu_dblive.Unlock()
			} else { // still connected
				KBs := kilobytes_db + (int64(bw) * seconds / 8192)
				mu_dblive.Lock()
				dblive.Exec("UPDATE players SET time = ?, total_time = ?, kilobytes = ?, timestamp = ? WHERE id = ?", time_db+seconds, total_time_db+seconds, KBs, timestamp, id)
				mu_dblive.Unlock()
			}
		} else {
			Error.Println(err)
		}
	}
}
