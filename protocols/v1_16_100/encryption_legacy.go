//go:build legacytransport

package v1_16_100

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// Encryption selects the CFB8 cipher used before Minecraft 1.16.220.
func (Protocol) Encryption(key [32]byte) packet.Encryption { return packet.NewCFB8Encryption(key) }
