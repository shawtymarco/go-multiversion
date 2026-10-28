//go:build legacytransport

package v1_16_100

import (
	"context"
	"testing"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestProtocol419LoginFirstEncryptedSpawn(t *testing.T) {
	nativeBlocks := protocol419IntegrationBlockRegistry(t)
	nativeItems := []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}, {Name: "minecraft:shield", RuntimeID: 2}}
	target, err := NewWithRegistries(nativeBlocks, nativeItems)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := (minecraft.ListenConfig{AuthenticationDisabled: true, AcceptedProtocols: []minecraft.Protocol{target}}).Listen("raknet", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverResult := make(chan error, 1)
	go func() {
		raw, err := listener.Accept()
		if err != nil {
			serverResult <- err
			return
		}
		conn := raw.(*minecraft.Conn)
		defer conn.Close()
		err = conn.StartGameContext(ctx, minecraft.GameData{
			WorldName:                    "protocol-419-integration",
			BaseGameVersion:              Version,
			PlayerGameMode:               packet.GameTypeCreative,
			WorldGameMode:                packet.GameTypeCreative,
			PlayerPosition:               mgl32.Vec3{0, 64, 0},
			WorldSpawn:                   protocol.BlockPos{0, 64, 0},
			Items:                        nativeItems,
			GameRules:                    []protocol.GameRule{{Name: "doDaylightCycle", Value: false}},
			ServerAuthoritativeInventory: true,
			PlayerMovementSettings:       protocol.PlayerMovementSettings{ServerAuthoritativeBlockBreaking: true},
		})
		if err == nil {
			err = conn.WritePacket(&packet.Text{TextType: packet.TextTypeSystem, Message: "protocol-419-ready"})
		}
		serverResult <- err
	}()

	id := uuid.New()
	dialer := minecraft.Dialer{
		Protocol:     target,
		IdentityData: login.IdentityData{Identity: id.String(), DisplayName: "Protocol419E2E"},
	}
	client, err := dialer.DialContext(ctx, "raknet-v10", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := client.DoSpawnContext(ctx); err != nil {
		t.Fatal(err)
	}
	_ = client.SetReadDeadline(time.Now().Add(3 * time.Second))
	for {
		pk, err := client.ReadPacket()
		if err != nil {
			t.Fatal(err)
		}
		if text, ok := pk.(*packet.Text); ok && text.Message == "protocol-419-ready" {
			break
		}
	}
	select {
	case err := <-serverResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
