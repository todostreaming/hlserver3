package main

import (
	"database/sql"
	"github.com/oschwald/geoip2-golang"
	"golang.org/x/sync/syncmap"
	"log"
	"sync"
)

var (
	ident    int64      // identifier for every streaming session openned by an individual player
	mu_ident sync.Mutex // exclusive mutex for the identifier
	// DB Live vars
	dblive    *sql.DB    // db only with live players raw info
	mu_dblive sync.Mutex // also exclusive mutex for
	// GeoIP2 vars
	dbgeoip    *geoip2.Reader
	mu_dbgeoip sync.Mutex
	// error loggers
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	// bandwidths map of the rawstreams (encoders) in bps
	// dont forget to empty it at the end of the day after dayly resume
	Bw_int *syncmap.Map
	// referer map ( ["rawstream"] = "domain1.com;domain2.com" )
	Referer *syncmap.Map
	// forecasters map ( ["near_proxy=rawstream"] = UNIXtimestamp_int64 )
	Forecaster *syncmap.Map
)