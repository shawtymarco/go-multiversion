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

func TestItemMapperResolvesHistoricalAliases(t *testing.T) {
	native := []protocol.ItemEntry{{Name: "minecraft:crimson_door", RuntimeID: 2}}
	target := map[string]TargetItem{"minecraft:item.crimson_door": {RuntimeID: -244, Version: 1}}
	mapper, err := NewItemMapperWithResolver(native, target, func(name string) string {
		if name == "minecraft:item.crimson_door" {
			return "minecraft:crimson_door"
		}
		return name
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := mapper.NativeToTarget(2); !ok || got != -244 {
		t.Fatalf("native alias mapping: got %d/%v", got, ok)
	}
	if got, ok := mapper.TargetToNative(-244); !ok || got != 2 {
		t.Fatalf("target alias mapping: got %d/%v", got, ok)
	}
	if got, ok := mapper.TargetIdentifier(-244); !ok || got != "minecraft:crimson_door" {
		t.Fatalf("resolved target identifier: got %q/%v", got, ok)
	}
	if got, ok := mapper.TargetWireIdentifier("minecraft:crimson_door"); !ok || got != "minecraft:item.crimson_door" {
		t.Fatalf("target wire identifier: got %q/%v", got, ok)
	}
	if got, ok := mapper.TargetSemanticIdentifier("minecraft:item.crimson_door"); !ok || got != "minecraft:crimson_door" {
		t.Fatalf("target semantic identifier: got %q/%v", got, ok)
	}
	wantEntries := []protocol.ItemEntry{{Name: "minecraft:item.crimson_door", RuntimeID: -244, Version: 1}}
	if got := mapper.TargetEntries(); !reflect.DeepEqual(got, wantEntries) {
		t.Fatalf("target alias entry changed on wire: got %#v, want %#v", got, wantEntries)
	}
}

func TestBlockMapperPreservesTargetOrder(t *testing.T) {
	native := testBlockRegistry{states: []BlockState{
		{Name: "minecraft:air"},
		{Name: "minecraft:stone"},
	}}
	target := []BlockState{
		{Name: "minecraft:stone"},
		{Name: "minecraft:air"},
	}
	mapper, err := NewBlockMapperWithTargetOrder(native, target)
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := mapper.TargetRuntimeID("minecraft:stone", nil); !ok || got != 0 {
		t.Fatalf("stone target runtime ID: got %d/%v, want 0/true", got, ok)
	}
	if got, ok := mapper.TargetRuntimeID("minecraft:air", nil); !ok || got != 1 {
		t.Fatalf("air target runtime ID: got %d/%v, want 1/true", got, ok)
	}
	states := mapper.TargetStates()
	if len(states) != 2 || states[0].Name != "minecraft:stone" || states[1].Name != "minecraft:air" {
		t.Fatalf("target state order changed: %#v", states)
	}
}

func TestItemMapperAllowsExplicitTargetOnlyEntries(t *testing.T) {
	native := []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 2}}
	target := map[string]TargetItem{
		"minecraft:stone":       {RuntimeID: 7},
		"minecraft:old_feature": {RuntimeID: -9},
	}
	mapper, err := NewItemMapperAllowingTargetOnly(native, target, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := mapper.TargetToNative(-9); ok {
		t.Fatal("target-only item unexpectedly mapped serverbound")
	}
	if got := mapper.TargetEntries(); len(got) != 2 {
		t.Fatalf("advertised target registry size: got %d, want 2", len(got))
	}
	fallbacks := mapper.TargetFallbacks()
	if len(fallbacks) != 1 || fallbacks[0].TargetRuntimeID != -9 || fallbacks[0].WireName != "minecraft:old_feature" {
		t.Fatalf("target-only fallback report: %#v", fallbacks)
	}
}
