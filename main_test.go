package main

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/require"
	"io"
	"strings"
	"testing"
	"time"
)

func TestParseLogins(t *testing.T) {
	var s string
	var lln *LLn
	var e error
	var tm time.Time
	tm, e = time.Parse(time.Stamp, "Jun 27 21:05:41")
	require.NoError(t, e)
	s = "Jun 27 21:05:41 proxy-profesores logportalauth[11593]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8"
	lln, e = parseLogins(s)
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
	dln, e = parseDownloads(s)
	require.NoError(t, e)
	require.True(t,
		dln.IP == "10.2.41.23" &&
			dln.Dwn == 4390 &&
			dln.URL == "http://proxy-profesores.upr.edu.cu/squid-internal-static/icons/SN.png",
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
		dl, e = parseDownloads(ss[i])
		if e == nil {
			ds, i = append(ds, dl), i+1
		}
	}
	require.NoError(t, e, "i: %d", i)
}

func TestSessionToBytes(t *testing.T) {
	var ns *Session
	var w time.Time
	w, _ = time.Parse(time.UnixDate,
		"Mon Jan 2 15:04:05 CDT 2006")
	var d time.Duration
	d, _ = time.ParseDuration("30s")
	ns = &Session{
		user:  "lamg",
		start: w,
		end:   w.Add(d).Add(d).Add(d),
		dwnls: []*DLn{
			&DLn{
				IP:   "10.2.23.1",
				URL:  "https://google.com.cu",
				Dwn:  4020,
				Time: w.Add(d),
			},
			&DLn{
				IP:   "10.2.23.1",
				URL:  "https://en.wikipedia.com",
				Dwn:  2030,
				Time: w.Add(d).Add(d),
			},
		},
	}
	var bs []byte
	bs = sessionToBytes(ns)
	var exLns []string
	exLns = []string{"10.2.23.1 - lamg [2006-01-02 15:04:35 -0400 CDT] \"GET HTTP/1.0\" 200 4020",
		"10.2.23.1 - lamg [2006-01-02 15:05:05 -0400 CDT] \"GET HTTP/1.0\" 200 2030",
	}
	var br io.Reader
	br = bytes.NewReader(bs)
	var sc *bufio.Scanner
	sc = bufio.NewScanner(br)
	var i int
	i = 0
	for sc.Scan() {
		require.True(t, sc.Text() == exLns[i], "%s ≠ %s",
			sc.Text(), exLns[i])
		i = i + 1
	}
	require.True(t, i == len(exLns))
}

func TestDelimSessions(t *testing.T) {
	var lgi map[string][]*Session
	lgi = make(map[string][]*Session)
	var bf *bytes.Buffer
	bf = bytes.NewBufferString(l)
	var e error
	e = delimSessions(bf, lgi)
	require.NoError(t, e)
	require.True(t, len(lgi) > 0)
}

func TestFillSessions(t *testing.T) {
	var e error
	var lgi map[string][]*Session
	lgi = make(map[string][]*Session)
	var bf *bytes.Buffer
	bf = bytes.NewBufferString(l)
	e = delimSessions(bf, lgi)
	require.NoError(t, e)
	var bd *bytes.Buffer
	bd = bytes.NewBufferString(d)
	e = fillSessions(bd, lgi)
	require.NoError(t, e)
	var ss *Session
	ss = lgi["10.2.9.8"][0]
	require.True(t, len(ss.dwnls) > 0,
		"user: %s start: %s end: %s ds: %v",
		ss.user,
		ss.start.String(),
		ss.end.String(),
		ss.dwnls,
	)
}

func TestJoinLns(t *testing.T) {
	var lr, dr io.Reader
	lr, dr = strings.NewReader(l), strings.NewReader(d)
	var ow *bytes.Buffer
	var s string
	ow = bytes.NewBufferString(s)
	var e error
	e = joinLns(lr, dr, ow)
	require.NoError(t, e)
	require.True(t, ow.Len() > 0)
}

//TODO include year in l's lines
var (
	l = `Jun 27 21:04:55 proxy-profesores logportalauth[8212]: Zone: proxy_profes - Reconfiguring captive portal(Proxy_Profes).
Jul 06 06:05:41 proxy-profesores logportalauth[11593]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8
Jul 06 21:06:13 proxy-profesores logportalauth[48372]: Zone: proxy_profes - DISCONNECT: ymtnez, , 10.2.9.8
Jun 27 21:09:38 proxy-profesores logportalauth[25822]: Zone: proxy_profes - Reconfiguring captive portal(Proxy_Profes).
Jun 27 21:09:57 proxy-profesores logportalauth[47811]: Zone: proxy_profes - USER LOGIN: ymtnez, , 10.2.9.8
Jun 27 21:10:13 proxy-profesores logportalauth[73868]: Zone: proxy_profes - DISCONNECT: ymtnez, , 10.2.9.8`
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
