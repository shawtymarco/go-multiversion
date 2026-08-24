package mapping

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"

	"github.com/df-mc/worldupgrader/blockupgrader"
)

// BlockRegistry is the native registry surface required to build a direct
// protocol mapping. Dragonfly's world.BlockRegistry satisfies this interface.
type BlockRegistry interface {
	BlockCount() int
	AirRuntimeID() uint32
	RuntimeIDToState(runtimeID uint32) (name string, properties map[string]any, found bool)
}

// BlockState is one historical runtime block state.
type BlockState struct {
	Name       string
	Properties map[string]any
	Version    int32
}

// BlockFallback records one native state that protocol 1001 cannot represent.
type BlockFallback struct {
	NativeRuntimeID uint32
	Name            string
	Properties      map[string]any
	TargetRuntimeID uint32
}

// BlockMapper maps current native runtime IDs directly to historical target
// runtime IDs and back using identifier plus complete typed properties.
type BlockMapper struct {
	nativeToTarget []uint32
	nativeExact    []bool
	targetToNative []uint32
	targetStates   []BlockState
	targetByKey    map[string]uint32
	fallbacks      []BlockFallback
	targetAir      uint32
}

// NewBlockMapper builds an immutable direct mapping. Historical states are
// ordered using the stable name-only FNV-1 sort used by recent Dragonfly block
// registries before their runtime IDs are assigned.
func NewBlockMapper(native BlockRegistry, historical []BlockState) (*BlockMapper, error) {
	return newBlockMapper(native, historical, true)
}

// NewBlockMapperWithTargetOrder builds an immutable direct mapping while
// preserving the input historical state order as the target runtime-ID order.
// It is used by releases whose Dragonfly registry registered the embedded NBT
// stream directly without a final sorting pass.
func NewBlockMapperWithTargetOrder(native BlockRegistry, historical []BlockState) (*BlockMapper, error) {
	return newBlockMapper(native, historical, false)
}

func newBlockMapper(native BlockRegistry, historical []BlockState, sortTarget bool) (*BlockMapper, error) {
	if native == nil {
		return nil, fmt.Errorf("native block registry is nil")
	}
	targetStates := append([]BlockState(nil), historical...)
	if sortTarget {
		sort.SliceStable(targetStates, func(i, j int) bool {
			return fnv1String(targetStates[i].Name) < fnv1String(targetStates[j].Name)
		})
	}

	targetByKey := make(map[string]uint32, len(targetStates))
	targetAir, foundAir := uint32(0), false
	for index, state := range targetStates {
		key, err := StateKey(state.Name, state.Properties)
		if err != nil {
			return nil, fmt.Errorf("target runtime ID %d: %w", index, err)
		}
		if _, exists := targetByKey[key]; exists {
			return nil, fmt.Errorf("duplicate target block state %s at runtime ID %d", state.Name, index)
		}
		targetByKey[key] = uint32(index)
		if state.Name == "minecraft:air" {
			targetAir, foundAir = uint32(index), true
		}
	}
	if !foundAir {
		return nil, fmt.Errorf("target block registry has no minecraft:air state")
	}

	nativeByKey := make(map[string]uint32, native.BlockCount())
	nativeStates := make([]BlockState, native.BlockCount())
	for index := range native.BlockCount() {
		runtimeID := uint32(index)
		name, properties, ok := native.RuntimeIDToState(runtimeID)
		if !ok {
			return nil, fmt.Errorf("native runtime ID %d has no state", runtimeID)
		}
		key, err := StateKey(name, properties)
		if err != nil {
			return nil, fmt.Errorf("native runtime ID %d: %w", runtimeID, err)
		}
		if _, exists := nativeByKey[key]; exists {
			return nil, fmt.Errorf("duplicate native block state %s at runtime ID %d", name, runtimeID)
		}
		nativeByKey[key] = runtimeID
		nativeStates[index] = BlockState{Name: name, Properties: cloneProperties(properties)}
	}

	targetToNative := make([]uint32, len(targetStates))
	unresolved := make([]string, 0)
	for targetRuntimeID, state := range targetStates {
		key, _ := StateKey(state.Name, state.Properties)
		nativeRuntimeID, ok := nativeByKey[key]
		if !ok {
			upgraded := blockupgrader.Upgrade(blockupgrader.BlockState{
				Name:       state.Name,
				Properties: cloneProperties(state.Properties),
				Version:    state.Version,
			})
			upgradedKey, err := StateKey(upgraded.Name, upgraded.Properties)
			if err != nil {
				return nil, fmt.Errorf("upgraded target runtime ID %d: %w", targetRuntimeID, err)
			}
			nativeRuntimeID, ok = nativeByKey[upgradedKey]
			if !ok {
				if len(unresolved) < 32 {
					unresolved = append(unresolved, fmt.Sprintf("%s%v (RID %d) -> %s%v", state.Name, state.Properties, targetRuntimeID, upgraded.Name, upgraded.Properties))
				}
				continue
			}
		}
		targetToNative[targetRuntimeID] = nativeRuntimeID
	}
	if len(unresolved) != 0 {
		return nil, fmt.Errorf("target registry has unresolved native states:\n%s", strings.Join(unresolved, "\n"))
	}

	nativeToTarget := make([]uint32, native.BlockCount())
	nativeExact := make([]bool, native.BlockCount())
	for runtimeID := range nativeToTarget {
		nativeToTarget[runtimeID] = targetAir
		state := nativeStates[runtimeID]
		key, _ := StateKey(state.Name, state.Properties)
		if targetRuntimeID, ok := targetByKey[key]; ok {
			nativeToTarget[runtimeID], nativeExact[runtimeID] = targetRuntimeID, true
		}
	}
	for targetRuntimeID, nativeRuntimeID := range targetToNative {
		if !nativeExact[nativeRuntimeID] {
			nativeToTarget[nativeRuntimeID], nativeExact[nativeRuntimeID] = uint32(targetRuntimeID), true
		}
	}
	fallbacks := make([]BlockFallback, 0)
	for runtimeID, exact := range nativeExact {
		if exact {
			continue
		}
		state := nativeStates[runtimeID]
		fallbacks = append(fallbacks, BlockFallback{
			NativeRuntimeID: uint32(runtimeID),
			Name:            state.Name,
			Properties:      cloneProperties(state.Properties),
			TargetRuntimeID: targetAir,
		})
	}

	return &BlockMapper{
		nativeToTarget: nativeToTarget,
		nativeExact:    nativeExact,
		targetToNative: targetToNative,
		targetStates:   targetStates,
		targetByKey:    targetByKey,
		fallbacks:      fallbacks,
		targetAir:      targetAir,
	}, nil
}

