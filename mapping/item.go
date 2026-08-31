package mapping

import (
	"fmt"
	"math"
	"sort"
	"strings"

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

// TargetItemFallback records one historical target item that has no semantic
// item in the current native registry and is therefore rejected serverbound.
type TargetItemFallback struct {
	TargetRuntimeID int32  `json:"target_runtime_id"`
	WireName        string `json:"wire_name"`
	ResolvedName    string `json:"resolved_name"`
}

// ItemMapper maps current and target item network IDs by semantic identifier.
type ItemMapper struct {
	nativeToTarget       map[int32]int32
	targetToNative       map[int32]int32
	nativeByName         map[string]int32
	nativeNames          map[int32]string
	targetNames          map[int32]string
	targetByName         map[string]int32
	targetByWire         map[string]int32
	targetWireByName     map[string]string
	targetResolvedByWire map[string]string
	targetEntries        []protocol.ItemEntry
	fallbacks            []ItemFallback
	targetFallbacks      []TargetItemFallback
}

// NewItemMapper builds an immutable item mapping from the current ItemRegistry
// packet and the historical identifier-keyed snapshot.
func NewItemMapper(native []protocol.ItemEntry, target map[string]TargetItem) (*ItemMapper, error) {
	return NewItemMapperWithResolver(native, target, func(name string) string { return name })
}

// NewItemMapperWithResolver builds an immutable item mapping after resolving
// historical target identifiers to their current semantic identifiers. The
// original target identifiers remain unchanged in TargetEntries.
func NewItemMapperWithResolver(native []protocol.ItemEntry, target map[string]TargetItem, resolve func(string) string) (*ItemMapper, error) {
	return newItemMapperWithResolver(native, target, resolve, false)
}

// NewItemMapperAllowingTargetOnly builds a mapping that keeps historical
// target-only entries in the advertised registry while rejecting them when
// received serverbound. Use only when the exact target registry contains
// removed or experimental items with no valid native semantic replacement.
func NewItemMapperAllowingTargetOnly(native []protocol.ItemEntry, target map[string]TargetItem, resolve func(string) string) (*ItemMapper, error) {
	return newItemMapperWithResolver(native, target, resolve, true)
}

func newItemMapperWithResolver(native []protocol.ItemEntry, target map[string]TargetItem, resolve func(string) string, allowTargetOnly bool) (*ItemMapper, error) {
	if resolve == nil {
		resolve = func(name string) string { return name }
	}
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
	targetByWire := make(map[string]int32, len(target))
	targetSourceNames := make(map[string]string, len(target))
	targetResolvedByWire := make(map[string]string, len(target))
	targetNames := make(map[int32]string, len(target))
	targetWireNames := make(map[int32]string, len(target))
	targetEntries := make([]protocol.ItemEntry, 0, len(target))
	targetIdentifiers := make([]string, 0, len(target))
	for name := range target {
		targetIdentifiers = append(targetIdentifiers, name)
	}
	sort.Strings(targetIdentifiers)
	for _, name := range targetIdentifiers {
		entry := target[name]
		if entry.RuntimeID < math.MinInt16 || entry.RuntimeID > math.MaxInt16 {
			return nil, fmt.Errorf("target item %s runtime ID %d exceeds int16", name, entry.RuntimeID)
		}
		if other, ok := targetWireNames[entry.RuntimeID]; ok && other != name {
			return nil, fmt.Errorf("target item runtime ID %d is shared by %s and %s", entry.RuntimeID, other, name)
		}
		resolvedName := resolve(name)
		if resolvedName == "" {
			return nil, fmt.Errorf("target item %s resolved to an empty identifier", name)
		}
		targetWireNames[entry.RuntimeID], targetNames[entry.RuntimeID] = name, resolvedName
		targetByWire[name] = entry.RuntimeID
		targetResolvedByWire[name] = resolvedName
		if previousID, exists := targetByName[resolvedName]; !exists || (name == resolvedName && targetSourceNames[resolvedName] != resolvedName) || (name != resolvedName && targetSourceNames[resolvedName] != resolvedName && entry.RuntimeID < previousID) {
			targetByName[resolvedName] = entry.RuntimeID
			targetSourceNames[resolvedName] = name
		}
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
	targetRuntimeIDs := make([]int, 0, len(targetNames))
	for targetRuntimeID := range targetNames {
		targetRuntimeIDs = append(targetRuntimeIDs, int(targetRuntimeID))
	}
	sort.Ints(targetRuntimeIDs)
	var missing []string
	targetFallbacks := make([]TargetItemFallback, 0)
	for _, rawRuntimeID := range targetRuntimeIDs {
		targetRuntimeID := int32(rawRuntimeID)
		name := targetNames[targetRuntimeID]
		nativeRuntimeID, ok := nativeByName[name]
		if !ok {
			missing = append(missing, fmt.Sprintf("%d=%s->%s", targetRuntimeID, targetWireNames[targetRuntimeID], name))
			targetFallbacks = append(targetFallbacks, TargetItemFallback{TargetRuntimeID: targetRuntimeID, WireName: targetWireNames[targetRuntimeID], ResolvedName: name})
			continue
		}
		targetToNative[targetRuntimeID] = nativeRuntimeID
	}
	if len(missing) != 0 && !allowTargetOnly {
		return nil, fmt.Errorf("%d target items have no native item: %s", len(missing), strings.Join(missing, ", "))
	}

	return &ItemMapper{
		nativeToTarget:       nativeToTarget,
		targetToNative:       targetToNative,
		nativeByName:         nativeByName,
		nativeNames:          nativeNames,
		targetNames:          targetNames,
		targetByName:         targetByName,
		targetByWire:         targetByWire,
		targetWireByName:     targetSourceNames,
		targetResolvedByWire: targetResolvedByWire,
		targetEntries:        targetEntries,
		fallbacks:            fallbacks,
		targetFallbacks:      targetFallbacks,
	}, nil
}

// TargetWireRuntimeID resolves the identifier exactly as advertised by the
// historical item registry. Historical block state names use these wire names
// even when item upgrading resolves them to a flattened current identifier.
func (m *ItemMapper) TargetWireRuntimeID(name string) (int32, bool) {
	if m == nil {
		return 0, false
	}
	runtimeID, ok := m.targetByWire[name]
	return runtimeID, ok
}

// NativeRuntimeID resolves a current semantic item identifier to its network
// ID. It is used for historical block-items whose old registry exposed one
// generic item identifier while BlockRuntimeID carried the concrete variant.
func (m *ItemMapper) NativeRuntimeID(name string) (int32, bool) {
	if m == nil {
		return 0, false
	}
	runtimeID, ok := m.nativeByName[name]
	return runtimeID, ok
}

// TargetFallbacks returns historical target entries that are advertised for
// registry fidelity but rejected serverbound because native has no equivalent.
func (m *ItemMapper) TargetFallbacks() []TargetItemFallback {
	if m == nil {
		return nil
	}
	return append([]TargetItemFallback(nil), m.targetFallbacks...)
}

// TargetWireIdentifier resolves a current semantic item identifier to the
// historical identifier advertised in the target ItemRegistry.
func (m *ItemMapper) TargetWireIdentifier(name string) (string, bool) {
	if m == nil {
		return "", false
	}
	wire, ok := m.targetWireByName[name]
	return wire, ok
}

// TargetSemanticIdentifier resolves a historical wire identifier to the
// current semantic identifier.
func (m *ItemMapper) TargetSemanticIdentifier(name string) (string, bool) {
	if m == nil {
		return "", false
	}
	resolved, ok := m.targetResolvedByWire[name]
	return resolved, ok
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
