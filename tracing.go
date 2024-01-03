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
)

func getHttpTrace() *httptrace.ClientTrace {
	var (
		dnsStart, dnsEnd, connStart,
		connEnd, connectStart, connectEnd,
		tlsHandShakeStart time.Time
	)

	trace := &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			connStart = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			connEnd = time.Now()
			if !info.Reused {
				avgGotConn = append(avgGotConn, connEnd.Sub(connStart).Milliseconds())
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
				avgConnect = append(avgConnect, connectEnd.Sub(connectStart).Milliseconds())
			}
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			dnsEnd = time.Now()
			avgDns = append(avgDns, dnsEnd.Sub(dnsStart).Milliseconds())

		},
		TLSHandshakeStart: func() {
			tlsHandShakeStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if err != nil {
				log.Println("tls error", err)
			} else {
				avgTlsHandShake = append(avgTlsHandShake, time.Since(tlsHandShakeStart).Milliseconds())
			}

		},
		PutIdleConn: func(err error) {
			if err != nil {
				log.Println("error at putIdleConn", err)
			} else {
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
	log.Println("avg got conn", float64(gotConn)/float64(len(avgGotConn)), " ms")
	log.Println("new tcp connection open count", len(avgGotConn))

	for _, v := range avgConnect {
		connect += v
	}
	log.Println("avg got connect", float64(connect)/float64(len(avgConnect)), " ms")

	for _, v := range avgDns {
		dns += v
	}
	log.Println("avg dns", float64(dns)/float64(len(avgDns)), " ms")

	for _, v := range avgTlsHandShake {
		tlsHandshake += v
	}
	log.Println("avg tls handshake", float64(tlsHandshake)/float64(len(avgTlsHandShake)), " ms")

	for _, v := range avgTTFb {
		ttfb += v
	}
	log.Println("avg ttfb", float64(ttfb)/float64(len(avgTTFb)), " ms")
}
