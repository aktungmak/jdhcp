package jdhcp

import (
	"bytes"
	"encoding/binary"
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
	Cookie uint32
	Options
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

	if len(data) < 240 {
		// no options to process, just return
		msg.Options = make(Options)
		return msg, nil
	}

	var err error
	msg.Cookie = binary.BigEndian.Uint32(data[236:240])
	msg.Options, err = ParseOptions(data[240:])
	if err != nil {
		return nil, errors.Wrap(err, "parse options")
	}

	return msg, nil
}

func (m *Msg) MarshalBytes() []byte {
	var b bytes.Buffer
	b.Grow(364) // TODO check msg size

	b.WriteByte(m.Op)
	b.WriteByte(m.Htype)
	b.WriteByte(m.Hlen)
	b.WriteByte(m.Hops)

	binary.Write(&b, binary.BigEndian, m.Xid)
	binary.Write(&b, binary.BigEndian, m.Secs)
	binary.Write(&b, binary.BigEndian, m.Flags)

	b.Write(m.Ciaddr)
	b.Write(m.Yiaddr)
	b.Write(m.Siaddr)
	b.Write(m.Giaddr)
	b.Write(m.Chaddr)
	b.WriteString(m.Sname)
	b.WriteString(m.File)

	if len(m.Options) > 0 {
		binary.Write(&b, binary.BigEndian, m.Cookie)
		b.Write(m.Options.MarshalBytes())
	}

	return b.Bytes()
}