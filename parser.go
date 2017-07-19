package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type DLn struct {
	IP, URL, Method string
	Dwn             int
	Time            time.Time
	Written         bool
}

type Session struct {
	user       string
	start, end time.Time
	closed     bool
}

type LLn struct {
	Action, IP, User string
	Time             time.Time
}

func (l *LLn) IsLogin() (b bool) {
	b = l.Action == "USERLOGIN"
	return
}

func (l *LLn) IsLogout() (b bool) {
	b = l.Action == "DISCONNECT"
	return
}

// { l has format:
//  line = month day time host pid zone action user cs ip.
//  month = word.
//  day = number.
//  time = number colon number colon number
//  host = word.
//  pid = word "[" number "]" colon.
//  zone = "Zone" colon word "-".
//  action = word [word] colon.
//  user = word comma comma.
//  ip = number dot number dot number dot number.
// }
func parseUsrEvt(l string) (r *LLn, e error) {
	var i int
	i, r = 0, new(LLn)
	if e == nil {
		i, e = skipWord(l, i)
	}
	// { month skipped ≡ e = nil }
	if e == nil {
		i, e = skipNumber(l, i)
	}
	// { day skipped ≡ e = nil }
	if e == nil {
		i, e = skipNumber(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, ':')
	}
	if e == nil {
		i, e = skipNumber(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, ':')
	}
	if e == nil {
		i, e = skipDigits(l, i)
	}
	if e == nil {
		r.Time, e = time.ParseInLocation(time.Stamp, l[:i],
			time.Now().Location())
	}
	// { time skipped ≡ e = nil }
	if e == nil {
		r.Time = r.Time.AddDate(time.Now().Year(), 0, 0)
		i = skipSpaces(l, i)
		i, e = skipWord(l, i)
	}
	// { host skipped ≡ e = nil }
	if e == nil {
		i, e = skipWord(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, '[')
	}
	if e == nil {
		i, e = skipNumber(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, ']')
	}
	if e == nil {
		i, e = skipChar(l, i, ':')
	}
	if e == nil {
		i, e = skipWord(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, ':')
	}
	if e == nil {
		i, e = skipWord(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, '-')
	}
	// { zone skipped ≡ e = nil }
	if e == nil {
		r.Action, i, e = getAction(l, i)
	}
	if e == nil {
		r.User, i, e = getWord(l, i)
	}
	// { got user ≡ e = nil }
	if e == nil {
		i, e = skipChar(l, i, ',')
	}
	if e == nil {
		i, e = skipChar(l, i, ',')
	}
	// { ", ," skipped ≡ e = nil }
	if e == nil {
		r.IP, i, e = getIP(l, i)
	}
	// { got IP ≡ e = nil }
	return
}

// { l has format:
//   line = number '.' number number ip word '/' number size
//      word url '-' word '/-' word '/' word.
// }
func parseDownload(l string) (r *DLn, e error) {
	var i int
	r = &DLn{Written: false}
	i = 0
	r.Time, i, e = getTime(l, i)
	if e == nil {
		i, e = skipNumber(l, i)
	}
	if e == nil {
		r.IP, i, e = getIP(l, i)
	}
	if e == nil {
		i = skipSpaces(l, i)
		i, e = skipWord(l, i)
	}
	if e == nil {
		i, e = skipChar(l, i, '/')
	}
	if e == nil {
		i, e = skipNumber(l, i)
	}
	var k int
	if e == nil {
		k, e = skipDigits(l, i)
	}
	if e == nil {
		var sz string
		sz = l[i:k]
		r.Dwn, e = strconv.Atoi(sz)
		k = skipSpaces(l, k)
	}
	if e == nil {
		i = k
		r.Method, i, e = getWord(l, i)
		// { got Method }
	}
	if e == nil {
		for i != len(l) && !unicode.IsSpace(rune(l[i])) {
			// { r.URL as a sequence of non-blank characters }
			r.URL, i = r.URL+string(l[i]), i+1
		}
		// { nothing more to parse }
	}
	if r.IP == "10.2.9.17" {
		println(r.Time.String())
	}
	if strings.Contains(l, "10.2.9.17") {
		println("contains")
	}
	println(l)
	return
}

