package v1_26_45

import (
	"reflect"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/shawtymarco/go-multiversion/mapping"
)

type mappingTestRegistry struct {
	states []mapping.BlockState
}

func (r mappingTestRegistry) BlockCount() int    { return len(r.states) }
func (mappingTestRegistry) AirRuntimeID() uint32 { return 0 }
func (r mappingTestRegistry) RuntimeIDToState(runtimeID uint32) (string, map[string]any, bool) {
	if runtimeID >= uint32(len(r.states)) {
		return "", nil, false
	}
	state := r.states[runtimeID]
	return state.Name, state.Properties, true
}

func testMappedProtocol(t *testing.T) Protocol {
	t.Helper()
	nativeBlocks := mappingTestRegistry{states: []mapping.BlockState{
		{Name: "minecraft:air"},
		{Name: "minecraft:stone", Properties: map[string]any{"kind": int32(1)}},
		{Name: "minecraft:new_block"},
	}}
	blocks, err := mapping.NewBlockMapper(nativeBlocks, []mapping.BlockState{
		{Name: "minecraft:air"},
		{Name: "minecraft:stone", Properties: map[string]any{"kind": int32(1)}},
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := mapping.NewItemMapper(
		[]protocol.ItemEntry{{Name: "minecraft:stone", RuntimeID: 2}, {Name: "minecraft:new_item", RuntimeID: 3}},
		map[string]mapping.TargetItem{"minecraft:stone": {RuntimeID: 7}},
	)
	if err != nil {
		t.Fatal(err)
	}
	return Protocol{runtime: &runtimeData{blocks: blocks, items: items}}
}

func TestGameplayItemMappingPreservesInput(t *testing.T) {
	p := testMappedProtocol(t)
	original := &packet.InventoryContent{Content: []protocol.ItemInstance{
		{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1}},
		{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 3}, Count: 1}},
	}}
	wantOriginal := &packet.InventoryContent{Content: []protocol.ItemInstance{
		{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1}},
		{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 3}, Count: 1}},
	}}
	converted := p.convertGameplayFromLatest(original, nil)
	if !reflect.DeepEqual(original, wantOriginal) {
		t.Fatalf("mapping mutated input: got %#v, want %#v", original, wantOriginal)
	}
	content := converted[0].(*packet.InventoryContent).Content
	if got := content[0].Stack.NetworkID; got != 7 {
		t.Fatalf("mapped item ID: got %d, want 7", got)
	}
	if got := content[1].Stack.NetworkID; got != 0 {
		t.Fatalf("unsupported clientbound item ID: got %d, want air", got)
	}
	upgraded := p.convertGameplayToLatest(&packet.InventorySlot{NewItem: content[0]}, nil)
	if got := upgraded[0].(*packet.InventorySlot).NewItem.Stack.NetworkID; got != 2 {
		t.Fatalf("reverse item ID: got %d, want 2", got)
	}
}

func TestGameplayBlockMappingFallbackAndRoundTrip(t *testing.T) {
	p := testMappedProtocol(t)
	stoneTarget, exact := p.runtime.blocks.NativeToTarget(1)
	if !exact {
		t.Fatal("stone mapping unexpectedly used fallback")
	}
	converted := p.convertGameplayFromLatest(&packet.UpdateBlock{NewBlockRuntimeID: 1}, nil)
	if got := converted[0].(*packet.UpdateBlock).NewBlockRuntimeID; got != stoneTarget {
		t.Fatalf("mapped block ID: got %d, want %d", got, stoneTarget)
	}
	upgraded := p.convertGameplayToLatest(converted[0], nil)
	if got := upgraded[0].(*packet.UpdateBlock).NewBlockRuntimeID; got != 1 {
		t.Fatalf("reverse block ID: got %d, want 1", got)
	}
	fallback := p.convertGameplayFromLatest(&packet.UpdateBlock{NewBlockRuntimeID: 2}, nil)
	if got := fallback[0].(*packet.UpdateBlock).NewBlockRuntimeID; got != p.runtime.blocks.TargetAir() {
		t.Fatalf("unsupported block fallback: got %d, want air", got)
	}
}

