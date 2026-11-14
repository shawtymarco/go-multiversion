package packetio

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// TrustedSkinBool encodes the native three-state trusted skin flag as the plain bool that older protocols
// carry. An unset flag is written as false, which is how clients treat it. Writing never changes the packet:
// the flag is only stored when the wire value differs from it or the IO reports that it is reading.
func TrustedSkinBool(io protocol.IO, x *protocol.Optional[bool]) {
	old, _ := x.Value()
	trusted := old
	io.Bool(&trusted)
	if trusted != old || reading(io) {
		*x = protocol.Option(trusted)
	}
}

func reading(io protocol.IO) bool {
	switch r := io.(type) {
	case *protocol.Reader:
		return true
	case interface{ LegacyWireReading() bool }:
		return r.LegacyWireReading()
	}
	return false
}
