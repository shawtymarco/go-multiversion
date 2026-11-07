package dragonfly

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	multiversion "github.com/shawtymarco/go-multiversion"
)

func TestUpstreamChunkEncoderAcrossCatalogue(t *testing.T) {
	reg := world.DefaultBlockRegistry
	reg.Finalize()
	protocols, err := multiversion.ProtocolsWithBlockRegistry(reg)
	if err != nil {
		t.Fatal(err)
	}
	native, err := NativeProtocol(reg)
	if err != nil {
		t.Fatal(err)
	}
	protocols = append(protocols, native)
	stone := reg.BlockRuntimeID(block.Stone{})
	for _, p := range protocols {
		t.Run(p.Ver(), func(t *testing.T) {
			if p.ID() < 748 {
				t.Fatal("legacy transport advertised")
			}
			mapper := p.(blockMapper).MapBlockRuntimeID
			biomes := func(id uint32) (uint32, bool) { return id, true }
			if m, ok := p.(biomeMapper); ok {
				biomes = m.MapBiomeRuntimeID
			}
			for _, dim := range []world.Dimension{world.Overworld, world.Nether, world.End} {
				c := chunk.New(reg, dim.Range())
				y := int16(dim.Range().Min())
				c.SetBlock(1, y, 2, 0, stone)
				c.SetBlock(1, y, 2, 1, stone)
				c.SetBiome(1, y, 2, 195)
				dimID, _ := world.DimensionID(dim)
				sub := chunk.EncodeSubChunk(c, chunk.NetworkEncoding, 0)
				bio := chunk.EncodeBiomes(c, chunk.NetworkEncoding)
				raw := append(append(bytes.Clone(sub), bio...), 0)
				raw = append(raw, 10, 0, 42)
				pk := &packet.LevelChunk{Dimension: int32(dimID), SubChunkCount: 1, RawPayload: raw}
				before := bytes.Clone(raw)
				out, err := convertChunk(pk, mapper, biomes)
				if err != nil {
					t.Fatal(err)
				}
				buf := bytes.NewBuffer(out.(*packet.LevelChunk).RawPayload)
				decoded, err := chunk.NetworkDecodeBuffer(reg, buf, 1, dim.Range())
				if err != nil {
					t.Fatal(err)
				}
				want, _ := mapper(stone)
				for _, layer := range []uint8{0, 1} {
					if got := decoded.Block(1, y, 2, layer); got != want {
						t.Fatalf("block layer %d: %d != %d", layer, got, want)
					}
				}
				wantBiome, _ := biomes(195)
				if decoded.Biome(1, y, 2) != wantBiome {
					t.Fatal("biome not mapped")
				}
				if !bytes.Equal(buf.Bytes(), []byte{0, 10, 0, 42}) {
					t.Fatal("border/NBT tail changed")
				}
				if !bytes.Equal(raw, before) || c.Block(1, y, 2, 0) != stone {
					t.Fatal("source packet or world mutated")
				}
				// Request mode sends only biomes in LevelChunk, then block data in SubChunk.
				request := &packet.LevelChunk{Dimension: int32(dimID), SubChunkLimit: protocol.Option(int32(0)), RawPayload: append(bytes.Clone(bio), 0)}
				if _, err := convertChunk(request, mapper, biomes); err != nil {
					t.Fatal(err)
				}
				sp := &packet.SubChunk{SubChunkEntries: []protocol.SubChunkEntry{{Result: protocol.SubChunkResultSuccess, RawPayload: protocol.Option(sub)}}}
				mapped, err := convertChunk(sp, mapper, biomes)
				if err != nil {
					t.Fatal(err)
				}
				mappedSub, _ := mapped.(*packet.SubChunk).SubChunkEntries[0].RawPayload.Value()
				if !bytes.HasPrefix(out.(*packet.LevelChunk).RawPayload, mappedSub) {
					t.Fatal("LevelChunk and SubChunk disagree")
				}
			}
		})
	}
}

