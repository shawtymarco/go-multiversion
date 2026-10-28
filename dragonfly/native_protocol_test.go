package dragonfly

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/df-mc/dragonfly/server/block"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestNativeHashProtocolPreservesRegistrySemantics(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	p, err := NativeProtocol(world.DefaultBlockRegistry)
	if err != nil {
		t.Fatal(err)
	}
	stone := world.BlockRuntimeID(block.Stone{})
	hash, _ := world.DefaultBlockRegistry.RuntimeIDToHash(stone)
	if stone == hash {
		t.Fatal("fixture did not distinguish local IDs from network hashes")
	}
	itemID, meta, _ := world.ItemRuntimeID(block.Stone{})
	stack := protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: itemID, MetadataValue: uint32(meta)}, Count: 1, BlockRuntimeID: int32(stone)}
	instance := protocol.ItemInstance{Stack: stack, StackNetworkID: 9}
	for _, pk := range []packet.Packet{
		&packet.StartGame{},
		&packet.UpdateBlock{NewBlockRuntimeID: stone},
		&packet.InventoryContent{Content: []protocol.ItemInstance{instance}},
		&packet.InventorySlot{NewItem: instance, StorageItem: protocol.Option(instance)},
		&packet.AddPlayer{HeldItem: instance},
		&packet.AddActor{EntityType: "minecraft:falling_block", EntityMetadata: protocol.EntityMetadata{protocol.EntityDataKeyVariant: int32(stone)}},
		&packet.CreativeContent{Groups: []protocol.CreativeGroup{{Icon: stack}}, Items: []protocol.CreativeItem{{CreativeItemNetworkID: 42, Item: stack}}},
		&packet.InventoryTransaction{Actions: []protocol.InventoryAction{{OldItem: instance, NewItem: instance}}, TransactionData: &protocol.UseItemTransactionData{HeldItem: instance, BlockRuntimeID: stone}},
		&packet.PlayerAuthInput{ItemInteractionData: protocol.Option(protocol.UseItemTransactionData{HeldItem: instance, BlockRuntimeID: stone})},
		&packet.LevelEvent{EventType: packet.LevelEventParticlesDestroyBlock, EventData: int32(stone)},
		&packet.LevelSoundEvent{SoundType: packet.SoundEventPlace, ExtraData: int32(stone)},
	} {
		t.Run(fmt.Sprintf("%T", pk), func(t *testing.T) {
			converted := p.ConvertFromLatest(pk, nil)
			if len(converted) != 1 {
				t.Fatal("valid native packet was dropped")
			}
			restored := p.ConvertToLatest(converted[0], nil)
			if len(restored) != 1 || !reflect.DeepEqual(restored[0], pk) {
				t.Fatalf("native semantic roundtrip differs: got %#v, want %#v", restored, pk)
			}
			switch v := converted[0].(type) {
			case *packet.UpdateBlock:
				if v.NewBlockRuntimeID != hash {
					t.Fatal("local block ID leaked to native client")
				}
			case *packet.StartGame:
				if !v.UseBlockNetworkIDHashes {
					t.Fatal("client was not told to use hashes")
				}
			case *packet.InventoryContent:
				if uint32(v.Content[0].Stack.BlockRuntimeID) != hash {
					t.Fatal("inventory block hash was not mapped")
				}
				if pk.(*packet.InventoryContent).Content[0].Stack.BlockRuntimeID != int32(stone) {
					t.Fatal("source inventory mutated")
				}
			}
		})
	}
	for face := int32(0); face < 6; face++ {
		pk := &packet.LevelEvent{EventType: packet.LevelEventParticlesCrackBlock, EventData: int32(stone) | face<<24}
		out := p.ConvertFromLatest(pk, nil)[0].(*packet.LevelEvent)
		if out.EventType != packet.LevelEventParticlesCrackBlockDown+face || uint32(out.EventData) != hash {
			t.Fatal("crack particle truncated its hash or lost its face")
		}
		back := p.ConvertToLatest(out, nil)[0]
		if !reflect.DeepEqual(back, pk) {
			t.Fatal("crack particle semantic roundtrip failed")
		}
	}
	// Ordinary non-block item stacks retain their zero block-ID sentinel.
	id, _, _ := world.ItemRuntimeID(item.Diamond{})
	pk := &packet.MobEquipment{NewItem: protocol.ItemInstance{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: id}, Count: 1}}}
	if !reflect.DeepEqual(p.ConvertFromLatest(pk, nil)[0], pk) {
		t.Fatal("non-block item changed")
	}
}

func TestNativeHashProtocolRejectsUnknownClientBlock(t *testing.T) {
	world.DefaultBlockRegistry.Finalize()
	p, err := NativeProtocol(world.DefaultBlockRegistry)
	if err != nil {
		t.Fatal(err)
	}
	invalid := uint32(0)
	for {
		if _, ok := p.(*nativeProtocol).nativeIDs[invalid]; !ok {
			break
		}
		invalid++
	}
	if result := p.ConvertToLatest(&packet.InventoryTransaction{TransactionData: &protocol.UseItemTransactionData{BlockRuntimeID: invalid}}, nil); len(result) != 0 {
		t.Fatal("unknown client hash reached native gameplay")
	}
}
