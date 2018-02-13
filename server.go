package jdhcp

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
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
type MsgCallback func(Msg) (*Msg, error)

// a Server parses incoming DHCP messages and then
// calls the relevant callbacks with the received information
// based on the result of the callback, it will send a response
type Server struct {
	ctx       context.Context
	cancel    context.CancelFunc
	address   net.IP
	port      uint16
	socket    net.PacketConn
	listening bool

	cbMutex sync.RWMutex
	msgCb   MsgCallback
}

// create and initialise a new Server
func NewServer(ctx context.Context, address net.IP, port uint16) *Server {
	ctx, cancel := context.WithCancel(ctx)
	return &Server{
		ctx:     ctx,
		cancel:  cancel,
		address: address,
		port:    port,
	}
}

// begin listening for incoming DHCP messages
func (l *Server) Start() error {
	if l.listening {
		return nil
	}

	var err error
	l.socket, err = net.ListenPacket("udp4",
		fmt.Sprintf("%s:%d", l.address, l.port))
	if err != nil {
		return errors.Wrap(err, "open listening socket")
	}

	return nil
}

// stop the Server and close down all resources
func (l *Server) Stop() error {
	if !l.listening {
		return nil
	}
	l.cancel()
	return l.socket.Close()
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
	buf := make([]byte, 0, 4096)
	for {
		select {
		case <-l.ctx.Done():
			return
		default:
			// time out often so we go round the loop and check ctx
			l.socket.SetReadDeadline(time.Now().Add(time.Second))

			// try to read a packet
			n, addr, err := l.socket.ReadFrom(buf)
			if err != nil {
				if e, ok := err.(net.Error); ok && e.Timeout() {
					continue // just a timeout
				}
				panic("can't read")
			}

			err = l.handleMsg(buf[:n], addr)
			if err != nil {
				print("error handling message from ", addr, err)
				continue
			}

			print("successfully handled message from ", addr)
		}
	}
}

// process an incoming DHCP message, dispatch it
// to the right callback and send a response (if needed)
func (l *Server) handleMsg(data []byte, from net.Addr) error {
	req, err := ParseMsg(data)
	if err != nil {
		return errors.Wrapf(err, "parse message")
	}

	l.cbMutex.RLock()
	res, err := l.msgCb(*req)
	l.cbMutex.RUnlock()
	if err != nil {
		return errors.Wrap(err, "execute callback")
	}

	// check if there is a response to send
	if res != nil {
		payload := res.MarshalBytes()
		_, err = l.socket.WriteTo(payload, from)
		if err != nil {
			return errors.Wrap(err, "write response")
		}
		print("sent response to ", from)
	}

	return nil
}
