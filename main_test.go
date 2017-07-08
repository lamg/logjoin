package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseLogins(t *testing.T) {
	var s string
	var lln *LLn
	var e error
	s = "Jun 27 21:05:41 proxy-profesores logportalauth[11593]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8"
	lln, e = parseLogins(s)
	if assert.NoError(t, e) {
		assert.True(t,
			lln.Action == "USERLOGIN" &&
				lln.User == "ymtnez" &&
				lln.IP == "10.2.9.8",
			"Got %v", lln,
		)
	}
}

func TestParseDownloads(t *testing.T) {
	var s string
	var dln *DLn
	var e error
	s = "1499345783.622      0 10.2.41.23 TCP_DENIED/403 4390 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html"
	dln, e = parseDownloads(s)
	if assert.NoError(t, e) {
		assert.True(t,
			dln.IP == "10.2.41.23" &&
				dln.Dwn == 4390 &&
				dln.URL == "http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png",
			"Got %v", dln,
		)
	}
}
