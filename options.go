package jdhcp

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

// Options stored the contents of the options field of a DHCP message
// as a mapping from OptionCode to the raw bytes of the message.
// This means that it only supports one occurrence of each option,
// although the spec is not clear whether that is a problem (other
// implementations also use a map structure)
//
// There are accessor methods defined which parse the raw bytes and
// return a typed interpretation of the value. these should be used
// unless the actual bytes are needed for some reason.

// it is not safe for concurrent use
// TODO consider adding a mutex.
type Options map[OptionCode][]byte

func ParseOptions(data []byte) (Options, error) {
	opts := make(Options)

	buf := bytes.NewBuffer(data)

	var err error
	var code byte
	for {
		code, err = buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return opts, err
		}

		switch OptionCode(code) {
		case OptionPad:
			// do nothing
		case OptionEnd:
			break
		case OptionRequestedIPAddress:
			buf.ReadByte() // ignore len
			opts[OptionRequestedIPAddress] = buf.Next(4)
		case OptionDHCPMessageType:
			buf.ReadByte() // ignore len
			opts[OptionDHCPMessageType] = buf.Next(1)
		default:
			print("ignoring unknown optioncode ", code)
			l, _ := buf.ReadByte()
			buf.Next(int(l))
		}
	}
	return opts, nil
}

// convert to a []byte suitable for sending over the wire
func (o Options) MarshalBytes() []byte {
	var b bytes.Buffer

	for k, v := range o {
		b.WriteByte(byte(k))
		b.WriteByte(byte(len(v)))
		b.Write(v)
	}
	b.WriteByte(byte(OptionEnd))

	return b.Bytes()
}

// insert an option to the set
func (o Options) Insert(oc OptionCode, v interface{}) error {
	var b bytes.Buffer
	err := binary.Write(&b, binary.BigEndian, v)
	if err != nil {
		return err
	}

	o[oc] = b.Bytes()
	return nil
}

func (o Options) RequestedIPAddress() (net.IP, error) {
	a, ok := o[OptionRequestedIPAddress]
	if !ok {
		return nil, ErrOptionNotPresent
	}
	return net.IP(a), nil
}

func (o Options) DHCPMessageType() (MessageType, error) {
	t, ok := o[OptionDHCPMessageType]
	if !ok {
		return 0, ErrOptionNotPresent
	}
	if len(t) < 1 {
		return 0, ErrShortRead
	}

	return MessageType(t[0]), nil
}
