package main

import (
	"fmt"
	"net/http"
	"strings"
)

// we will send the HTML5 player for a responsive bootstrap player:
/*
 <!-- 16:9 aspect ratio -->
<div class="embed-responsive embed-responsive-16by9">
  <iframe class="embed-responsive-item" src="http://hlserver/embed/rawstream"></iframe>
</div>

<!-- 4:3 aspect ratio -->
<div class="embed-responsive embed-responsive-4by3">
  <iframe class="embed-responsive-item" src="http://hlserver/embed/rawstream"></iframe>
</div>
*/
func embed(w http.ResponseWriter, r *http.Request) {
	// r.URL.Path[1:] = embed/luztv-livestream
	response := `
	<!DOCTYPE html>
	<html lang="en">
	
	<head>
	    <meta charset="utf-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body>
		<script src="http://domainplayers.org/js/hls.min.js"></script>
		<script src="http://domainplayers.org/js/html5play.min.js"></script>
		<video id="video_x890" controls width="100%" height="100%">
		<source id="src_x890">Your browser does not support HTML5 video. We recommend using <a href="https://www.google.es/chrome/browser/desktop/">Google Chrome</a></video>
		<script>var url = "%s";html5player(url, 1, "video_x890", "src_x890");</script>
	</body>
	</html>	
	`
	mu_cloud.RLock()
	domain := cloud["cloudserver"]
	mu_cloud.RUnlock()
	p := strings.Split(r.URL.Path[1:], "/")
	if len(p) != 2 {
		http.NotFound(w, r)
		return
	}
	url := "http://" + domain + "/" + p[1] + "-playlist.m3u8"
	response = fmt.Sprintf(response, url)

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Accept-Ranges", "bytes")
	fmt.Fprintln(w, response)
}
