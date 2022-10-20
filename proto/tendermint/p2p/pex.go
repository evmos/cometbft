package p2p

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
)

// Wrap implements the p2p Wrapper interface and wraps a PEX message.
func (m *Message) Wrap() (proto.Message, error) {
	switch msg := pb.(type) {
	case *PexRequest:
		m.Sum = &Message_PexRequest{PexRequest: msg}
	case *PexAddrs:
		m.Sum = &Message_PexAddrs{PexAddrs: msg}
	default:
		return fmt.Errorf("unknown pex message: %T", msg)
	}
	return nil
}

// Unwrap implements the p2p Wrapper interface and unwraps a wrapped PEX
// message.
func (m *Message) Unwrap() (proto.Message, error) {
	switch msg := m.Sum.(type) {
	case *Message_PexRequest:
		return msg.PexRequest, nil
	case *Message_PexAddrs:
		return msg.PexAddrs, nil
	default:
		return nil, fmt.Errorf("unknown pex message: %T", msg)
	}
}
