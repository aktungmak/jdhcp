package jdhcp

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

var (
	testAddr = net.IPv4(127, 67, 67, 67)
	testPort = 6767
	testLogg = log.New(os.Stderr, "", log.LstdFlags)
)

func TestServerE2E(t *testing.T) {
	serv := NewServer(context.Background(), testLogg, testAddr, testPort)

	err := serv.Start()
	if err != nil {
		t.Fatalf("could not start server: %s", err)
	}

	msg := NewMsg()
	msg.Op = 1
	msg.Htype = 1
	msg.Hlen = 6
	msg.Xid = 0x12345678
	msg.Chaddr = net.HardwareAddr([]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55})

	// use a channel to receive result from the callback
	result := make(chan error)
	defer close(result)
	serv.RegisterCallback(func(got Msg) *Msg {
		diff := cmp.Diff(*msg, got)
		if diff != "" {
			result <- errors.Errorf("sent message does not match received: %s", diff)
			return nil
		}
		result <- nil

		return nil
	})

	conn, err := net.DialUDP("udp4", nil,
		&net.UDPAddr{testAddr, testPort, ""})
	if err != nil {
		t.Fatalf("can't dial test host: %s", err)
	}
	defer conn.Close()

	_, err = conn.Write(msg.MarshalBytes())
	if err != nil {
		t.Fatalf("cannot write message to socket: %s", err)
	}

	select {
	case err = <-result:
		if err != nil {
			t.Fatalf("callback returned an error: %s", err)
		}
	case <-time.After(time.Second):
		t.Error("timed out waiting for callback to complete")
	}

	err = serv.Stop()
	if err != nil {
		t.Fatalf("could not stop server: %s", err)
	}
}
