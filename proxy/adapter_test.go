package proxy

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func boundAdapter(t testing.TB, version string, hashed bool, itemID int16) *Adapter {
	t.Helper()
	a, err := NewAdapter(version)
	if err != nil {
		t.Fatal(err)
	}
	err = a.BindUpstream(UpstreamSnapshot{StartGame: &packet.StartGame{UseBlockNetworkIDHashes: hashed}, ItemRegistry: &packet.ItemRegistry{Items: []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: itemID}, {Name: "minecraft:shield", RuntimeID: itemID + 1}, {Name: "minecraft:diamond_sword", RuntimeID: itemID + 2}}}})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func BenchmarkPlayerInputTranslation(b *testing.B) {
	a := boundAdapter(b, "1.18.12", true, 40)
	pk := &packet.PlayerAuthInput{InputData: protocol.NewInputFlags(packet.InputFlagCount), InputMode: packet.InputModeMouse}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if _, err := a.Translate(ClientToServer, pk); err != nil {
			b.Fatal(err)
		}
	}
}

func stateID(t *testing.T, a *Adapter, name string) uint32 {
	t.Helper()
	for i, s := range a.registry.states {
		if s.Name == name {
			return uint32(i)
		}
	}
	t.Fatalf("missing %s", name)
	return 0
}

func TestSessionRegistryIsolationAndRoundTrip(t *testing.T) {
	for _, version := range []string{"1.18.2", "1.18.12"} {
		t.Run(version, func(t *testing.T) {
			a := boundAdapter(t, version, true, 40)
			b := boundAdapter(t, version, true, 120)
			stone := stateID(t, a, "minecraft:stone")
			network, _ := a.registry.Network(stone)
			original := &packet.UpdateBlock{NewBlockRuntimeID: network}
			copy := *original
			out, err := a.Translate(ServerToClient, original)
			if err != nil {
				t.Fatal(err)
			}
			target := out[0].(*packet.UpdateBlock)
			state, ok := a.blocks.TargetState(target.NewBlockRuntimeID)
			if !ok || state.Name != "minecraft:stone" {
				t.Fatalf("wrong target block %+v", state)
			}
			back, err := a.Translate(ClientToServer, target)
			if err != nil {
				t.Fatal(err)
			}
			if back[0].(*packet.UpdateBlock).NewBlockRuntimeID != network || *original != copy {
				t.Fatal("block mapping changed identity/input")
			}
			input := &packet.InventorySlot{WindowID: 0, Slot: 0, NewItem: protocol.ItemInstance{StackNetworkID: 101, Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 42}, Count: 1}}}
			out, err = a.Translate(ServerToClient, input)
			if err != nil {
				t.Fatal(err)
			}
			back, err = a.Translate(ClientToServer, out[0])
			if err != nil {
				t.Fatal(err)
			}
			if back[0].(*packet.InventorySlot).NewItem.Stack.NetworkID != 42 {
				t.Fatal("item roundtrip")
			}
			if _, err = b.Translate(ServerToClient, input); err == nil {
				t.Fatal("another server's item ID accepted")
			}
		})
	}
}

func TestChunkPaletteRepackingAndReuse(t *testing.T) {
	var encoded bytes.Buffer
	encoded.WriteByte(3)
	words := make([]uint32, 128)
	for i := range words {
		words[i] = 0xaaaaaaaa
	}
	_ = binary.Write(&encoded, binary.LittleEndian, words)
	_ = protocol.WriteVarint32(&encoded, 2)
	_ = protocol.WriteVarint32(&encoded, 10)
	_ = protocol.WriteVarint32(&encoded, 20)
	got, err := rewriteStorage(bytes.NewReader(encoded.Bytes()), func(uint32) (uint32, error) { return 7, nil }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte{1, 14}) {
		t.Fatalf("dedup=%x", got)
	}
	reused, err := rewriteStorage(bytes.NewReader([]byte{0xff}), func(v uint32) (uint32, error) { return v, nil }, got)
	if err != nil || !bytes.Equal(got, reused) {
		t.Fatal("biome reuse")
	}
	if _, err := rewriteStorage(bytes.NewReader([]byte{0xff}), nil, nil); err == nil {
		t.Fatal("accepted leading reuse")
	}
	if _, err := rewriteStorage(bytes.NewReader([]byte{15}), nil, nil); err == nil {
		t.Fatal("accepted invalid bits")
	}
}