// NativeToTarget maps a current runtime ID to protocol 1001. exact is false
// when the explicit minecraft:air fallback was used.
func (m *BlockMapper) NativeToTarget(runtimeID uint32) (target uint32, exact bool) {
	if m == nil || runtimeID >= uint32(len(m.nativeToTarget)) {
		if m == nil {
			return 0, false
		}
		return m.targetAir, false
	}
	return m.nativeToTarget[runtimeID], m.nativeExact[runtimeID]
}

// MapNative maps a current runtime ID and distinguishes an explicit fallback
// from an out-of-range ID.
func (m *BlockMapper) MapNative(runtimeID uint32) (target uint32, valid bool, exact bool) {
	if m == nil || runtimeID >= uint32(len(m.nativeToTarget)) {
		return 0, false, false
	}
	return m.nativeToTarget[runtimeID], true, m.nativeExact[runtimeID]
}

// TargetToNative maps a protocol-1001 runtime ID to the exact current state.
func (m *BlockMapper) TargetToNative(runtimeID uint32) (uint32, bool) {
	if m == nil || runtimeID >= uint32(len(m.targetToNative)) {
		return 0, false
	}
	return m.targetToNative[runtimeID], true
}

// TargetRuntimeID resolves a historical state by identifier and properties.
func (m *BlockMapper) TargetRuntimeID(name string, properties map[string]any) (uint32, bool) {
	if m == nil {
		return 0, false
	}
	key, err := StateKey(name, properties)
	if err != nil {
		return 0, false
	}
	runtimeID, ok := m.targetByKey[key]
	return runtimeID, ok
}

// TargetStates returns the historical states in runtime-ID order.
func (m *BlockMapper) TargetStates() []BlockState {
	if m == nil {
		return nil
	}
	states := make([]BlockState, len(m.targetStates))
	for index, state := range m.targetStates {
		states[index] = BlockState{Name: state.Name, Properties: cloneProperties(state.Properties), Version: state.Version}
	}
	return states
}

// Fallbacks returns every native state substituted with historical air.
func (m *BlockMapper) Fallbacks() []BlockFallback {
	if m == nil {
		return nil
	}
	fallbacks := make([]BlockFallback, len(m.fallbacks))
	for index, fallback := range m.fallbacks {
		fallbacks[index] = fallback
		fallbacks[index].Properties = cloneProperties(fallback.Properties)
	}
	return fallbacks
}

// TargetAir returns the protocol-1001 air runtime ID.
func (m *BlockMapper) TargetAir() uint32 {
	if m == nil {
		return 0
	}
	return m.targetAir
}

// NativeCount returns the current registry size used to build the mapping.
func (m *BlockMapper) NativeCount() int {
	if m == nil {
		return 0
	}
	return len(m.nativeToTarget)
}

// TargetCount returns the historical registry size.
func (m *BlockMapper) TargetCount() int {
	if m == nil {
		return 0
	}
	return len(m.targetToNative)
}

func cloneProperties(properties map[string]any) map[string]any {
	if properties == nil {
		return nil
	}
	cloned := make(map[string]any, len(properties))
	for key, value := range properties {
		cloned[key] = value
	}
	return cloned
}

func fnv1String(value string) uint64 {
	hash := fnv.New64()
	_, _ = hash.Write([]byte(value))
	return hash.Sum64()
}