func TestCacheAndMalformedPayloadRejected(t *testing.T) {
	mapper := func(id uint32) (uint32, bool) { return id, true }
	for _, pk := range []packet.Packet{
		&packet.LevelChunk{CacheEnabled: true},
		&packet.LevelChunk{BlobHashes: []uint64{1}},
		&packet.LevelChunk{Dimension: 999},
		&packet.SubChunk{CacheEnabled: true},
		&packet.SubChunk{SubChunkEntries: []protocol.SubChunkEntry{{Result: protocol.SubChunkResultSuccess}}},
		&packet.SubChunk{SubChunkEntries: []protocol.SubChunkEntry{{BlobHash: protocol.Option(uint64(1))}}},
		&packet.ClientCacheMissResponse{},
	} {
		if _, err := convertChunk(pk, mapper, mapper); err == nil {
			t.Fatalf("accepted invalid %T", pk)
		}
	}
	if (&connection{}).ClientCacheEnabled() {
		t.Fatal("Dragonfly must use uncached encoder")
	}
	pk := &packet.SubChunk{SubChunkEntries: []protocol.SubChunkEntry{{Result: protocol.SubChunkResultSuccessAllAir}, {Result: protocol.SubChunkResultChunkNotFound}}}
	out, err := convertChunk(pk, mapper, mapper)
	if err != nil || !reflect.DeepEqual(out, pk) {
		t.Fatal("empty/error subchunk entries changed")
	}
}

func TestEnableValidationDoesNotChangeConfig(t *testing.T) {
	if err := Enable(nil, "127.0.0.1:0"); err == nil {
		t.Fatal("nil config accepted")
	}
	conf := server.Config{}
	if err := Enable(&conf, "bad address"); err == nil || len(conf.Listeners) != 0 {
		t.Fatal("invalid address changed config")
	}
	if err := Enable(&conf, "127.0.0.1:0"); err != nil || len(conf.Listeners) != 1 {
		t.Fatal(err)
	}
}

// This is an automated upstream-server smoke check, not a Mojang-client gate.
func TestUpstreamDragonflyJoin(t *testing.T) {
	conf := server.Config{AuthDisabled: true, ReadOnlyWorld: true, DisableResourceBuilding: true, MaxChunkRadius: 2, Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	var address string
	factory := (Config{}).Listener("127.0.0.1:0")
	conf.Listeners = []func(server.Config) (server.Listener, error){func(c server.Config) (server.Listener, error) {
		l, err := factory(c)
		if err == nil {
			address = l.(*listener).Addr().String()
		}
		return l, err
	}}
	srv := conf.New()
	if address == "" {
		t.Fatal("listener failed to initialise")
	}
	srv.Listen()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range srv.Accept() {
		}
	}()
	defer func() {
		_ = srv.Close()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("server accept loop did not stop")
		}
	}()

	for _, wire := range []minecraft.Protocol{minecraft.DefaultProtocol, multiversion.V1_26_50(), multiversion.V1_26_45(), multiversion.V1_26_44(), multiversion.V1_26_30(), multiversion.V1_26_20(), multiversion.V1_26_10(), multiversion.V1_26_0(), multiversion.V1_21_130(), multiversion.V1_21_110(), multiversion.V1_21_100(), multiversion.V1_21_50(), multiversion.V1_21_40()} {
		t.Run(wire.Ver(), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			dialer := minecraft.Dialer{Protocol: wire, IdentityData: login.IdentityData{Identity: uuid.NewString(), DisplayName: "UpstreamMV"}, ClientData: login.ClientData{GameVersion: wire.Ver()}}
			client, err := dialer.DialContext(ctx, "raknet", address)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			if err := client.DoSpawnContext(ctx); err != nil {
				t.Fatal(err)
			}
			if wire.ID() == protocol.CurrentProtocol && !client.GameData().UseBlockNetworkIDHashes {
				t.Fatal("native client was not given the hash bridge")
			}
			if wire.ID() != protocol.CurrentProtocol && client.GameData().UseBlockNetworkIDHashes {
				t.Fatal("historical client was told to use native hashes")
			}
			// 748/766 embed vanilla items in their StartGame wire layout; their
			// wire-only dialer does not expose that legacy field in native GameData.
			if wire.ID() >= 827 && len(client.GameData().Items) == 0 {
				t.Fatal("actual server item registry missing")
			}
			_ = client.SetReadDeadline(time.Now().Add(5 * time.Second))
			for {
				pk, err := client.ReadPacket()
				if err != nil {
					t.Fatal(err)
				}
				if level, ok := pk.(*packet.LevelChunk); ok {
					if level.CacheEnabled || len(level.BlobHashes) != 0 || len(level.RawPayload) == 0 {
						t.Fatal("invalid live chunk/cache boundary")
					}
					break
				}
			}
		})
	}
}