func TestPlayerAuthActionMappingPreservesInput(t *testing.T) {
	p := testMappedProtocol(t)
	input := &packet.PlayerAuthInput{ItemInteractionData: protocol.Option(protocol.UseItemTransactionData{
		HeldItem: protocol.ItemInstance{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, BlockRuntimeID: 1}},
		Actions: []protocol.InventoryAction{{
			OldItem: protocol.ItemInstance{Stack: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, BlockRuntimeID: 1}},
			NewItem: protocol.ItemInstance{},
		}},
		BlockRuntimeID: 1,
	})}
	want := *input
	wantInteraction, _ := input.ItemInteractionData.Value()

	converted := p.convertGameplayFromLatest(input, nil)
	if len(converted) != 1 {
		t.Fatalf("converted packet count: got %d, want 1", len(converted))
	}
	afterInput, _ := input.ItemInteractionData.Value()
	if !reflect.DeepEqual(*input, want) || !reflect.DeepEqual(afterInput, wantInteraction) {
		t.Fatal("PlayerAuthInput conversion mutated its input")
	}
	mapped := converted[0].(*packet.PlayerAuthInput)
	interaction, ok := mapped.ItemInteractionData.Value()
	if !ok || len(interaction.Actions) != 1 {
		t.Fatalf("mapped interaction: %#v", interaction)
	}
	if got, want := interaction.Actions[0].OldItem.Stack.NetworkID, int32(7); got != want {
		t.Fatalf("mapped action item ID: got %d, want %d", got, want)
	}
	stoneTarget, _ := p.runtime.blocks.NativeToTarget(1)
	if got := uint32(interaction.Actions[0].OldItem.Stack.BlockRuntimeID); got != stoneTarget {
		t.Fatalf("mapped action block ID: got %d, want %d", got, stoneTarget)
	}

	upgraded := p.convertGameplayToLatest(mapped, nil)
	reversed, _ := upgraded[0].(*packet.PlayerAuthInput).ItemInteractionData.Value()
	if got, want := reversed.Actions[0].OldItem.Stack.NetworkID, int32(2); got != want {
		t.Fatalf("reverse action item ID: got %d, want %d", got, want)
	}
}

func TestCraftingDataFiltersUnmappedOutputs(t *testing.T) {
	p := testMappedProtocol(t)
	input := &packet.CraftingData{ShapelessRecipes: []protocol.ShapelessRecipe{
		{Output: []protocol.ItemStack{{ItemType: protocol.ItemType{NetworkID: 2}}}},
		{Output: []protocol.ItemStack{{ItemType: protocol.ItemType{NetworkID: 3}}}},
	}}
	converted := mapCraftingData(input, p.runtime.items, p.runtime.blocks)
	if got, want := len(converted.ShapelessRecipes), 1; got != want {
		t.Fatalf("filtered recipe count: got %d, want %d", got, want)
	}
	if got := converted.ShapelessRecipes[0].Output[0].NetworkID; got != 7 {
		t.Fatalf("mapped recipe output: got %d, want 7", got)
	}
	if got := input.ShapelessRecipes[0].Output[0].NetworkID; got != 2 {
		t.Fatalf("recipe conversion mutated input: got %d, want 2", got)
	}
}

func TestCreativeContentPreservesNativeSelectionIDs(t *testing.T) {
	p := testMappedProtocol(t)
	stoneTarget, exact := p.runtime.blocks.NativeToTarget(1)
	if !exact {
		t.Fatal("stone mapping unexpectedly used fallback")
	}
	original := &packet.CreativeContent{
		Groups: []protocol.CreativeGroup{{
			Name: "test", Icon: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1, BlockRuntimeID: 1},
		}},
		Items: []protocol.CreativeItem{
			{CreativeItemNetworkID: 1341, Item: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1, BlockRuntimeID: 1}},
			{CreativeItemNetworkID: 1355, Item: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 3}, Count: 1}},
		},
	}
	wantOriginal := &packet.CreativeContent{
		Groups: []protocol.CreativeGroup{{
			Name: "test", Icon: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1, BlockRuntimeID: 1},
		}},
		Items: []protocol.CreativeItem{
			{CreativeItemNetworkID: 1341, Item: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 2}, Count: 1, BlockRuntimeID: 1}},
			{CreativeItemNetworkID: 1355, Item: protocol.ItemStack{ItemType: protocol.ItemType{NetworkID: 3}, Count: 1}},
		},
	}
	converted := p.convertGameplayFromLatest(original, nil)
	if !reflect.DeepEqual(original, wantOriginal) {
		t.Fatalf("creative conversion mutated input: got %#v, want %#v", original, wantOriginal)
	}
	creative := converted[0].(*packet.CreativeContent)
	if got, want := len(creative.Items), 1; got != want {
		t.Fatalf("filtered creative item count: got %d, want %d", got, want)
	}
	item := creative.Items[0]
	if got, want := item.CreativeItemNetworkID, uint32(1341); got != want {
		t.Fatalf("creative selection ID: got %d, want native ID %d", got, want)
	}
	if got, want := item.Item.NetworkID, int32(7); got != want {
		t.Fatalf("creative item network ID: got %d, want target ID %d", got, want)
	}
	if got, want := item.Item.BlockRuntimeID, int32(stoneTarget); got != want {
		t.Fatalf("creative block runtime ID: got %d, want target ID %d", got, want)
	}
	if got, want := creative.Groups[0].Icon.NetworkID, int32(7); got != want {
		t.Fatalf("creative group icon network ID: got %d, want target ID %d", got, want)
	}

	request := &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{Actions: []protocol.StackRequestAction{
		&protocol.CraftCreativeStackRequestAction{CreativeItemNetworkID: 1341},
	}}}}
	upgraded := p.convertGameplayToLatest(request, nil)
	action := upgraded[0].(*packet.ItemStackRequest).Requests[0].Actions[0].(*protocol.CraftCreativeStackRequestAction)
	if got, want := action.CreativeItemNetworkID, uint32(1341); got != want {
		t.Fatalf("serverbound creative selection ID: got %d, want %d", got, want)
	}
}
