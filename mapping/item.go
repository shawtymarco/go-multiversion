package mapping

import (
	"fmt"
	"math"
	"sort"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// TargetItem is one historical identifier-to-network-ID entry.
type TargetItem struct {
	RuntimeID      int32
	ComponentBased bool
	Version        int32
	Data           map[string]any
}

// ItemFallback records one native item missing from the target registry.
type ItemFallback struct {
	NativeRuntimeID int32
	Name            string
}

// ItemMapper maps current and target item network IDs by exact identifier.
type ItemMapper struct {
	nativeToTarget map[int32]int32
	targetToNative map[int32]int32
	nativeNames    map[int32]string
	targetNames    map[int32]string
	targetByName   map[string]int32
	targetEntries  []protocol.ItemEntry
	fallbacks      []ItemFallback
}

// NewItemMapper builds an immutable item mapping from the current ItemRegistry
// packet and the historical identifier-keyed snapshot.
func NewItemMapper(native []protocol.ItemEntry, target map[string]TargetItem) (*ItemMapper, error) {
	nativeByName := make(map[string]int32, len(native))
	nativeNames := make(map[int32]string, len(native))
	for _, entry := range native {
		runtimeID := int32(entry.RuntimeID)
		if other, ok := nativeNames[runtimeID]; ok && other != entry.Name {
			return nil, fmt.Errorf("native item runtime ID %d is shared by %s and %s", runtimeID, other, entry.Name)
		}
		nativeByName[entry.Name], nativeNames[runtimeID] = runtimeID, entry.Name
	}

	targetByName := make(map[string]int32, len(target))
	targetNames := make(map[int32]string, len(target))
	targetEntries := make([]protocol.ItemEntry, 0, len(target))
	for name, entry := range target {
		if entry.RuntimeID < math.MinInt16 || entry.RuntimeID > math.MaxInt16 {
			return nil, fmt.Errorf("target item %s runtime ID %d exceeds int16", name, entry.RuntimeID)
		}
		if other, ok := targetNames[entry.RuntimeID]; ok && other != name {
			return nil, fmt.Errorf("target item runtime ID %d is shared by %s and %s", entry.RuntimeID, other, name)
		}
		targetByName[name], targetNames[entry.RuntimeID] = entry.RuntimeID, name
		targetEntries = append(targetEntries, protocol.ItemEntry{
			Name:           name,
			RuntimeID:      int16(entry.RuntimeID),
			ComponentBased: entry.ComponentBased,
			Version:        entry.Version,
			Data:           cloneProperties(entry.Data),
		})
	}
	sort.Slice(targetEntries, func(i, j int) bool { return targetEntries[i].RuntimeID < targetEntries[j].RuntimeID })

	nativeToTarget := make(map[int32]int32, len(native))
	fallbacks := make([]ItemFallback, 0)
	for _, entry := range native {
		nativeRuntimeID := int32(entry.RuntimeID)
		if targetRuntimeID, ok := targetByName[entry.Name]; ok {
			nativeToTarget[nativeRuntimeID] = targetRuntimeID
			continue
		}
		fallbacks = append(fallbacks, ItemFallback{NativeRuntimeID: nativeRuntimeID, Name: entry.Name})
	}
	sort.Slice(fallbacks, func(i, j int) bool {
		if fallbacks[i].NativeRuntimeID == fallbacks[j].NativeRuntimeID {
			return fallbacks[i].Name < fallbacks[j].Name
		}
		return fallbacks[i].NativeRuntimeID < fallbacks[j].NativeRuntimeID
	})

	targetToNative := make(map[int32]int32, len(target))
	for name, targetRuntimeID := range targetByName {
		nativeRuntimeID, ok := nativeByName[name]
		if !ok {
			return nil, fmt.Errorf("target item %s at runtime ID %d has no native item", name, targetRuntimeID)
		}
		targetToNative[targetRuntimeID] = nativeRuntimeID
	}

	return &ItemMapper{
		nativeToTarget: nativeToTarget,
		targetToNative: targetToNative,
		nativeNames:    nativeNames,
		targetNames:    targetNames,
		targetByName:   targetByName,
		targetEntries:  targetEntries,
		fallbacks:      fallbacks,
	}, nil
}

// NativeToTarget maps an item network ID. Air remains network ID zero.
func (m *ItemMapper) NativeToTarget(runtimeID int32) (int32, bool) {
	if runtimeID == 0 {
		return 0, true
	}
	if m == nil {
		return 0, false
	}
	target, ok := m.nativeToTarget[runtimeID]
	return target, ok
}

// TargetToNative maps a protocol-1001 item network ID.
func (m *ItemMapper) TargetToNative(runtimeID int32) (int32, bool) {
	if runtimeID == 0 {
		return 0, true
	}
	if m == nil {
		return 0, false
	}
	native, ok := m.targetToNative[runtimeID]
	return native, ok
}

// TargetRuntimeID resolves an item identifier in the historical registry.
func (m *ItemMapper) TargetRuntimeID(name string) (int32, bool) {
	if m == nil {
		return 0, false
	}
	runtimeID, ok := m.targetByName[name]
	return runtimeID, ok
}

// TargetIdentifier resolves a historical item network ID.
func (m *ItemMapper) TargetIdentifier(runtimeID int32) (string, bool) {
	if m == nil {
		return "", false
	}
	name, ok := m.targetNames[runtimeID]
	return name, ok
}

// NativeIdentifier resolves a current item network ID.
func (m *ItemMapper) NativeIdentifier(runtimeID int32) (string, bool) {
	if m == nil {
		return "", false
	}
	name, ok := m.nativeNames[runtimeID]
	return name, ok
}

// TargetEntries returns a deep copy of the historical ItemRegistry table.
func (m *ItemMapper) TargetEntries() []protocol.ItemEntry {
	if m == nil {
		return nil
	}
	entries := make([]protocol.ItemEntry, len(m.targetEntries))
	for index, entry := range m.targetEntries {
		entries[index] = entry
		entries[index].Data = cloneProperties(entry.Data)
	}
	return entries
}

// Fallbacks returns native items omitted from protocol 1001.
func (m *ItemMapper) Fallbacks() []ItemFallback {
	if m == nil {
		return nil
	}
	return append([]ItemFallback(nil), m.fallbacks...)
}

// NativeCount returns the current item registry size used by the mapper.
func (m *ItemMapper) NativeCount() int {
	if m == nil {
		return 0
	}
	return len(m.nativeNames)
}

// TargetCount returns the historical item registry size.
func (m *ItemMapper) TargetCount() int {
	if m == nil {
		return 0
	}
	return len(m.targetNames)
}
