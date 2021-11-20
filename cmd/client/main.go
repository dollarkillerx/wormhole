package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var (
	localAddr      string // 本地地址
	remoteAddr     string // 远程地址
	certificateCrt string // 证书crt
	certificateKey string // 证书key
	debug          bool   // debug mode
)

func init() {
	fmt.Println("Wormhole Client")
	flag.StringVar(&localAddr, "l", "0.0.0.0:8487", "local addr")
	flag.StringVar(&remoteAddr, "r", "127.0.0.1:8087", "remote addr")
	flag.StringVar(&certificateCrt, "c", "proxy.crt", "proxy.crt")
	flag.StringVar(&certificateKey, "k", "proxy.key", "proxy.key")
	flag.BoolVar(&debug, "d", false, "debug")
	flag.Parse()
}

type server struct {
	conn   net.Conn
	read   chan []byte
	write  chan []byte
	exit   chan error
	reConn chan bool
}

func (s *server) Read() {
	s.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	for {
		buf := make([]byte, 10240)
		n, err := s.conn.Read(buf)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				s.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				s.conn.Write([]byte("0x"))
				if debug {
					log.Println("timeout")
				}
				break
			}

			log.Println(err)
			s.exit <- err
			break
		}

		if buf[0] == '0' && buf[1] == 'x' {
			if debug {
				log.Println("heartbeat")
			}
			continue
		}
		s.read <- buf[:n]
	}
}

func (s *server) Write() {
	for {
		select {
		case data := <-s.write:
			_, err := s.conn.Write(data)
			if err != nil {
				s.exit <- err
				break
			}
		}
	}
}

type localServer struct {
	conn  net.Conn
	read  chan []byte
	write chan []byte
	exit  chan error
}

func (l *localServer) Read() {
	for {
		buf := make([]byte, 10240)
		n, err := l.conn.Read(buf)
		if err != nil {
			l.exit <- err
			break
		}

		l.read <- buf[:n]
	}
}

func (l *localServer) Write() {
	for {
		select {
		case data := <-l.write:
			_, err := l.conn.Write(data)
			if err != nil {
				l.exit <- err
				break
			}
		}
	}
}

func main() {

}