func TestSubChunkNegativeYAndAllAir475(t *testing.T) {
	a := boundAdapter(t, "1.18.2", false, 40)
	pk := &packet.SubChunk{Position: protocol.SubChunkPos{3, -4, 9}, SubChunkEntries: []protocol.SubChunkEntry{{Result: protocol.SubChunkResultSuccessAllAir}}}
	out, err := a.Translate(ServerToClient, pk)
	if err != nil {
		t.Fatal(err)
	}
	entry := out[0].(*packet.SubChunk).SubChunkEntries[0]
	payload, _ := entry.RawPayload.Value()
	if entry.Result != protocol.SubChunkResultSuccess || !bytes.Equal(payload, []byte{9, 0, 252}) {
		t.Fatalf("all air %v %x", entry.Result, payload)
	}
	if pk.SubChunkEntries[0].Result != protocol.SubChunkResultSuccessAllAir {
		t.Fatal("mutated source")
	}
}

func TestMovementPreservesTickAndCopiesFlags(t *testing.T) {
	a := boundAdapter(t, "1.18.12", false, 40)
	flags := protocol.NewInputFlags(packet.InputFlagCount)
	flags.Set(packet.InputFlagJumpDown)
	pk := &packet.PlayerAuthInput{Tick: 456, InputData: flags, InputMode: packet.InputModeMouse}
	out, err := a.Translate(ClientToServer, pk)
	if err != nil {
		t.Fatal(err)
	}
	got := out[0].(*packet.PlayerAuthInput)
	if got.Tick != 456 || !got.InputData.Load(packet.InputFlagJumpPressedRaw) || got.CameraOrientation[2] != 1 {
		t.Fatal("input semantics lost")
	}
	if pk.InputData.Load(packet.InputFlagJumpPressedRaw) {
		t.Fatal("mutated input flags")
	}
}

func TestCloneMutableOptionals(t *testing.T) {
	pk := &packet.PlayerAuthInput{ItemInteractionData: protocol.Option(protocol.UseItemTransactionData{HeldItem: protocol.ItemInstance{Stack: protocol.ItemStack{NBTData: map[string]any{"name": "original"}}}})}
	clone := clonePacket(pk).(*packet.PlayerAuthInput)
	item, _ := clone.ItemInteractionData.Value()
	item.HeldItem.Stack.NBTData["name"] = "changed"
	before, _ := pk.ItemInteractionData.Value()
	if before.HeldItem.Stack.NBTData["name"] != "original" {
		t.Fatal("optional payload alias")
	}
}

func TestNativeHashUsesTypedStates(t *testing.T) {
	a, _ := NetworkBlockHash("minecraft:test", map[string]any{"x": int32(1), "y": byte(0)})
	b, _ := NetworkBlockHash("minecraft:test", map[string]any{"y": byte(0), "x": int32(1)})
	c, _ := NetworkBlockHash("minecraft:test", map[string]any{"x": byte(1), "y": byte(0)})
	if a != b || a == c {
		t.Fatal("non-canonical typed hash")
	}
}

func TestRegistryBindInputPreserved(t *testing.T) {
	a, _ := NewAdapter("1.18.12")
	pk := &packet.StartGame{Blocks: []protocol.BlockEntry{{Name: "example:block", Properties: map[string]any{"components": map[string]any{}}}}}
	before := clonePacket(pk)
	err := a.BindUpstream(UpstreamSnapshot{StartGame: pk, ItemRegistry: &packet.ItemRegistry{Items: []protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 1}}}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pk, before) {
		t.Fatal("bootstrap mutated")
	}
}
