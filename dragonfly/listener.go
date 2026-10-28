// Package dragonfly connects go-multiversion to unmodified upstream Dragonfly.
// It translates encoded chunks outside the world/session implementation and
// disables blob caching so that native hashes cannot escape into target chunks.
package dragonfly

import (
	"fmt"
	"net"
	"slices"

	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	multiversion "github.com/shawtymarco/go-multiversion"
	"github.com/shawtymarco/go-multiversion/internal/chunkconv"
)

// Enable replaces the configured listeners with one multiversion RakNet listener.
// Call after UserConfig.Config and before Config.New. address is normally
// UserConfig.Network.Address. Existing authentication, resource packs, compression,
// player limits and admission policy are retained. Use Config.Listener to compose
// multiple listeners explicitly.
func Enable(conf *server.Config, address string) error {
	if conf == nil {
		return fmt.Errorf("nil Dragonfly config")
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return fmt.Errorf("listen address: %w", err)
	}
	conf.Listeners = []func(server.Config) (server.Listener, error){(Config{}).Listener(address)}
	return nil
}

// Config selects the supported 1.21.40+ families. Zero selects all modern
// adapters. Legacy transport builds do not expand this integration's scope.
type Config struct {
	MinimumProtocol int32
}

// Listener returns a factory for server.Config.Listeners. Dragonfly calls it
// after finalising the actual block registry and building server resource packs.
func (c Config) Listener(address string) func(server.Config) (server.Listener, error) {
	return func(conf server.Config) (server.Listener, error) {
		if err := (multiversion.Config{MinimumProtocol: c.MinimumProtocol}).Validate(); err != nil {
			return nil, err
		}
		if conf.Blocks == nil {
			return nil, fmt.Errorf("Dragonfly block registry has not been initialised")
		}
		if c.MinimumProtocol != 0 && c.MinimumProtocol < 748 {
			return nil, fmt.Errorf("upstream Dragonfly integration requires minimum protocol 748 or newer")
		}
		// Item mapping is initialised from the real StartGame/ItemRegistry payload,
		// including custom items, rather than a second copy of the vanilla registry.
		protocols, err := multiversion.ProtocolsWithBlockRegistry(conf.Blocks)
		if err != nil {
			return nil, err
		}
		accepted := make([]minecraft.Protocol, 0, len(protocols)+1)
		for _, p := range protocols {
			if p.ID() >= 748 && p.ID() >= c.MinimumProtocol {
				accepted = append(accepted, p)
			}
		}
		native, err := NativeProtocol(conf.Blocks)
		if err != nil {
			return nil, err
		}
		accepted = append(accepted, native)
		cfg := minecraft.ListenConfig{
			MaximumPlayers: conf.MaxPlayers, StatusProvider: conf.StatusProvider,
			AuthenticationDisabled: conf.AuthDisabled, ResourcePacks: conf.Resources,
			TexturePacksRequired: conf.ResourcesRequired, Compression: conf.Compression,
			AcceptedProtocols: accepted, ErrorLog: conf.Log,
		}
		if conf.Allower != nil {
			cfg.Allow = conf.Allower.Allow
		}
		l, err := cfg.Listen("raknet", address)
		if err != nil {
			return nil, err
		}
		return &listener{Listener: l}, nil
	}
}

type listener struct{ *minecraft.Listener }

func (l *listener) Accept() (session.Conn, error) {
	raw, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	conn := raw.(*minecraft.Conn)
	m, ok := conn.Proto().(blockMapper)
	if !ok {
		_ = conn.Close()
		return nil, fmt.Errorf("protocol %d has no block mapper", conn.Proto().ID())
	}
	biomes := chunkconv.Mapper(func(id uint32) (uint32, bool) { return id, true })
	if mapper, ok := conn.Proto().(biomeMapper); ok {
		biomes = mapper.MapBiomeRuntimeID
	}
	return &connection{Conn: conn, blocks: m.MapBlockRuntimeID, biomes: biomes}, nil
}

func (l *listener) Disconnect(conn session.Conn, reason string) error {
	wrapped, ok := conn.(*connection)
	if !ok {
		return fmt.Errorf("connection does not belong to the multiversion listener")
	}
	return l.Listener.Disconnect(wrapped.Conn, reason)
}

type blockMapper interface{ MapBlockRuntimeID(uint32) (uint32, bool) }
type biomeMapper interface{ MapBiomeRuntimeID(uint32) (uint32, bool) }

type connection struct {
	*minecraft.Conn
	blocks, biomes chunkconv.Mapper
}

// ClientCacheEnabled controls Dragonfly's encoder, not just packet metadata.
// Dragonfly therefore never creates cache hashes or cache-miss transactions.
func (*connection) ClientCacheEnabled() bool { return false }

func (c *connection) WritePacket(pk packet.Packet) error {
	converted, err := convertChunk(pk, c.blocks, c.biomes)
	if err != nil {
		_ = c.Close()
		return err
	}
	return c.Conn.WritePacket(converted)
}

func convertChunk(pk packet.Packet, blocks, biomes chunkconv.Mapper) (packet.Packet, error) {
	switch source := pk.(type) {
	case *packet.LevelChunk:
		if source.CacheEnabled || len(source.BlobHashes) != 0 {
			return nil, fmt.Errorf("unexpected cached LevelChunk")
		}
		dim, ok := world.DimensionByID(int(source.Dimension))
		if !ok {
			return nil, fmt.Errorf("unknown chunk dimension %d", source.Dimension)
		}
		r := dim.Range()
		sections := (r[1] - r[0] + 1) / 16
		cloned := *source
		var err error
		cloned.RawPayload, err = chunkconv.LevelChunk(source.RawPayload, source.SubChunkCount, sections, blocks, biomes)
		return &cloned, err
	case *packet.SubChunk:
		if source.CacheEnabled {
			return nil, fmt.Errorf("unexpected cached SubChunk")
		}
		cloned := *source
		cloned.SubChunkEntries = slices.Clone(source.SubChunkEntries)
		for i, entry := range cloned.SubChunkEntries {
			if _, exists := entry.BlobHash.Value(); exists {
				return nil, fmt.Errorf("unexpected sub-chunk blob hash")
			}
			if entry.Result != protocol.SubChunkResultSuccess {
				continue
			}
			payload, exists := entry.RawPayload.Value()
			if !exists {
				return nil, fmt.Errorf("successful sub-chunk has no payload")
			}
			mapped, err := chunkconv.SubChunk(payload, blocks)
			if err != nil {
				return nil, err
			}
			cloned.SubChunkEntries[i].RawPayload = protocol.Option(mapped)
		}
		return &cloned, nil
	case *packet.ClientCacheMissResponse:
		return nil, fmt.Errorf("unexpected cache-miss response while chunk cache is disabled")
	}
	return pk, nil
}
