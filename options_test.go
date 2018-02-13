package jdhcp

import (
	"bytes"
	"net"
	"testing"
)

var optionsParseCases = []struct {
	asBytes []byte
	asMap   Options
}{
	// 0
	{[]byte{0x35, 0x01, 0x01, 0xFF},
		Options{OptionDHCPMessageType: {0x01}}},
	// 1
	{[]byte{0x32, 0x04, 0xC0, 0xA8, 0x01, 0x01, 0xFF},
		Options{OptionRequestedIPAddress: {0xC0, 0xA8, 0x01, 0x01}}},
}

func TestParseOptions(t *testing.T) {
	for i, tc := range optionsParseCases {
		got, err := ParseOptions(tc.asBytes)
		if err != nil {
			t.Errorf("case %d returned error: %s", i, err)
			continue
		}

		if len(tc.asMap) != len(got) {
			t.Errorf("case %d returned incorrect length, expected %d got %d",
				i, len(tc.asMap), len(got))
			continue
		}

		for ek, ev := range tc.asMap {
			gv, ok := got[ek]
			if !ok {
				t.Errorf("case %d key %s missing", i, ek)
				continue
			}
			if bytes.Compare(ev, gv) != 0 {
				t.Errorf("case %d key %v expected %v got %v", i, ek, ev, gv)
			}
		}
	}
}

func TestOptionsMarshalBytes(t *testing.T) {
	for i, tc := range optionsParseCases {
		got := tc.asMap.MarshalBytes()

		if bytes.Compare(tc.asBytes, got) != 0 {
			t.Errorf("case %d expected %v got %v", i, tc.asBytes, got)
		}
	}
}

func TestRequestedIPAddress(t *testing.T) {
	o := make(Options)
	a1 := net.IPv4(192, 168, 1, 1)
	o.Insert(OptionRequestedIPAddress, []byte(a1))

	a2, err := o.RequestedIPAddress()
	if err != nil {
		t.Fatalf("o.RequestedIPAddress() returned error: %s", err)
	}

	if !a1.Equal(a2) {
		t.Fatalf("returned incorrect IP, expected %s got %s", a1, a2)
	}
}

func TestDHCPMessageType(t *testing.T) {
	o := make(Options)
	t1 := MessageType(1)
	o.Insert(OptionDHCPMessageType, []byte{byte(t1)})

	t2, err := o.DHCPMessageType()
	if err != nil {
		t.Fatalf("o.DHCPMessageType() returned error: %s", err)
	}

	if t1 != t2 {
		t.Fatalf("returned incorrect type, expected %s got %s", t1, t2)
	}
}