func getWord(s string, i int) (r string, k int, e error) {
	k = i
	for k != len(s) && (unicode.IsLetter(rune(s[k])) ||
		s[k] == '-' || s[k] == '_' || s[k] == '.') {
		k = k + 1
	}
	if k != len(s) && i != k {
		r = s[i:k]
		k = skipSpaces(s, k)
	} else {
		e = fmt.Errorf("Error in %s at %d", s, k)
	}
	return
}

func getIP(s string, i int) (r string, k int, e error) {
	k, e = skipDigits(s, i)
	if e == nil {
		r = s[i:k]
		i = k
		k, e = skipChar(s, i, '.')
	}
	if e == nil {
		r += "."
		i = k
		k, e = skipDigits(s, i)
	}
	if e == nil {
		r += s[i:k]
		i = k
		k, e = skipChar(s, i, '.')
	}
	if e == nil {
		r += "."
		i = k
		k, e = skipDigits(s, i)
	}
	if e == nil {
		r += s[i:k]
		i = k
		k, e = skipChar(s, i, '.')
	}
	if e == nil {
		r += "."
		i = k
		k, e = skipDigits(s, i)
	}
	if e == nil {
		r += s[i:k]
		k = skipSpaces(s, k)
	}
	return
}

func getAction(s string, i int) (r string, k int, e error) {
	r, k, e = getWord(s, i)
	if e == nil {
		i = k
		k, e = skipChar(s, i, ':')
	}
	if e != nil {
		i = skipSpaces(s, k)
		var r0 string
		r0, k, e = getWord(s, i)
		if e == nil {
			r = r + r0
			k, e = skipChar(s, k, ':')
		}
	}
	return
}

func skipChar(s string, i int, c rune) (r int, e error) {
	r = i
	if rune(s[i]) == c {
		r = i + 1
	} else {
		e = fmt.Errorf("Expected char %c at %d in %s", c, i, s)
	}
	r = skipSpaces(s, r)
	return
}

func skipDigits(s string, i int) (r int, e error) {
	r = i
	for r != len(s) && unicode.IsDigit(rune(s[r])) {
		r = r + 1
	}
	if r == i {
		e = fmt.Errorf("Error in %s at %d", s, i)
	}
	return
}

func skipNumber(s string, i int) (r int, e error) {
	r, e = skipDigits(s, i)
	if e == nil {
		r = skipSpaces(s, r)
	}
	return
}

// { 0 ≤ i < len(s) }
// { r = index of char seq followed by (letter seq followed by
//   space seq) }
func skipWord(s string, i int) (r int, e error) {
	r = i
	for r != len(s) && (unicode.IsLetter(rune(s[r])) || s[r] == '-' || s[r] == '_') {
		r = r + 1
	}
	if r == i {
		e = fmt.Errorf("Expected letter at %d in %s", i, s)
	} else {
		r = skipSpaces(s, r)
	}
	return
}

func skipSpaces(s string, i int) (r int) {
	for i != len(s) && unicode.IsSpace(rune(s[i])) {
		i = i + 1
	}
	r = i
	return
}

func getTime(l string, i int) (t time.Time, k int, e error) {
	var r int64
	k, e = skipDigits(l, i)
	if e == nil {
		r, e = strconv.ParseInt(l[i:k], 10, 64)
	}
	if e == nil {
		k, e = skipChar(l, k, '.')
	}
	if e == nil {
		i = k
		k, e = skipDigits(l, i)
	}
	var ms int64
	if e == nil {
		ms, e = strconv.ParseInt(l[i:k], 10, 64)
	}
	if e == nil {
		k = skipSpaces(l, k)
		t = time.Unix(r, ms*1000000)
	}
	return
}
