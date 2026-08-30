package v1_18_0

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type placeInContainerAction struct {
	Count               byte
	Source, Destination protocol.StackRequestSlotInfo
}

func (a *placeInContainerAction) Marshal(io protocol.IO) {
	legacy := asWireIO(io)
	legacy.Uint8(&a.Count)
	marshalStackRequestSlotInfo(legacy, &a.Source)
	marshalStackRequestSlotInfo(legacy, &a.Destination)
}

type takeOutContainerAction placeInContainerAction

func (a *takeOutContainerAction) Marshal(io protocol.IO) {
	(*placeInContainerAction)(a).Marshal(io)
}

func marshalItemStackRequestPacket(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ItemStackRequest)
	protocol.FuncIOSlice(io.directional(), &pk.Requests, func(raw protocol.IO, request *protocol.ItemStackRequest) {
		marshalItemStackRequest(asWireIO(raw), request)
	})
}

func marshalItemStackRequest(io *wireIO, request *protocol.ItemStackRequest) {
	io.Varint32(&request.RequestID)
	protocol.FuncSlice(io.directional(), &request.Actions, func(action *protocol.StackRequestAction) {
		marshalStackRequestAction(io, action)
	})
	protocol.FuncSlice(io.directional(), &request.FilterStrings, io.String)
	if io.reading {
		request.FilterCause = 0
	}
}

func marshalStackRequestAction(io *wireIO, value *protocol.StackRequestAction) {
	var actionID uint8
	if io.reading {
		io.Uint8(&actionID)
		if !newStackRequestAction(actionID, value) {
			io.UnknownEnumOption(actionID, "stack request action type")
			return
		}
	} else {
		if !stackRequestActionID(*value, &actionID) {
			io.UnknownEnumOption(fmt.Sprintf("%T", *value), "stack request action type")
			return
		}
		io.Uint8(&actionID)
	}
	marshalStackRequestActionBody(io, *value)
}

func newStackRequestAction(id uint8, value *protocol.StackRequestAction) bool {
	if id >= 7 {
		id += 2
	}
	switch id {
	case protocol.StackRequestActionTake:
		*value = &protocol.TakeStackRequestAction{}
	case protocol.StackRequestActionPlace:
		*value = &protocol.PlaceStackRequestAction{}
	case protocol.StackRequestActionSwap:
		*value = &protocol.SwapStackRequestAction{}
	case protocol.StackRequestActionDrop:
		*value = &protocol.DropStackRequestAction{}
	case protocol.StackRequestActionDestroy:
		*value = &protocol.DestroyStackRequestAction{}
	case protocol.StackRequestActionConsume:
		*value = &protocol.ConsumeStackRequestAction{}
	case protocol.StackRequestActionCreate:
		*value = &protocol.CreateStackRequestAction{}
	case protocol.StackRequestActionLabTableCombine:
		*value = &protocol.LabTableCombineStackRequestAction{}
	case protocol.StackRequestActionBeaconPayment:
		*value = &protocol.BeaconPaymentStackRequestAction{}
	case protocol.StackRequestActionMineBlock:
		*value = &protocol.MineBlockStackRequestAction{}
	case protocol.StackRequestActionCraftRecipe:
		*value = &protocol.CraftRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeAuto:
		*value = &protocol.AutoCraftRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftCreative:
		*value = &protocol.CraftCreativeStackRequestAction{}
	case protocol.StackRequestActionCraftRecipeOptional:
		*value = &protocol.CraftRecipeOptionalStackRequestAction{}
	case protocol.StackRequestActionCraftGrindstone:
		*value = &protocol.CraftGrindstoneRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftLoom:
		*value = &protocol.CraftLoomRecipeStackRequestAction{}
	case protocol.StackRequestActionCraftNonImplementedDeprecated:
		*value = &protocol.CraftNonImplementedStackRequestAction{}
	case protocol.StackRequestActionCraftResultsDeprecated:
		*value = &protocol.CraftResultsDeprecatedStackRequestAction{}
	default:
		return false
	}
	return true
}

