package proxy

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/shawtymarco/go-multiversion/data/v2193"
	"github.com/shawtymarco/go-multiversion/mapping"
)

var nativeStates = sync.OnceValues(v2193.BlockStates)

// NetworkRegistry separates sparse network hashes from dense canonical IDs.
// Existing server-facing mapping.BlockRegistry implementations remain unchanged.
type NetworkRegistry struct {
	states    []mapping.BlockState
	air       uint32
	toOrdinal map[uint32]uint32
	toNetwork []uint32
}

func NewNetworkRegistry(hashed bool, custom []protocol.BlockEntry) (*NetworkRegistry, error) {
	states, err := nativeStates()
	if err != nil {
		return nil, err
	}
	states = append([]mapping.BlockState(nil), states...)
	for _, entry := range custom {
		known := false
		for _, s := range states {
			if s.Name == entry.Name {
				known = true
				break
			}
		}
		if known {
			continue
		}
		additional, err := customStates(entry)
		if err != nil {
			return nil, err
		}
		states = append(states, additional...)
	}
	hashName := func(s string) uint64 { h := fnv.New64(); _, _ = h.Write([]byte(s)); return h.Sum64() }
	sort.SliceStable(states, func(i, j int) bool { return hashName(states[i].Name) < hashName(states[j].Name) })
	r := &NetworkRegistry{states: states, toOrdinal: make(map[uint32]uint32, len(states)), toNetwork: make([]uint32, len(states))}
	for i, state := range states {
		id := uint32(i)
		if hashed {
			id, err = NetworkBlockHash(state.Name, state.Properties)
			if err != nil {
				return nil, err
			}
		}
		if old, exists := r.toOrdinal[id]; exists {
			return nil, fmt.Errorf("ambiguous block network ID %d at states %d and %d", id, old, i)
		}
		r.toOrdinal[id], r.toNetwork[i] = uint32(i), id
		if state.Name == "minecraft:air" {
			r.air = uint32(i)
		}
	}
	return r, nil
}

// Custom block property definitions enumerate a finite Cartesian product. Their
// declaration order is retained, just like the native registry loader.
func customStates(entry protocol.BlockEntry) ([]mapping.BlockState, error) {
	states := []mapping.BlockState{{Name: entry.Name, Properties: map[string]any{}}}
	raw, exists := entry.Properties["properties"]
	if !exists {
		return states, nil
	}
	var properties []map[string]any
	switch raw := raw.(type) {
	case []map[string]any:
		properties = raw
	case []any:
		for _, p := range raw {
			v, ok := p.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("custom block %s: invalid property", entry.Name)
			}
			properties = append(properties, v)
		}
	default:
		return nil, fmt.Errorf("custom block %s: unsupported property definitions", entry.Name)
	}
	for _, prop := range properties {
		name, ok := prop["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("custom block %s: unnamed property", entry.Name)
		}
		var values []any
		switch v := prop["enum"].(type) {
		case []any:
			values = v
		case []string:
			for _, x := range v {
				values = append(values, x)
			}
		case []int32:
			for _, x := range v {
				values = append(values, x)
			}
		case []byte:
			for _, x := range v {
				values = append(values, x)
			}
		default:
			return nil, fmt.Errorf("custom block %s: unsupported enum %s", entry.Name, name)
		}
		if len(values) == 0 || len(states)*len(values) > 65536 {
			return nil, fmt.Errorf("custom block %s: invalid state count", entry.Name)
		}
		var next []mapping.BlockState
		for _, s := range states {
			for _, value := range values {
				props := make(map[string]any, len(s.Properties)+1)
				for k, v := range s.Properties {
					props[k] = v
				}
				props[name] = value
				next = append(next, mapping.BlockState{Name: entry.Name, Properties: props})
			}
		}
		states = next
	}
	return states, nil
}

func (r *NetworkRegistry) BlockCount() int      { return len(r.states) }
func (r *NetworkRegistry) AirRuntimeID() uint32 { return r.air }
func (r *NetworkRegistry) RuntimeIDToState(id uint32) (string, map[string]any, bool) {
	if int64(id) >= int64(len(r.states)) {
		return "", nil, false
	}
	s := r.states[id]
	return s.Name, s.Properties, true
}
func (r *NetworkRegistry) Ordinal(network uint32) (uint32, bool) {
	n, ok := r.toOrdinal[network]
	return n, ok
}
func (r *NetworkRegistry) Network(ordinal uint32) (uint32, bool) {
	if int64(ordinal) >= int64(len(r.toNetwork)) {
		return 0, false
	}
	return r.toNetwork[ordinal], true
}

// NetworkBlockHash hashes sorted, typed little-endian NBT (without a version).
// This matches Dragonfly's network_block_hash.go, not its internal state hash.
func NetworkBlockHash(name string, states map[string]any) (uint32, error) {
	if name == "minecraft:unknown" {
		return 0xfffffffe, nil
	}
	b := []byte{10, 0, 0, 8}
	str := func(v string) { b = binary.LittleEndian.AppendUint16(b, uint16(len(v))); b = append(b, v...) }
	str("name")
	str(name)
	b = append(b, 10)
	str("states")
	keys := make([]string, 0, len(states))
	for k := range states {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch v := states[k].(type) {
		case string:
			b = append(b, 8)
			str(k)
			str(v)
		case uint8:
			b = append(b, 1)
			str(k)
			b = append(b, v)
		case int8:
			b = append(b, 1)
			str(k)
			b = append(b, byte(v))
		case bool:
			b = append(b, 1)
			str(k)
			if v {
				b = append(b, 1)
			} else {
				b = append(b, 0)
			}
		case int16:
			b = append(b, 2)
			str(k)
			b = binary.LittleEndian.AppendUint16(b, uint16(v))
		case uint16:
			b = append(b, 2)
			str(k)
			b = binary.LittleEndian.AppendUint16(b, v)
		case int32:
			b = append(b, 3)
			str(k)
			b = binary.LittleEndian.AppendUint32(b, uint32(v))
		case uint32:
			b = append(b, 3)
			str(k)
			b = binary.LittleEndian.AppendUint32(b, v)
		default:
			return 0, fmt.Errorf("block %s property %s has unsupported type %T", name, k, v)
		}
	}
	b = append(b, 0, 0)
	h := fnv.New32a()
	_, _ = h.Write(b)
	return h.Sum32(), nil
}
