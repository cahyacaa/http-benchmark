package main

import (
	"crypto/tls"
	"log"
	"net/http/httptrace"
	"time"
)

var (
	avgGotConn              []int64
	avgConnect              []int64
	avgDns, avgTlsHandShake []int64
	avgTTFb                 []int64
	l                       []int64
)

func getHttpTrace() *httptrace.ClientTrace {
	var (
		dnsStart, dnsEnd, connStart,
		connEnd, connectStart, connectEnd,
		tlsHandShakeStart, tlsHandShakeEnd time.Time
	)

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			connStart = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			connEnd = time.Now()
			if info.Reused {
				log.Println("connection reused", info.WasIdle, info.IdleTime)
			} else {
				avgGotConn = append(avgGotConn, connEnd.Sub(connStart).Microseconds())

			}

		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()

		},
		ConnectDone: func(network, addr string, err error) {
			connectEnd = time.Now()
			if err != nil {
				log.Println("error at ConnectDone", err)

			} else {
				avgConnect = append(avgConnect, connectEnd.Sub(connectStart).Microseconds())
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
			avgDns = append(avgDns, dnsEnd.Sub(dnsStart).Microseconds())

		},
		TLSHandshakeStart: func() {
			tlsHandShakeStart = time.Now()
			l = append(l, int64(1))
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				log.Println("tls error", err)

			} else {
				tlsHandShakeEnd = time.Now()
				avgTlsHandShake = append(avgTlsHandShake, tlsHandShakeEnd.Sub(tlsHandShakeStart).Microseconds())
			}

		},
		PutIdleConn: func(err error) {
			if err != nil {
				log.Println("error at putIdleConn", err)
			} else {
				log.Println("put idle connection")
			}

		},
		GotFirstResponseByte: func() {
			avgTTFb = append(avgTTFb, time.Since(connStart).Milliseconds())
		},
	}

	return trace

}

// finding average of each operation
func findAvg() {
	var (
		gotConn, connect, dns, tlsHandshake, ttfb int64
	)
	for _, v := range avgGotConn {
		gotConn += v
	}
	log.Println("avg got conn", float64(gotConn)/float64(len(avgGotConn)), len(avgGotConn))

	for _, v := range avgConnect {
		connect += v
	}
	log.Println("avg got connect", float64(connect)/float64(len(avgConnect)))

	for _, v := range avgDns {
		dns += v
	}
	log.Println("avg dns", float64(dns)/float64(len(avgDns)))

	for _, v := range avgTlsHandShake {
		tlsHandshake += v
	}
	log.Println("avg tls handshake", float64(tlsHandshake)/float64(len(avgTlsHandShake)), len(l))

	for _, v := range avgTTFb {
		ttfb += v
	}
	log.Println("avg ttfb", float64(ttfb)/float64(len(avgTTFb)))
}
