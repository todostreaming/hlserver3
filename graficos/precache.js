load();
setInterval(load,10000); // every 10 secs  
function load(){
	// all the info here: https://www.w3schools.com/js/js_ajax_http.asp
	var xhttp = new XMLHttpRequest(); // GET request
	xhttp.onreadystatechange = function() {
	    if (this.readyState == 4 && this.status == 200) {
		      //now lets make the HEAD request
		      var xhttp2 = new XMLHttpRequest(); // HEAD request after
		      xhttp2.onreadystatechange = function() {
		    	  if (this.readyState == 4 && this.status == 200) {
		    		  // OK done	
		    	  }
		      }
		      xhttp2.open("HEAD", "http://mydomain/live/" + this.responseText, true);
		      xhttp2.send();
	    }
	};
	xhttp.open("GET", "http://mydomain/rawstream.lst", true);
	xhttp.send();
}
