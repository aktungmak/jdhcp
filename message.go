package jdhcp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"net"
)

type Msg struct {
	Op     byte
	Htype  byte
	Hlen   byte
	Hops   byte
	Xid    uint32
	Secs   uint16
	Flags  uint16
	Ciaddr net.IP
	Yiaddr net.IP
	Siaddr net.IP
	Giaddr net.IP
	Chaddr net.HardwareAddr
	Sname  string
	File   string
	Options
}

// initialise a blank Msg, should be used to ensure correct init
func NewMsg() *Msg {
	return &Msg{
		Ciaddr:  net.IPv4(0, 0, 0, 0),
		Yiaddr:  net.IPv4(0, 0, 0, 0),
		Siaddr:  net.IPv4(0, 0, 0, 0),
		Giaddr:  net.IPv4(0, 0, 0, 0),
		Chaddr:  net.HardwareAddr([]byte{0, 0, 0, 0, 0, 0}),
		Options: Options{},
	}
}

func ParseMsg(data []byte) (*Msg, error) {
	if len(data) < 236 {
		return nil, ErrShortRead
	}

	msg := &Msg{
		Op:     data[0],
		Htype:  data[1],
		Hlen:   data[2],
		Hops:   data[3],
		Xid:    binary.BigEndian.Uint32(data[4:8]),
		Secs:   binary.BigEndian.Uint16(data[8:10]),
		Flags:  binary.BigEndian.Uint16(data[10:12]),
		Ciaddr: net.IP(data[12:16]),
		Yiaddr: net.IP(data[16:20]),
		Siaddr: net.IP(data[20:24]),
		Giaddr: net.IP(data[24:28]),
		Chaddr: net.HardwareAddr(data[28:34]), // 6-byte MAC only
		Sname:  string(bytes.TrimRight(data[44:108], "\000")),
		File:   string(bytes.TrimRight(data[108:236], "\000")),
	}

	if msg.Hlen != 6 {
		return nil, errors.Errorf("unsupported hlen of %d", msg.Hlen)
	}

	cookie := binary.BigEndian.Uint32(data[236:240])
	if Cookie != cookie {
		fmt.Printf("data: %v\n", data)
		return nil, errors.Errorf("incorrect cookie, expected %d got %d", Cookie, cookie)
	}

	var err error
	msg.Options, err = ParseOptions(data[240:])
	if err != nil {
		return nil, errors.Wrap(err, "parse options")
	}

	return msg, nil
}

// convert a Msg structure to the network representation
// TODO optimise the method of padding
func (m *Msg) MarshalBytes() []byte {
	var b bytes.Buffer
	b.Grow(272) // min size

	b.WriteByte(m.Op)
	b.WriteByte(m.Htype)
	b.WriteByte(m.Hlen)
	b.WriteByte(m.Hops)

	binary.Write(&b, binary.BigEndian, m.Xid)
	binary.Write(&b, binary.BigEndian, m.Secs)
	binary.Write(&b, binary.BigEndian, m.Flags)

	b.Write(m.Ciaddr.To4())
	b.Write(m.Yiaddr.To4())
	b.Write(m.Siaddr.To4())
	b.Write(m.Giaddr.To4())

	// pad to 16 bytes
	chaddr := make([]byte, 16)
	copy(chaddr, m.Chaddr)
	b.Write(chaddr)

	// pad to 64 bytes
	sname := make([]byte, 64)
	copy(sname, m.Sname)
	b.Write(sname)

	// pad to 128 bytes
	file := make([]byte, 128)
	copy(file, m.File)
	b.Write(file)

	binary.Write(&b, binary.BigEndian, Cookie)
	b.Write(m.Options.MarshalBytes())

	// pad out msg to at least 272
	if b.Len() < 272 {
		padding := make([]byte, 272)
		copy(padding, b.Bytes())
		return padding
	}

	return b.Bytes()
}