func stackRequestActionID(value protocol.StackRequestAction, id *uint8) bool {
	switch value.(type) {
	case *protocol.TakeStackRequestAction:
		*id = protocol.StackRequestActionTake
	case *protocol.PlaceStackRequestAction:
		*id = protocol.StackRequestActionPlace
	case *protocol.SwapStackRequestAction:
		*id = protocol.StackRequestActionSwap
	case *protocol.DropStackRequestAction:
		*id = protocol.StackRequestActionDrop
	case *protocol.DestroyStackRequestAction:
		*id = protocol.StackRequestActionDestroy
	case *protocol.ConsumeStackRequestAction:
		*id = protocol.StackRequestActionConsume
	case *protocol.CreateStackRequestAction:
		*id = protocol.StackRequestActionCreate
	case *protocol.LabTableCombineStackRequestAction:
		*id = protocol.StackRequestActionLabTableCombine
	case *protocol.BeaconPaymentStackRequestAction:
		*id = protocol.StackRequestActionBeaconPayment
	case *protocol.MineBlockStackRequestAction:
		*id = protocol.StackRequestActionMineBlock
	case *protocol.CraftRecipeStackRequestAction:
		*id = protocol.StackRequestActionCraftRecipe
	case *protocol.AutoCraftRecipeStackRequestAction:
		*id = protocol.StackRequestActionCraftRecipeAuto
	case *protocol.CraftCreativeStackRequestAction:
		*id = protocol.StackRequestActionCraftCreative
	case *protocol.CraftRecipeOptionalStackRequestAction:
		*id = protocol.StackRequestActionCraftRecipeOptional
	case *protocol.CraftGrindstoneRecipeStackRequestAction:
		*id = protocol.StackRequestActionCraftGrindstone
	case *protocol.CraftLoomRecipeStackRequestAction:
		*id = protocol.StackRequestActionCraftLoom
	case *protocol.CraftNonImplementedStackRequestAction:
		*id = protocol.StackRequestActionCraftNonImplementedDeprecated
	case *protocol.CraftResultsDeprecatedStackRequestAction:
		*id = protocol.StackRequestActionCraftResultsDeprecated
	default:
		return false
	}
	if *id >= protocol.StackRequestActionLabTableCombine {
		*id -= 2
	}
	return true
}

func marshalStackRequestActionBody(io *wireIO, value protocol.StackRequestAction) {
	switch action := value.(type) {
	case *protocol.TakeStackRequestAction:
		io.Uint8(&action.Count)
		marshalStackRequestSlotInfo(io, &action.Source)
		marshalStackRequestSlotInfo(io, &action.Destination)
	case *protocol.PlaceStackRequestAction:
		io.Uint8(&action.Count)
		marshalStackRequestSlotInfo(io, &action.Source)
		marshalStackRequestSlotInfo(io, &action.Destination)
	case *protocol.SwapStackRequestAction:
		marshalStackRequestSlotInfo(io, &action.Source)
		marshalStackRequestSlotInfo(io, &action.Destination)
	case *protocol.DropStackRequestAction:
		io.Uint8(&action.Count)
		marshalStackRequestSlotInfo(io, &action.Source)
		io.Bool(&action.Randomly)
	case *protocol.DestroyStackRequestAction:
		io.Uint8(&action.Count)
		marshalStackRequestSlotInfo(io, &action.Source)
	case *protocol.ConsumeStackRequestAction:
		io.Uint8(&action.Count)
		marshalStackRequestSlotInfo(io, &action.Source)
	case *protocol.CreateStackRequestAction:
		io.Uint8(&action.ResultsSlot)
	case *protocol.LabTableCombineStackRequestAction, *protocol.CraftNonImplementedStackRequestAction:
	case *protocol.BeaconPaymentStackRequestAction:
		io.Varint32(&action.PrimaryEffect)
		io.Varint32(&action.SecondaryEffect)
	case *protocol.MineBlockStackRequestAction:
		io.Varint32(&action.HotbarSlot)
		io.Varint32(&action.PredictedDurability)
		io.Varint32(&action.StackNetworkID)
	case *protocol.CraftRecipeStackRequestAction:
		io.Varuint32(&action.RecipeNetworkID)
		if io.reading {
			action.NumberOfCrafts = 1
		}
	case *protocol.AutoCraftRecipeStackRequestAction:
		io.Varuint32(&action.RecipeNetworkID)
		io.Uint8(&action.NumberOfCrafts)
		if io.reading {
			action.Ingredients = nil
		}
	case *protocol.CraftCreativeStackRequestAction:
		io.Varuint32(&action.CreativeItemNetworkID)
		if io.reading {
			action.NumberOfCrafts = 1
		}
	case *protocol.CraftRecipeOptionalStackRequestAction:
		protocol.IntegerFunc(&action.RecipeNetworkID, io.Varint32)
		io.Int32(&action.FilterStringIndex)
	case *protocol.CraftGrindstoneRecipeStackRequestAction:
		io.Varuint32(&action.RecipeNetworkID)
		io.Varint32(&action.Cost)
		if io.reading {
			action.NumberOfCrafts = 1
		}
	case *protocol.CraftLoomRecipeStackRequestAction:
		io.String(&action.Pattern)
		if io.reading {
			action.TimesCrafted = 1
		}
	case *protocol.CraftResultsDeprecatedStackRequestAction:
		protocol.FuncSlice(io.directional(), &action.ResultItems, io.StackRequestItem)
		io.Uint8(&action.TimesCrafted)
	default:
		io.UnknownEnumOption(fmt.Sprintf("%T", value), "stack request action type")
	}
}

