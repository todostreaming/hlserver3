package main

const (
	// config variables for the HTTP server
	rootdir                = "/var/segments/"                  // website root
	session           bool = true                              // session control by cookies enabled
	session_timeout        = 1200                              // timeout for a session (20 min?)
	first_page             = "index"                           // login page (default root page - always .html)
	enter_page             = "now.html"                        // Publisher user enter page after login
	enter_page_admin       = "monthly_admin.html"              // Admin user enter page after login
	http_port              = "80"                              // HTTP server port
	name_username          = "user"                            // name of input username in the login page
	name_password          = "password"                        // name of input password in the login page
	CookieName             = "GOSESSID"                        // cookie name used for user/admin sessions
	login_cgi              = "/login.cgi"                      // action cgi login in login page
	logout_cgi             = "/logout.cgi"                     // logout link at any page
	session_value_len      = 26                                // len of the value of the session cookie
	spanHTMLlogerr         = "<span id='loginerr'></span>"     // <span> where to publish the login error
	ErrorText              = "Login Error"                     // message to show when error login
	logFile                = "/var/log/hlserver.log"           // error logs file path
	DirDB                  = "/usr/local/bin/"                 // path for the empty original DB files
	DirRamDB               = "/var/db/"                        // path for the RAMdisk
	daylyDB                = "/usr/local/bin/dayly.db"         // dayly.sqlite DB
	monthlyDB              = "/usr/local/bin/monthly.db"       // mohtly.sqlite DB
	dirDaylys              = "/usr/local/bin/daylys/"          // path for dayly filled DBs
	dirMonthlys            = "/usr/local/bin/monthlys/"        // path for monthly filled DBs
	playingsRoot           = "/usr/local/bin/playings.reg"     // file with some settings inside
	dirGeoip               = "/usr/local/bin/GeoIP2-City.mmdb" // Maxmind GeoIP2 City database file's path
)
