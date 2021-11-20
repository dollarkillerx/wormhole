package main

import (
	"crypto/rand"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	localPort      int    // 本地地址
	remotePort     int    // 远程地址
	certificateCrt string // 证书crt
	certificateKey string // 证书key
	debug          bool   // debug mode
)

func init() {
	fmt.Println("Wormhole Server")
	flag.IntVar(&localPort, "l", 8087, "local port")
	flag.IntVar(&remotePort, "r", 8087, "remote port")
	flag.StringVar(&certificateCrt, "c", "proxy.crt", "proxy.crt")
	flag.StringVar(&certificateKey, "k", "proxy.key", "proxy.key")
	flag.BoolVar(&debug, "d", false, "debug")
	flag.Parse()
}

type localServer struct {
	conn  net.Conn
	read  chan []byte
	write chan []byte

	exit   chan error
	reConn chan bool
}

func (l *localServer) Read() {
	l.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	for {
		data := make([]byte, 10240)
		n, err := l.conn.Read(data)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "timeout") {
				l.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				l.conn.Write([]byte("0x"))
				if debug {
					log.Println("timeout")
				}
				break
			}
			log.Println(err)
			l.exit <- err
			break
		}

		if data[0] == '0' && data[1] == 'x' {
			if debug {
				log.Println("heartbeat")
			}
			continue
		}
		l.read <- data[:n]
	}
}

func (l *localServer) Write() {
	for {
		select {
		case data, ex := <-l.write:
			if !ex {
				return
			}

			_, err := l.conn.Write(data)
			if err != nil {
				l.exit <- err
				if debug {
					log.Println(err)
				}
				return
			}
		}
	}
}

type remoteServer struct {
	conn  net.Conn
	read  chan []byte
	write chan []byte

	exit chan error
}

func (r *remoteServer) Read() {
	r.conn.SetReadDeadline(time.Now().Add(time.Second * 20))
	for {
		data := make([]byte, 10240)
		n, err := r.conn.Read(data)
		if err != nil && err != io.EOF {
			r.exit <- err
			log.Println(err)
			break
		}
		r.read <- data[:n]
	}
}

func (r *remoteServer) Write() {
	for {
		select {
		case data, ex := <-r.write:
			if !ex {
				break
			}

			_, err := r.conn.Write(data)
			if err != nil {
				r.exit <- err
				log.Println(err)
				break
			}
		}
	}
}

var taskManagerConn net.Conn
var mu sync.Mutex

func main() {
	if debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	crt, err := tls.LoadX509KeyPair(certificateCrt, certificateKey)
	if err != nil {
		log.Fatalln(err)
	}
	tlsConfig := &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{crt}
	tlsConfig.Time = time.Now
	tlsConfig.Rand = rand.Reader
	localConn, err := tls.Listen("tcp", fmt.Sprintf(":%d", localPort), tlsConfig)
	if err != nil {
		log.Fatalln(err)
	}

	remoteConn, err := net.Listen("tcp", fmt.Sprintf(":%d", remotePort))
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("LocalConn: 0.0.0.0:%d RemoteConn: 0.0.0.0:%d \n", localPort, remotePort)
	var connChan = make(chan net.Conn, 100)

	// remote server
	go func() {
		for {
			accept, err := remoteConn.Accept()
			if err != nil {
				if debug {
					log.Println(err)
				}
				continue
			}

			fmt.Println("Remote conn: ", accept.RemoteAddr())
			if err := newConnection(); err != nil {
				if debug {
					log.Println(err)
				}
				continue
			}

			select {
			case <-time.After(time.Second * 10):
				if debug {
					log.Println("Client Timeout")
				}
				continue
			case clientConn := <-connChan:
				// 得到 client conn
				cs := &localServer{
					conn:   clientConn,
					read:   make(chan []byte),
					write:  make(chan []byte),
					exit:   make(chan error),
					reConn: make(chan bool),
				}

				rs := &remoteServer{
					conn:  accept,
					read:  make(chan []byte),
					write: make(chan []byte),
					exit:  make(chan error),
				}

				go cs.Write()
				go cs.Read()
				go rs.Write()
				go rs.Read()

				go handle(cs, rs)
			}

		}
	}()

	// 管理 client 链接
	for {
		accept, err := localConn.Accept()
		if err != nil {
			if debug {
				log.Println(err)
			}
			continue
		}

		fmt.Println("Client conn: ", accept.RemoteAddr())

		buf := make([]byte, 512)
		read, err := accept.Read(buf)
		if err != nil {
			if debug {
				log.Println(err)
			}
			continue
		}
		if "start" == string(buf[:read]) {
			mu.Lock()
			taskManagerConn = accept
			mu.Unlock()
			continue
		}

		connChan <- accept
	}
}

func newConnection() error {
	mu.Lock()
	defer mu.Unlock()
	if taskManagerConn == nil {
		return errors.New("无客户端")
	}

	_, err := taskManagerConn.Write([]byte("new"))
	if err != nil {
		if debug {
			log.Println(err)
		}
		return err
	}

	return nil
}

func handle(local *localServer, remote *remoteServer) {
	for {
		select {
		case ur := <-remote.read:
			local.write <- ur
		case cr := <-local.read:
			remote.write <- cr
		case err := <-remote.exit:
			if debug {
				if err != nil {
					log.Println(err)
				}
			}
			break
		case err := <-local.exit:
			if debug {
				if err != nil {
					log.Println(err)
				}
			}
			local.conn.Close()
			remote.conn.Close()
			break
		}
	}
}
