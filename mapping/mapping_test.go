package mapping

import (
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

type testBlockRegistry struct {
	states []BlockState
	air    uint32
}

func (r testBlockRegistry) BlockCount() int { return len(r.states) }
func (r testBlockRegistry) AirRuntimeID() uint32 {
	return r.air
}
func (r testBlockRegistry) RuntimeIDToState(runtimeID uint32) (string, map[string]any, bool) {
	if runtimeID >= uint32(len(r.states)) {
		return "", nil, false
	}
	state := r.states[runtimeID]
	return state.Name, cloneProperties(state.Properties), true
}

func TestBlockMapperExactAndFallback(t *testing.T) {
	native := testBlockRegistry{states: []BlockState{
		{Name: "minecraft:air"},
		{Name: "minecraft:stone", Properties: map[string]any{"kind": int32(1)}},
		{Name: "minecraft:new_block", Properties: map[string]any{"enabled": uint8(1)}},
	}}
	target := []BlockState{
		{Name: "minecraft:stone", Properties: map[string]any{"kind": int32(1)}},
		{Name: "minecraft:air"},
	}
	mapper, err := NewBlockMapper(native, target)
	if err != nil {
		t.Fatal(err)
	}
	stone, exact := mapper.NativeToTarget(1)
	if !exact {
		t.Fatal("stone unexpectedly used fallback")
	}
	if got, ok := mapper.TargetToNative(stone); !ok || got != 1 {
		t.Fatalf("stone reverse mapping: got %d/%v", got, ok)
	}
	if got, exact := mapper.NativeToTarget(2); exact || got != mapper.TargetAir() {
		t.Fatalf("new block fallback: got %d/%v, want air/false", got, exact)
	}
	if got := mapper.Fallbacks(); len(got) != 1 || got[0].Name != "minecraft:new_block" {
		t.Fatalf("fallbacks: %#v", got)
	}
}

func TestStateKeyPreservesTypesAndOrder(t *testing.T) {
	left, err := StateKey("minecraft:test", map[string]any{"b": int32(1), "a": uint8(1)})
	if err != nil {
		t.Fatal(err)
	}
	right, err := StateKey("minecraft:test", map[string]any{"a": uint8(1), "b": int32(1)})
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatal("property iteration order changed state key")
	}
	different, _ := StateKey("minecraft:test", map[string]any{"a": int32(1), "b": int32(1)})
	if left == different {
		t.Fatal("state key discarded numeric NBT type")
	}
	boolean, _ := StateKey("minecraft:test", map[string]any{"a": true, "b": int32(1)})
	if left != boolean {
		t.Fatal("state key did not normalise Dragonfly bool to the BDS byte representation")
	}
}

func TestItemMapperRoundTripAndFallback(t *testing.T) {
	native := []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 2}, {Name: "minecraft:new", RuntimeID: 3}}
	target := map[string]TargetItem{"minecraft:stone": {RuntimeID: 7, Version: 1}}
	mapper, err := NewItemMapper(native, target)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := mapper.NativeToTarget(2); !ok || got != 7 {
		t.Fatalf("native item mapping: got %d/%v", got, ok)
	}
	if got, ok := mapper.TargetToNative(7); !ok || got != 2 {
		t.Fatalf("target item mapping: got %d/%v", got, ok)
	}
	if _, ok := mapper.NativeToTarget(3); ok {
		t.Fatal("new item unexpectedly mapped")
	}
	wantEntries := []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 7, Version: 1}}
	if got := mapper.TargetEntries(); !reflect.DeepEqual(got, wantEntries) {
		t.Fatalf("target entries: got %#v, want %#v", got, wantEntries)
	}
}