func marshalStackRequestSlotInfo(io *wireIO, value *protocol.StackRequestSlotInfo) {
	containerID := toLegacyContainerID(value.Container.ContainerID)
	io.Uint8(&containerID)
	value.Container.ContainerID = toNativeContainerID(containerID)
	if io.reading {
		value.Container.DynamicContainerID = protocol.Optional[uint32]{}
	}
	io.Uint8(&value.Slot)
	io.Varint32(&value.StackNetworkID)
}

func marshalItemStackResponsePacket(io *wireIO, raw packet.Packet) {
	pk := raw.(*packet.ItemStackResponse)
	protocol.FuncIOSlice(io.directional(), &pk.Responses, func(io protocol.IO, response *protocol.ItemStackResponse) {
		marshalItemStackResponse(asWireIO(io), response)
	})
}

func marshalItemStackResponse(io *wireIO, response *protocol.ItemStackResponse) {
	io.Uint8(&response.Status)
	io.Varint32(&response.RequestID)
	if response.Status == protocol.ItemStackResponseStatusOK {
		protocol.FuncIOSlice(io.directional(), &response.ContainerInfo, func(io protocol.IO, info *protocol.StackResponseContainerInfo) {
			legacy := asWireIO(io)
			containerID := toLegacyContainerID(info.Container.ContainerID)
			legacy.Uint8(&containerID)
			info.Container.ContainerID = toNativeContainerID(containerID)
			if legacy.reading {
				info.Container.DynamicContainerID = protocol.Optional[uint32]{}
			}
			protocol.FuncIOSlice(io, &info.SlotInfo, func(io protocol.IO, slot *protocol.StackResponseSlotInfo) {
				legacy := asWireIO(io)
				legacy.Uint8(&slot.Slot)
				legacy.Uint8(&slot.HotbarSlot)
				legacy.Uint8(&slot.Count)
				legacy.Varint32(&slot.StackNetworkID)
				if slot.Slot != slot.HotbarSlot {
					legacy.InvalidValue(slot.HotbarSlot, "hotbar slot", "must equal the normal slot")
				}
				legacy.String(&slot.CustomName)
				legacy.Varint32(&slot.DurabilityCorrection)
				if legacy.reading {
					slot.FilteredCustomName = ""
				}
			})
		})
	} else if io.reading {
		response.ContainerInfo = nil
	}
}

func toLegacyContainerID(id byte) byte {
	if id > protocol.ContainerRecipeBook {
		return id - 1
	}
	return id
}

func toNativeContainerID(id byte) byte {
	if id >= protocol.ContainerRecipeBook {
		return id + 1
	}
	return id
}
