package sttypes

import (
	"github.com/hashicorp/go-version"
	libp2p_network "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	p2ptypes "github.com/zennittians/intelchain/p2p/types"
)

// Protocol is the interface of protocol to be registered to libp2p.
type Protocol interface {
	p2ptypes.LifeCycle

	Specifier() string
	Version() *version.Version
	ProtoID() ProtoID
	// ShardProtoID() ProtoID
	IsBeaconNode() bool
	Match(id protocol.ID) bool
	HandleStream(st libp2p_network.Stream)
}

// Request is the interface of a stream request used for common stream utils.
type Request interface {
	ReqID() uint64
	SetReqID(rid uint64)
	String() string
	IsSupportedByProto(ProtoSpec) bool
	Encode() ([]byte, error)
}

// Response is the interface of a stream response used for common stream utils
type Response interface {
	ReqID() uint64
	String() string
}
