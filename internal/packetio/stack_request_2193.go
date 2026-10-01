package packetio

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

const (
	StackRequestActionCraftNonImplemented2193 uint8 = 18
	StackRequestActionCraftResults2193        uint8 = 19
)

var actions2193 = []func() protocol.StackRequestAction{
	func() protocol.StackRequestAction { return &protocol.TakeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.PlaceStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.SwapStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.DropStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.DestroyStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.ConsumeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CreateStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.LabTableCombineStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.BeaconPaymentStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.MineBlockStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftRecipeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.AutoCraftRecipeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftCreativeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftRecipeOptionalStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftGrindstoneRecipeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftLoomRecipeStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftNonImplementedStackRequestAction{} },
	func() protocol.StackRequestAction { return &protocol.CraftResultsDeprecatedStackRequestAction{} },
}

func StackRequestAction2193(io protocol.IO, value *protocol.StackRequestAction, reading bool) {
	var variant uint32
	var legacy byte
	if reading {
		io.Varuint32(&variant)
		io.Uint8(&legacy)
		if variant >= uint32(len(actions2193)) {
			io.UnknownEnumOption(variant, "2193 stack request variant")
			return
		}
		*value = actions2193[variant]()
	} else {
		switch (*value).(type) {
		case *protocol.TakeStackRequestAction:
			variant = 0
		case *protocol.PlaceStackRequestAction:
			variant = 1
		case *protocol.SwapStackRequestAction:
			variant = 2
		case *protocol.DropStackRequestAction:
			variant = 3
		case *protocol.DestroyStackRequestAction:
			variant = 4
		case *protocol.ConsumeStackRequestAction:
			variant = 5
		case *protocol.CreateStackRequestAction:
			variant = 6
		case *protocol.LabTableCombineStackRequestAction:
			variant = 7
		case *protocol.BeaconPaymentStackRequestAction:
			variant = 8
		case *protocol.MineBlockStackRequestAction:
			variant = 9
		case *protocol.CraftRecipeStackRequestAction:
			variant = 10
		case *protocol.AutoCraftRecipeStackRequestAction:
			variant = 11
		case *protocol.CraftCreativeStackRequestAction:
			variant = 12
		case *protocol.CraftRecipeOptionalStackRequestAction:
			variant = 13
		case *protocol.CraftGrindstoneRecipeStackRequestAction:
			variant = 14
		case *protocol.CraftLoomRecipeStackRequestAction:
			variant = 15
		case *protocol.CraftNonImplementedStackRequestAction:
			variant = 16
		case *protocol.CraftResultsDeprecatedStackRequestAction:
			variant = 17
		default:
			io.UnknownEnumOption(fmt.Sprintf("%T", *value), "2193 stack request action")
			return
		}
		legacy = byte(variant)
		if variant >= 7 {
			legacy += 2
		}
		io.Varuint32(&variant)
		io.Uint8(&legacy)
	}
	(*value).Marshal(io)
}
