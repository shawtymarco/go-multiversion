// Package packetio contains bounded helpers shared by historical wire adapters.
package packetio

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// MaxSubChunkRequestOffsets permits the legitimate burst emitted when a client
// replaces a backend at high render distance while retaining a finite packet
// allocation limit. Each offset occupies three bytes on the wire.
const MaxSubChunkRequestOffsets uint32 = 4096

type sliceReader interface {
	SliceLength(value uint32, max uint32)
}

// SubChunkOffsets reads or writes exactly count sub-chunk offsets using the
// larger packet-specific limit required by fast backend transfers.
func SubChunkOffsets(io protocol.IO, count uint32, offsets *[]protocol.SubChunkOffset) {
	if reader, ok := io.(sliceReader); ok {
		reader.SliceLength(count, MaxSubChunkRequestOffsets)
		*offsets = make([]protocol.SubChunkOffset, count)
	}
	for index := uint32(0); index < count; index++ {
		(*offsets)[index].Marshal(io)
	}
}
