package jdhcp

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"sort"
	"time"
)

// Options stores the contents of the options field of a DHCP message
// as a mapping from OptionCode to the raw bytes of the message.
// This means that it allows only one occurrence of each option,
// which is specified in chapter 4.1 of RFC2131.
//
// There are access methods defined which parse the raw bytes and
// return a typed interpretation of the value. these should be used
// unless the actual bytes are needed for some reason.
//
// it is not safe for concurrent access by multiple goroutines
type Options map[OptionCode][]byte

func ParseOptions(data []byte) (Options, error) {
	opts := make(Options)
	buf := bytes.NewBuffer(data)

	for {
		// read code
		code, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return opts, err
		}

		// handle options without length
		if OptionCode(code) == OptionEnd {
			break
		}
		if OptionCode(code) == OptionPad {
			continue
		}

		// read length
		l, err := buf.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return opts, err
		}

		opts[OptionCode(code)] = buf.Next(int(l))
	}
	return opts, nil
}

// convert to a []byte suitable for sending over the wire
// sort options before sending so that the result is deterministic
func (o Options) MarshalBytes() []byte {
	// first get all keys and sort them
	ks := make([]OptionCode, 0, len(o))
	for k, _ := range o {
		ks = append(ks, k)
	}
	sort.Slice(ks, func(i, j int) bool { return ks[i] < ks[j] })

	// write out the TLVs as bytes
	var b bytes.Buffer
	for _, k := range ks {
		b.WriteByte(byte(k))
		b.WriteByte(byte(len(o[k])))
		b.Write(o[k])
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

// option 1
func (o Options) SubnetMask() (net.IPMask, error) {
	a, ok := o[OptionSubnetMask]
	if !ok {
		return nil, ErrOptionNotPresent
	}
	return net.IPMask(a), nil
}

// option 50
func (o Options) RequestedIPAddress() (net.IP, error) {
	a, ok := o[OptionRequestedIPAddress]
	if !ok {
		return nil, ErrOptionNotPresent
	}
	return net.IP(a), nil
}

// option 53
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

// option 55
func (o Options) ParameterRequestList() ([]OptionCode, error) {
	pl, ok := o[OptionParameterRequestList]
	if !ok {
		return nil, ErrOptionNotPresent
	}

	ret := make([]OptionCode, 0, len(pl))
	for _, b := range pl {
		ret = append(ret, OptionCode(b))
	}

	return ret, nil
}

// option 58
func (o Options) RenewalTime() (time.Duration, error) {
	d, ok := o[OptionRenewalTime]
	if !ok {
		return 0, ErrOptionNotPresent
	}

	return time.Duration(binary.BigEndian.Uint64(d)), nil
}

// option 59
func (o Options) RebindingTime() (time.Duration, error) {
	d, ok := o[OptionRebindingTime]
	if !ok {
		return 0, ErrOptionNotPresent
	}

	return time.Duration(binary.BigEndian.Uint64(d)), nil
}

// option 61
func (o Options) ClientID() (kind byte, id []byte, err error) {
	ci, ok := o[OptionClientID]
	if !ok {
		err = ErrOptionNotPresent
		return
	}
	if len(ci) < 2 {
		err = ErrShortRead
		return
	}
	kind = ci[0]
	id = ci[1:]
	return
}
