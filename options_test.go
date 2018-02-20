package jdhcp

import (
	"bytes"
	"net"
	"testing"
	"time"
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
	// 2
	{[]byte{0x37, 0x04, 0x01, 0x03, 0x06, 0x2a, 0xFF},
		Options{OptionParameterRequestList: {0x01, 0x03, 0x06, 0x2a}}},
	// 3
	{[]byte{0x3d, 0x07, 0x01, 0x00, 0x0b, 0x82, 0x01, 0xfc, 0x42, 0xFF},
		Options{OptionClientID: {0x01, 0x00, 0x0b, 0x82, 0x01, 0xfc, 0x42}}},
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

func TestSubnetMask(t *testing.T) {
	o := make(Options)
	m1 := net.IPv4Mask(255, 255, 252, 0)
	o.Insert(OptionSubnetMask, []byte(m1))

	m2, err := o.SubnetMask()
	if err != nil {
		t.Fatalf("o.SubnetMask() returned error: %s", err)
	}

	if m1.String() != m2.String() {
		t.Fatalf("returned incorrect mask, expected %s got %s", m1, m2)
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

func TestParameterRequestList(t *testing.T) {
	o := make(Options)
	l1 := []OptionCode{0x01, 0x02, 0x03}
	o.Insert(OptionParameterRequestList, l1)

	l2, err := o.ParameterRequestList()
	if err != nil {
		t.Fatalf("o.ParameterRequestList() returned error: %s", err)
	}

	if len(l1) != len(l2) {
		t.Fatalf("incorrect length, expected %v got %v", l1, l2)
	}
	for i, v := range l1 {
		if v != l2[i] {
			t.Fatalf("parameter list is different, expected %v got %v", l1, l2)
		}
	}
}

func TestRenewalTime(t *testing.T) {
	o := make(Options)
	d1 := time.Hour
	o.Insert(OptionRenewalTime, d1)

	d2, err := o.RenewalTime()
	if err != nil {
		t.Fatalf("o.RenewalTime() returned error: %s", err)
	}

	if d1 != d2 {
		t.Fatalf("duration is different, expected %v got %v", d1, d2)
	}
}

func TestRebindingTime(t *testing.T) {
	o := make(Options)
	d1 := time.Hour
	o.Insert(OptionRebindingTime, d1)

	d2, err := o.RebindingTime()
	if err != nil {
		t.Fatalf("o.RebindingTime() returned error: %s", err)
	}

	if d1 != d2 {
		t.Fatalf("duration is different, expected %v got %v", d1, d2)
	}
}

func TestClientID(t *testing.T) {
	o := make(Options)
	t1 := byte(0x01)
	i1 := []byte{t1, 0x0f, 0xce, 0x0f, 0xce, 0x0f, 0xce}
	o.Insert(OptionClientID, i1)

	t2, i2, err := o.ClientID()
	if err != nil {
		t.Fatalf("o.ClientID() returned error: %s", err)
	}

	if t1 != t2 {
		t.Errorf("client ID type is wrong, expected %d got %d", t1, t2)
	}

	if len(i1)-1 != len(i2) {
		t.Fatalf("incorrect ID length, expected %v got %v", i1, i2)
	}
	for i, v := range i1[1:] {
		if v != i2[i] {
			t.Fatalf("ID is different, expected %v got %v", i1, i2)
		}
	}
}
