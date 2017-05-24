package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func play(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	loadSettings(playingsRoot)
	r.ParseForm() // recupera campos del form tanto GET como POST
	allname := username + "-" + r.FormValue("stream")
	mu_cloud.RLock()
	stream := "http://" + cloud["cloudserver"] + "/live/" + allname + ".m3u8"
	mu_cloud.RUnlock()
	//video := fmt.Sprintf("<script type='text/javascript' src='http://www.domainplayers.org/js/jwplayer.js'></script><div id='container'><video width='600' height='409' controls autoplay src='%s'/></div><script type='text/javascript'>jwplayer('container').setup({ width: '600', height: '409', skin: 'http://www.domainplayers.org/newtubedark.zip', plugins: { 'http://www.domainplayers.org/qualitymonitor.swf' : {} }, image: '', modes: [{ type:'flash', src:'http://www.domainplayers.org/player.swf', config: { autostart: 'true', provider:'http://www.domainplayers.org/HLSProvider5.swf', file:'%s' } }]});</script>", stream, stream)
	video := fmt.Sprintf("<script src=\"http://domainplayers.org/js/hls.min.js\"></script><script src=\"http://domainplayers.org/js/html5play.min.js\"></script><video id=\"video_x890\" controls width=\"600\" height=\"409\"><source id=\"src_x890\">Your browser does not support HTML5 video. We recommend using <a href=\"https://www.google.es/chrome/browser/desktop/\">Google Chrome</a></video><script>var url = \"%s\";html5player(url, 1, \"video_x890\", \"src_x890\");</script>", stream)
	fmt.Fprintf(w, "%s", video)
}

func encoderStatNow(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := id_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// actualizamos la cookie actual
	expiration := time.Now().Add(time.Duration(session_timeout) * time.Second)
	newcookie := http.Cookie{Name: CookieName, Value: key, Expires: expiration}
	http.SetCookie(w, &newcookie)
	mu_user.Lock()
	time_[key] = expiration
	mu_user.Unlock()
	// ---- end of session identification -------------------------------

	anio, mes, dia := time.Now().Date()
	fecha := fmt.Sprintf("%02d/%02d/%02d", dia, mes, anio)
	hh, mm, _ := time.Now().Clock()
	hora := fmt.Sprintf("%02d:%02d", hh, mm)
	tiempo_limite := time.Now().Unix() - 6 //tiempo limite de 6 seg
	query, err := dblive.Query("SELECT streamname, isocode, ip, country, time, bitrate, info FROM encoders WHERE username = ? AND timestamp > ?", username, tiempo_limite)
	if err != nil {
		Error.Println(err)
		return
	}
	fmt.Fprintf(w, "<h1>%s</h1><p><b>Conected on %s at %s UTC</b></p><table class=\"table table-striped table-bordered table-hover\"><th>Play</th><th>INFO</th><th>Country</th><th>IP</th><th>Stream</th><th>Connected time</th>", username, fecha, hora)
	for query.Next() {
		var isocode, country, streamname, ip, time_connect, info string
		var tiempo, bitrate int
		query.Scan(&streamname, &isocode, &ip, &country, &tiempo, &bitrate, &info)
		isocode = strings.ToLower(isocode)
		time_connect = secs2time(tiempo)
		INFO := fmt.Sprintf("%s [%d kbps]", info, bitrate/1000)
		fmt.Fprintf(w, "<tr><td><a href=\"javascript:launchRemote('play.cgi?stream=%s')\"><img src='images/play.jpg' border='0' title='Play %s'/></a></td><td>%s</td><td><img src=\"images/flags/%s.png\" title=\"%s\"></td><td>%s</td><td>%s</td><td>%s</td></tr>",
			streamname, streamname, INFO, isocode, country, ip, streamname, time_connect)
	}
	query.Close()
	fmt.Fprintf(w, "</table>")
}

func playerStatNow(w http.ResponseWriter, r *http.Request) {

	// --- we must identify the session user 1st ------------------------
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	key := cookie.Value
	mu_user.RLock()
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
	mu_user.RUnlock()
	if !ok {
		http.Redirect(w, r, "/"+first_page+".html", http.StatusFound)
		return
	}
	// ---- end of session identification -------------------------------

	var contador int
	tiempo_limite := time.Now().Unix() - 30 //tiempo limite de 30 seg
	err = dblive.QueryRow("SELECT count(id) FROM players WHERE username = ? AND timestamp > ? AND time > 0", username, tiempo_limite).Scan(&contador)
	if err != nil {
		Error.Println(err)
		return
	}
	if contador >= 100 {
		query, err := dblive.Query("SELECT isocode, country, count(ipclient) AS count, streamname FROM players WHERE username = ? AND timestamp > ? AND time > 0 GROUP BY isocode, streamname ORDER BY streamname, count DESC", username, tiempo_limite)
		if err != nil {
			Error.Println(err)
			return
		}
		fmt.Fprintf(w, "<table class=\"table table-striped table-bordered table-hover\"><th>Country</th><th>Number of Players</th><th>Stream</th>")
		fmt.Fprintf(w, "<tr><td align=\"center\" colspan='3'><b>Total:</b> %d players conectados</td></tr>", contador)
		for query.Next() {
			var isocode, country, ips, streamname string
			query.Scan(&isocode, &country, &ips, &streamname)
			isocode = strings.ToLower(isocode)
			fmt.Fprintf(w, "<tr><td>%s <img class='pull-right' src=\"images/flags/%s.png\" title=\"%s\"></td><td>%s</td><td>%s</td></tr>",
				country, isocode, country, ips, streamname)
		}
		query.Close()
		fmt.Fprintf(w, "</table>")
	} else {
		query, err := dblive.Query("SELECT isocode, country, city, ipclient, os, streamname, time FROM players WHERE username = ? AND timestamp > ? AND time > 0 ORDER BY streamname, time DESC", username, tiempo_limite)
		if err != nil {
			Warning.Println(err)
		}
		fmt.Fprintf(w, "<table class=\"table table-striped table-bordered table-hover\"><th>País</th><th>Region</th><th>Ciudad</th><th>Dirección IP</th><th>Stream</th><th>O.S</th><th>Tiempo conectado</th>")
		fmt.Fprintf(w, "<tr><td align=\"center\" colspan='7'><b>Total:</b> %d players conectados</td></tr>", contador)
		for query.Next() {
			var isocode, country, city, ipclient, os, streamname, time_connect string
			var tiempo int
			err = query.Scan(&isocode, &country, &city, &ipclient, &os, &streamname, &tiempo)
			if err != nil {
				Warning.Println(err)
			}
			isocode = strings.ToLower(isocode)
			time_connect = secs2time(tiempo)
			fmt.Fprintf(w, "<tr><td>%s <img class='pull-right' src=\"images/flags/%s.png\" title=\"%s\"></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
				country, isocode, country, city, ipclient, streamname, os, time_connect)
		}
		query.Close()
		fmt.Fprintf(w, "</table>")
	}
}
