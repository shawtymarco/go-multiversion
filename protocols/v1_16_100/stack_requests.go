package v1_16_100

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
	if io.reading {
		request.FilterStrings = nil
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
	switch id {
	case 0:
		*value = &protocol.TakeStackRequestAction{}
	case 1:
		*value = &protocol.PlaceStackRequestAction{}
	case 2:
		*value = &protocol.SwapStackRequestAction{}
	case 3:
		*value = &protocol.DropStackRequestAction{}
	case 4:
		*value = &protocol.DestroyStackRequestAction{}
	case 5:
		*value = &protocol.ConsumeStackRequestAction{}
	case 6:
		*value = &protocol.CreateStackRequestAction{}
	case 7:
		*value = &protocol.LabTableCombineStackRequestAction{}
	case 8:
		*value = &protocol.BeaconPaymentStackRequestAction{}
	case 9:
		*value = &protocol.CraftRecipeStackRequestAction{}
	case 10:
		*value = &protocol.AutoCraftRecipeStackRequestAction{}
	case 11:
		*value = &protocol.CraftCreativeStackRequestAction{}
	case 12:
		*value = &protocol.CraftNonImplementedStackRequestAction{}
	case 13:
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
		*id = 6
	case *protocol.LabTableCombineStackRequestAction:
		*id = 7
	case *protocol.BeaconPaymentStackRequestAction:
		*id = 8
	case *protocol.CraftRecipeStackRequestAction:
		*id = 9
	case *protocol.AutoCraftRecipeStackRequestAction:
		*id = 10
	case *protocol.CraftCreativeStackRequestAction:
		*id = 11
	case *protocol.CraftNonImplementedStackRequestAction:
		*id = 12
	case *protocol.CraftResultsDeprecatedStackRequestAction:
		*id = 13
	default:
		return false
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
		if io.reading {
			action.NumberOfCrafts = 1
			action.Ingredients = nil
		}
	case *protocol.CraftCreativeStackRequestAction:
		io.Varuint32(&action.CreativeItemNetworkID)
		if io.reading {
			action.NumberOfCrafts = 1
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
				if legacy.reading {
					slot.CustomName = ""
					slot.FilteredCustomName = ""
					slot.DurabilityCorrection = 0
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
