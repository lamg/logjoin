// Watches the changes in two log files generated by a proxy
// server and uses that information for generating a file with
// Common Log Format.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"
	"unicode"
)

func main() {
	var lgn, dwn, out string
	flag.StringVar(&lgn, "l", "", "Logins log")
	flag.StringVar(&dwn, "d", "", "Downloads log")
	flag.StringVar(&out, "o", "", "Output log")
	flag.Parse()
	var lgR, dwR io.Reader
	var e error
	lgR, e = os.Open(lgn)
	if e == nil {
		dwR, e = os.Open(dwn)
	}
	var otW io.Writer
	if e == nil {
		otW, e = os.Create(out)
	}
	if e == nil {
		e = joinLns(lgR, dwR, otW)
	}
	if e != nil {
		log.Fatal(e.Error())
	}
}

type DLn struct {
	IP, URL string
	Dwn     int
	Time    time.Time
}

type Session struct {
	user       string
	start, end time.Time
	dwnls      []*DLn
	closed     bool
}

func (s *Session) Belongs(t time.Time) (r bool) {
	r = s.start.Before(t) && s.end.After(t)
	return
}

func (s *Session) AppendDwn(d *DLn) {
	var i int
	i = 0
	for i != len(s.dwnls) && !s.dwnls[i].Time.Equal(d.Time) {
		i = i + 1
	}
	if i == len(s.dwnls) {
		s.dwnls = append(s.dwnls, d)
	}
}

// { l contains lines with users associated to IP addresses
//   and d contains lines with IP addresses associated to
//   download requests }
// { o contains lines with users associated to IP ddresses
//   and download requests, in Common Log Format }
func joinLns(l, d io.Reader, o io.Writer) (e error) {
	var lgi map[string][]*Session
	lgi = make(map[string][]*Session)
	e = delimSessions(l, lgi)
	// { sessions delimited }
	if e == nil {
		e = fillSessions(d, lgi)
	}
	// { sessions filled ≡ e = nil }
	if e == nil {
		e = writeDwns(o, lgi)
	}
	// { downloads per user written ≡ e = nil }
	return
}

func delimSessions(l io.Reader, lgi map[string][]*Session) (e error) {
	var ls *bufio.Scanner
	ls = bufio.NewScanner(l)
	for ls.Scan() {
		var ln string
		ln = ls.Text()
		var lln *LLn
		lln, e = parseLogins(ln)
		if e == nil {
			var ss []*Session
			var ns *Session
			var ok bool
			ss, ok = lgi[lln.IP]
			if lln.IsLogin() {
				if !ok {
					lgi[lln.IP] = make([]*Session, 0)
				}
				// { lgi[lln.IP] is initialized }
				ns = &Session{
					start: lln.Time,
					user:  lln.User,
				}
				if len(lgi[lln.IP]) != 0 && !lgi[lln.IP][len(lgi[lln.IP])-1].closed {
					// { a LOGIN from the IP of an unfinished session
					//   is made }
					var drt time.Duration
					drt, _ = time.ParseDuration("-1ms")
					lgi[lln.IP][len(lgi[lln.IP])-1].end = ns.start.Add(drt)
					lgi[lln.IP][len(lgi[lln.IP])-1].closed = true
					// { finished previous session }
				}
				lgi[lln.IP] = append(lgi[lln.IP], ns)
			} else if lln.IsLogout() {
				if ok {
					var i int
					for i != len(ss) && (ss[i].user != lln.User ||
						ss[i].closed) {
						i++
					}
					// { ns is last opened session of lln.User ≡
					//   foundSession}
					if i != len(ss) {
						ss[i].end = lln.Time
						ss[i].closed = true
					}
					// { closed lln.User's session ≡ foundSession }
				}
			}
		}
		// { sessionOpened ∨ sessionClosed ≡ e = nil }
	}
	e = ls.Err()
	if e == bufio.ErrTooLong {
		e = nil
	}
	return
}

func fillSessions(d io.Reader, lgi map[string][]*Session) (e error) {
	var ds *bufio.Scanner
	ds = bufio.NewScanner(d)
	for ds.Scan() {
		var ln string
		ln = ds.Text()
		var dln *DLn
		dln, _ = parseDownloads(ln)
		var s []*Session
		var ok bool
		s, ok = lgi[dln.IP]
		if ok {
			var i int
			i = 0
			for i != len(s) && !s[i].Belongs(dln.Time) {
				i = i + 1
			}
			if i != len(s) {
				s[i].AppendDwn(dln)
			}
		}
	}
	e = ds.Err()
	if e == bufio.ErrTooLong {
		e = nil
	}
	return
}

func writeDwns(o io.Writer, lgi map[string][]*Session) (e error) {
	for _, v := range lgi {
		for _, j := range v {
			var bs []byte
			bs = sessionToBytes(j)
			_, e = o.Write(bs)
		}
	}
	return
}

func sessionToBytes(j *Session) (bs []byte) {
	bs = make([]byte, 0)
	for _, k := range j.dwnls {
		var s string
		s = fmt.Sprintf(
			"%s - %s [%s] \"GET HTTP/1.0\" 200 %d\n",
			k.IP, j.user, k.Time.String(), k.Dwn,
		)
		bs = append(bs, []byte(s)...)
	}
	return
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
func parseLogins(l string) (r *LLn, e error) {
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

func getWord(s string, i int) (r string, k int, e error) {
	k = i
	for k != len(s) && (unicode.IsLetter(rune(s[k])) || s[k] == '-' || s[k] == '_') {
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

// { l has format:
//   line = number '.' number number ip word '/' number size
//      word url '-' word '/-' word '/' word.
// }
func parseDownloads(l string) (r *DLn, e error) {
	var i int
	r = new(DLn)
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
		i, e = skipWord(l, i)
	}
	if e == nil {
		for i != len(l) && !unicode.IsSpace(rune(l[i])) {
			// { r.URL as a sequence of non-blank characters }
			r.URL, i = r.URL+string(l[i]), i+1
		}
		// { nothing more to parse }
	}
	return
}

func toString(d *DLn, l *LLn) (r string) {
	//TODO
	r = fmt.Sprintf("%s %s %s %d %s", time.Now().String(), l.User,
		l.IP, d.Dwn, d.URL)
	return
}
