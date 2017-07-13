package main

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestParseLogins(t *testing.T) {
	var s string
	var lln *LLn
	var e error
	var tm time.Time
	tm, e = time.ParseInLocation(time.Stamp, "Jun 27 21:05:41",
		time.Now().Location())
	require.NoError(t, e)
	tm = tm.AddDate(time.Now().Year(), 0, 0)
	s = "Jun 27 21:05:41 proxy-profesores logportalauth[11593]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8"
	lln, e = parseLogin(s)
	require.NoError(t, e)
	require.True(t,
		lln.Action == "USERLOGIN" &&
			lln.User == "ymtnez" &&
			lln.IP == "10.2.9.8" &&
			lln.Time.Equal(tm),
		"Got %v", lln,
	)
}

func TestParseDownloads(t *testing.T) {
	var s string
	var dln *DLn
	var e error
	s = "1499345783.622      0 10.2.41.23 TCP_DENIED/403 4390 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html"
	dln, e = parseDownload(s)
	require.NoError(t, e)
	require.True(t,
		dln.IP == "10.2.41.23" &&
			dln.Dwn == 4390 &&
			dln.URL == "http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png",
		dln.Method == "GET",
		dln.Time.Equal(time.Unix(1499345783, 0)),
		"Got %v", dln,
	)
}

func TestParseDownloads0(t *testing.T) {
	var ss []string
	ss = strings.Split(d, "\n")
	var ds []*DLn
	ds = make([]*DLn, 0, len(ss))
	var e error
	var i int
	for i != len(ss)-1 && e == nil {
		var dl *DLn
		dl, e = parseDownload(ss[i])
		if e == nil {
			ds, i = append(ds, dl), i+1
		}
	}
	require.NoError(t, e, "i: %d", i)
	/*for _, j := range ds {
		t.Logf("ip: %s time: %s", j.IP, j.Time.String())
	}*/
}

func TestLogProc(t *testing.T) {
	var lb, db *bytes.Buffer
	lb, db = bytes.NewBufferString(l), bytes.NewBufferString(d)
	var lc, dc, out chan string
	lc, dc, out = make(chan string), make(chan string),
		make(chan string)
	go buffChan(lb, lc)
	go buffChan(db, dc)
	go logProc(lc, dc, out)
	var b bool
	var c int
	c = 0
	_, b = <-out
	for b {
		c = c + 1
		_, b = <-out
	}
	require.True(t, c == 9, "%d ≠ 9", c)
}

func buffChan(bf *bytes.Buffer, cs chan<- string) {
	var e error
	for e == nil {
		var ls string
		ls, e = bf.ReadString('\n')
		if e == nil {
			cs <- ls
		} else {
			close(cs)
		}
	}
}

// func TestPrintTimes(t *testing.T) {
// 	var t0, t1, t2 time.Time
// 	t1 = time.Unix(1499717148, 206*1000000)
// 	t0 = time.Unix(1499659201, 020*1000000)
// 	t.Logf("t0: %s", t0.String())
// 	t.Logf("t1: %s", t1.String())
// 	t2, _ = time.ParseInLocation(time.Stamp, "Jun 28 07:20:48",
// 		time.Local)
// 	t2 = t2.AddDate(2017, 0, 0)
// 	t.Logf("%d", t2.Unix())

// }

var (
	l = `Jul 06 06:05:41 proxy-profesores logportalauth[11593]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8
Jul 06 21:06:13 proxy-profesores logportalauth[48372]: Zone: proxy_profes - DISCONNECT: ymtnez, , 10.2.9.8
Jul 06 21:09:38 proxy-profesores logportalauth[25822]: Zone: proxy_profes - Reconfiguring captive portal(Proxy_Profes).
Jul 06 21:09:57 proxy-profesores logportalauth[47811]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8
Jul 06 21:10:13 proxy-profesores logportalauth[73868]: Zone: proxy_profes - DISCONNECT: ymtnez, , 10.2.9.8`
	d = `1499336852.856      1 212.237.54.71 TAG_NONE/400 4007 GET / - HIER_NONE/- text/html
1499344036.349      0 10.2.9.8 TAG_NONE/400 4004 GET / - HIER_NONE/- text/html
1499344036.376      1 10.2.9.8 TCP_DENIED/403 4395 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html
1499344036.451      0 10.2.9.8 TAG_NONE/400 4026 GET /favicon.ico - HIER_NONE/- text/html
1499345783.538      0 10.2.9.8 TAG_NONE/400 4004 GET / - HIER_NONE/- text/html
1499345783.622      0 10.2.9.8 TCP_DENIED/403 4390 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html
1499345783.634      0 10.2.9.8 TAG_NONE/400 4026 GET /favicon.ico - HIER_NONE/- text/html
1499345783.646      0 10.2.9.8 TAG_NONE/400 4026 GET /favicon.ico - HIER_NONE/- text/html
1499346127.473      0 10.2.74.201 TAG_NONE/400 4005 GET / - HIER_NONE/- text/html
1499346127.561      0 10.2.74.201 TCP_DENIED/403 4340 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html
1499346127.684      0 10.2.74.201 TAG_NONE/400 4027 GET /favicon.ico - HIER_NONE/- text/html
1499346127.753      0 10.2.74.201 TAG_NONE/400 4027 GET /favicon.ico - HIER_NONE/- text/html
1499354123.721      0 10.2.9.8 TAG_NONE/400 4004 GET / - HIER_NONE/- text/html
1499354123.729      0 10.2.9.8 TCP_DENIED/403 4449 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html
1499362631.526      0 10.2.132.12 TAG_NONE/400 4005 GET / - HIER_NONE/- text/html
1499362631.793      1 10.2.132.12 TCP_DENIED/403 4400 GET http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png - HIER_NONE/- text/html
1499362632.070      0 10.2.132.12 TAG_NONE/400 4027 GET /favicon.ico - HIER_NONE/- text/html
1499362632.087      0 10.2.132.12 TAG_NONE/400 4027 GET /favicon.ico - HIER_NONE/- text/html
1499362634.598      0 10.2.132.12 TAG_NONE/400 4005 GET / - HIER_NONE/- text/html
`
)
