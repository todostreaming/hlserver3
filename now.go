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
	username, ok := user_[key] // De aquí podemos recoger el id del usuario logeado
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

	response := "<table class=\"table table-hover table-condensed\"><thead class=\"bg-danger\"><tr class=\"row\"><th class=\"col-1\">&nbsp;</th><th class=\"hidden-xs col-sm-4\">Info</th><th class=\"hidden-xs col-sm-3\">IP</th><th class=\"col-xs-6 col-sm-2\">Stream</th><th class=\"col-1\">&nbsp;</th><th class=\"col-xs-4 col-sm-2\">Time</th></tr></thead><tbody>"
	counter := 0
	anio, mes, dia := time.Now().Date()
	fecha := fmt.Sprintf("%d/%02d/%02d", anio, mes, dia)
	hh, mm, _ := time.Now().Clock()
	hora := fmt.Sprintf("%02d:%02d", hh, mm)
	tiempo_limite := time.Now().Unix() - 6 //tiempo limite de 6 seg
	query, err := dblive.Query("SELECT streamname, isocode, ip, country, time, bitrate, info FROM encoders WHERE username = ? AND timestamp > ?", username, tiempo_limite)
	if err != nil {
		Error.Println(err)
		return
	}
	for query.Next() {
		counter++
		var isocode, country, streamname, ip, time_connect, info, shortstreamname string
		var tiempo, bitrate int
		query.Scan(&streamname, &isocode, &ip, &country, &tiempo, &bitrate, &info)
		isocode = strings.ToLower(isocode)
		time_connect = secs2time(tiempo)
		INFO := fmt.Sprintf("%s [%d kbps]", info, bitrate/1000)
		if len(streamname) > 16 {
			shortstreamname = streamname[0:15]
		}
		response = response + fmt.Sprintf("<tr class=\"row\"><td class=\"col-1\"><a href=\"javascript:launchRemote('play.cgi?stream=%s')\"><img src=\"images/play.jpg\" border=\"0\"/></a></td><td class=\"hidden-xs col-sm-4\">%s</td><td class=\"hidden-xs col-sm-3\">%s</td><td class=\"col-xs-6 col-sm-2\">%s</td><td class=\"col-1\"><img src=\"images/flags/%s.png\"/></td><td class=\"col-xs-4 col-sm-2\">%s</td></tr>",
			streamname, INFO, ip, shortstreamname, isocode, time_connect)
	}
	query.Close()
	// construir la respuesta
	response = fmt.Sprintf("<h4 class=\"text-center\">Connected the day %s at UTC %s <span class=\"visible-xs\"><small>(%s)</small></span></h4><h5 class=\"text-center text-danger\"><strong>%d encoders connected</strong></h5>", fecha, hora, username, counter) + response + "</tbody></table></div>"
	fmt.Fprintf(w, response)
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

	var contador, counter int
	response := ""
	tiempo_limite := time.Now().Unix() - 30 //tiempo limite de 30 seg
	err = dblive.QueryRow("SELECT count(id) FROM players WHERE username = ? AND timestamp > ? AND time > 0", username, tiempo_limite).Scan(&contador)
	if err != nil {
		Error.Println(err)
		return
	}
	if contador >= 100 {
		counter++
		query, err := dblive.Query("SELECT isocode, country, count(ipclient) AS count, streamname FROM players WHERE username = ? AND timestamp > ? AND time > 0 GROUP BY isocode, streamname ORDER BY streamname, count DESC", username, tiempo_limite)
		if err != nil {
			Error.Println(err)
			return
		}
		for query.Next() {
			var isocode, country, ips, streamname, shortstreamname string
			query.Scan(&isocode, &country, &ips, &streamname)
			isocode = strings.ToLower(isocode)
			if len(streamname) > 16 {
				shortstreamname = streamname[0:15]
			}
			response = response + fmt.Sprintf("<tr class=\"row\"><td class=\"hidden-xs col-sm-4\">%s</td><td class=\"col-2\"><img src=\"images/flags/%s.png\" title=\"%s\"/></td><td class=\"col-xs-5 col-sm-3\">25</td><td class=\"col-xs-5 col-sm-4\">%s</td></tr>",
				country, isocode, country, shortstreamname)
		}
		query.Close()
	} else {
		query, err := dblive.Query("SELECT isocode, country, city, ipclient, os, streamname, time FROM players WHERE username = ? AND timestamp > ? AND time > 0 ORDER BY streamname, time DESC", username, tiempo_limite)
		if err != nil {
			Warning.Println(err)
		}
		for query.Next() {
			var isocode, country, city, ipclient, os, streamname, time_connect, shortstreamname, shortcountry string
			var tiempo int
			err = query.Scan(&isocode, &country, &city, &ipclient, &os, &streamname, &tiempo)
			if err != nil {
				Warning.Println(err)
			}
			isocode = strings.ToLower(isocode)
			time_connect = secs2time(tiempo)
			if len(streamname) > 16 {
				shortstreamname = streamname[0:15]
			}
			if len(country) > 16 {
				shortcountry = country[0:15]
			}
			response = response + fmt.Sprintf("<tr class=\"row\"><td class=\"hidden-xs col-sm-2\">%s</td><td class=\"col-1\"><img src=\"images/flags/%s.png\" title=\"%s\"/></td><td class=\"hidden-xs col-sm-2\">%s</td><td class=\"hidden-xs col-sm-2\">%s</td><td class=\"col-xs-6 col-sm-2\">%s</td><td class=\"col-1\"><img src=\"images/os/%s.png\"/></td><td class=\"col-xs-4 col-sm-2\">%s</td></tr>",
				shortcountry, isocode, country, city, ipclient, shortstreamname, os, time_connect)
		}
		query.Close()
	}
	// creamos la salida
	if contador >= 100 {
		response = fmt.Sprintf("<h5 class=\"text-center text-primary\"><strong>%d players connected</strong></h5><table class=\"table table-hover table-condensed\"><thead class=\"bg-primary\"><tr class=\"row\"><th class=\"hidden-xs col-sm-2\">Country</th><th class=\"col-1\">&nbsp;</th><th class=\"hidden-xs col-sm-2\">City</th><th class=\"hidden-xs col-sm-2\">IP</th><th class=\"col-xs-6 col-sm-2\">Stream</th><th class=\"col-1\">OS</th><th class=\"col-xs-4 col-sm-2\">Time</th></tr></thead><tbody>", contador) + response + "</tbody></table></div>"
	} else {
		response = fmt.Sprintf("<div class=\"container\"><h5 class=\"text-center text-primary\"><strong>%d players connected from %d countries</strong></h5><table class=\"table table-hover table-condensed\"><thead class=\"bg-primary\"><tr class=\"row\"><th class=\"hidden-xs col-sm-4\">Country</th><th class=\"col-2\">&nbsp;</th><th class=\"col-xs-5 col-sm-3\">Players</th><th class=\"col-xs-5 col-sm-4\">Stream</th></tr></thead><tbody>", contador, counter) + response + "</tbody></table></div>"
	}
	fmt.Fprintf(w, response)

}
