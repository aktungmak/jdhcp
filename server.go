package jdhcp

import (
	"context"
	"github.com/pkg/errors"
	"log"
	"net"
	"sync"
	"time"
)

// a MsgCallback is provided by the user of this library to
// define the actions that should be taken when a DHCP message
// is received. The callback is provided a parsed version of the
// incoming message, and should return a Msg containing the response
// to be sent. If there is no response needed, the returned Msg
// should be nil.
type MsgCallback func(Msg) *Msg

// a Server parses incoming DHCP messages and then
// calls the relevant callbacks with the received information
// based on the result of the callback, it will send a response
type Server struct {
	ctx     context.Context
	cancel  context.CancelFunc
	address net.IP
	port    int
	socket  *net.UDPConn
	// socket    net.PacketConn
	listening bool
	log       *log.Logger

	cbMutex sync.RWMutex
	msgCb   MsgCallback
}

// create and initialise a new Server
func NewServer(ctx context.Context, lg *log.Logger, address net.IP, port int) *Server {
	ctx, cancel := context.WithCancel(ctx)
	return &Server{
		ctx:     ctx,
		cancel:  cancel,
		address: address,
		port:    port,
		log:     lg,
	}
}

// begin listening for incoming DHCP messages
func (l *Server) Start() error {
	l.log.Print("starting dhcp server")
	if l.Listening() {
		return nil
	}

	var err error
	l.socket, err = net.ListenUDP("udp4",
		&net.UDPAddr{l.address, l.port, ""})
	// l.socket, err = net.ListenPacket("udp4",
	// 	fmt.Sprintf("%s:%d", l.address, l.port))
	if err != nil {
		return errors.Wrap(err, "open listening socket")
	}

	l.listening = true

	go l.loop()

	l.log.Print("started dhcp server")
	return nil
}

// stop the Server and close down all resources
func (l *Server) Stop() error {
	l.log.Print("stopping dhcp server")
	if !l.Listening() {
		return nil
	}

	l.cancel()
	err := l.socket.Close()
	if err != nil {
		return err
	}

	l.listening = false

	l.log.Print("stopped dhcp server")
	return nil
}

// check whether the Server is currently running
func (l *Server) Listening() bool {
	return l.listening
}

// register a callback with the Server
func (l *Server) RegisterCallback(cb MsgCallback) {
	l.cbMutex.Lock()
	l.msgCb = cb
	l.cbMutex.Unlock()
}

func (l *Server) loop() {
	buf := make([]byte, 4096)
	for {
		select {
		case <-l.ctx.Done():
			return
		default:
			// time out often so we go round the loop and check ctx
			l.socket.SetReadDeadline(time.Now().Add(time.Second))

			// try to read a packet
			n, addr, err := l.socket.ReadFromUDP(buf)
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue // just a timeout
				}
				panic("can't read")
			}

			go l.handleMsg(buf[:n], addr)
			if err != nil {
				continue
			}
		}
	}
}

// process an incoming DHCP message, dispatch it
// to the right callback and send a response (if needed)
func (l *Server) handleMsg(data []byte, from *net.UDPAddr) {
	req, err := ParseMsg(data)
	if err != nil {
		l.log.Printf("error handling message from %s: %s", from, err)
		return
	}

	var res *Msg
	l.cbMutex.RLock()
	if l.msgCb != nil {
		res = l.msgCb(*req)
	}
	l.cbMutex.RUnlock()

	if res == nil {
		return // no response, so we are done
	}

	payload := res.MarshalBytes()
	_, err = l.socket.WriteToUDP(payload, from)
	if err != nil {
		l.log.Printf("error writing response to %s: %s", from, err)
		return
	}

	l.log.Print("successfully handled message from %s", from)
}
