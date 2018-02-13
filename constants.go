package jdhcp

import (
	"errors"
)

var (
	ErrOptionNotPresent = errors.New("option not present")
	ErrShortRead        = errors.New("short read")

	Cookie uint32 = 0x63825363
)

type OptionCode byte

const (
	OptionPad OptionCode = 0
	OptionEnd OptionCode = 255

	OptionRequestedIPAddress   OptionCode = 50
	OptionDHCPMessageType      OptionCode = 53
	OptionParameterRequestList OptionCode = 55
	OptionClientID             OptionCode = 61
)

type MessageType byte

const (
	Discover MessageType = 1
	Offer    MessageType = 2
	Request  MessageType = 3
	Decline  MessageType = 4
	ACK      MessageType = 5
	NAK      MessageType = 6
	Release  MessageType = 7
	Inform   MessageType = 8
)
