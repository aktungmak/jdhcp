package jdhcp

import (
	"context"
	"log"
	"net"
	"os"
	"testing"
	"time"
)

var (
	testAddr = net.IPv4(127, 67, 67, 67)
	testPort = uint16(6767)
	testLogg = log.New(os.Stderr, "", log.LstdFlags)
)

func TestServerE2E(t *testing.T) {
	serv := NewServer(context.Background(), testLogg, testAddr, testPort)

	err := serv.Start()
	if err != nil {
		t.Fatalf("could not start server: %s", err)
	}

	// first let it sit for a while, to see what happens
	time.Sleep(3 * time.Second)

	err = serv.Stop()
	if err != nil {
		t.Fatalf("could not stop server: %s", err)
	}
}
