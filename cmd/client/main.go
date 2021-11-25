package main

import (
	"crypto/rand"
	"crypto/tls"
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
	flag.StringVar(&localAddr, "l", "192.168.88.202:8001", "local addr 本地穿透地址")
	flag.StringVar(&remoteAddr, "r", "127.0.0.1:8087", "remote addr 远程服务地址")
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
	//s.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
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
				continue
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
	if debug {
		log.SetFlags(log.LstdFlags | log.Llongfile)
	}

	// 1. 初始化主要链接
	conn, err := newConn(remoteAddr)
	if err != nil {
		log.Fatalln(err)
	}

	conn.SetReadDeadline(time.Now().Add(time.Second * 10))

	// 初始化核心链接
	conn.Write([]byte("start"))

	// heartbeat
	go func() {
		for {
			conn.Write([]byte("0x"))
			time.Sleep(time.Second * 3)
		}
	}()

	for {
		buf := make([]byte, 10240)
		n, err := conn.Read(buf)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				conn.Write([]byte("0x"))
				continue
			}
		}

		// 检测 心跳
		if buf[0] == '0' && buf[1] == 'x' {
			if debug {
				log.Println("heartbeat")
			}
			continue
		}

		// 检测新链接
		if string(buf[:n]) == "new" {
			go func() {
				miniServer := NewMiniServer()
				miniServer.run()
			}()
		}
	}
}

// newConn 初始化新联接
func newConn(localAddr string) (net.Conn, error) {
	crt, err := tls.LoadX509KeyPair(certificateCrt, certificateKey)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, //这里是跳过证书验证，因为证书签发机构的CA证书是不被认证的
	}
	tlsConfig.Certificates = []tls.Certificate{crt}
	tlsConfig.Time = time.Now
	tlsConfig.Rand = rand.Reader
	localConn, err := tls.Dial("tcp", localAddr, tlsConfig)
	if err != nil {
		return nil, err
	}
	return localConn, nil
}

type MiniServer struct {
}

func NewMiniServer() *MiniServer {
	return &MiniServer{}
}

func (m *MiniServer) run() {
	conn, err := newConn(remoteAddr)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		conn.Close()
	}()

	conn.Write([]byte("new"))

	localConn, err := net.Dial("tcp", localAddr)
	if err != nil {
		log.Println(err)
		return
	}

	server := &server{
		conn:   conn,
		read:   make(chan []byte),
		write:  make(chan []byte),
		exit:   make(chan error),
		reConn: make(chan bool),
	}

	go server.Read()
	go server.Write()

	local := &localServer{
		conn:  localConn,
		read:  make(chan []byte),
		write: make(chan []byte),
		exit:  make(chan error),
	}

	go local.Read()
	go local.Write()

loop:
	for {
		select {
		case data, ex := <-server.read:
			if !ex {
				break loop
			}
			local.write <- data
		case data, ex := <-local.read:
			if !ex {
				break loop
			}
			server.write <- data
		case err := <-server.exit:
			fmt.Println(err)
			server.conn.Close()
			local.conn.Close()
			server.reConn <- true
		case err := <-server.exit:
			fmt.Println(err)
			local.conn.Close()
		}
	}
}
