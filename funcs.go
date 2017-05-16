package main

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

// separa la IPv4/6 del puerto usado con la misma
func getip(pseudoip string) string {
	var res string
	if strings.Contains(pseudoip, "]:") {
		part := strings.Split(pseudoip, "]:")
		res = part[0]
		res = res[1:]
	} else {
		part := strings.Split(pseudoip, ":")
		res = part[0]
	}
	return res
}

// convierte un string numérico en un entero int
func toInt(cant string) (res int) {
	res, _ = strconv.Atoi(cant)
	return
}

func random(min, max int) int { // [min,max)
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Intn(max-min) + min
}

func geoIP(ip_parsing string) (country, isocode, city string) {
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(ip_parsing)
	record, err := dbgeoip.City(ip)
	if err != nil {
		return
	}
	city = record.City.Names["en"]
	country = record.Country.Names["en"]
	isocode = record.Country.IsoCode

	return country, isocode, city
}

// from a complete url
func getdomain(url string) string {
	var domain string

	p := strings.Split(url, "/")
	if len(p) > 2 {
		domain = p[2]
	}

	return domain
}

func controlinternalsessions() {
	for {
		for k, v := range time_ {
			if time.Since(v).Seconds() > 0 { // it is negative up to expiration time
				mu_user.Lock()
				delete(user_, k)
				delete(time_, k)
				delete(type_, k)
				mu_user.Unlock()
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// genera una session id o Value del Cookie aleatoria y de la longitud que se quiera
func sessionid(r *rand.Rand, n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}
