package multiversion_test

import (
	"bytes"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	multiversion "github.com/shawtymarco/go-multiversion"
	"testing"
)

func wireAdapters2223() []minecraft.Protocol {
	return []minecraft.Protocol{multiversion.V1_26_50(), multiversion.V1_26_45(), multiversion.V1_26_30(), multiversion.V1_26_20(), multiversion.V1_26_10(), multiversion.V1_26_0(), multiversion.V1_21_130(), multiversion.V1_21_110(), multiversion.V1_21_100(), multiversion.V1_21_50(), multiversion.V1_21_40(), multiversion.V1_18_10(), multiversion.V1_18_0(), multiversion.V1_16_100()}
}

func TestPreviewOnlyPacketExclusionAcrossHistoricalAdapters(t *testing.T) {
	if protocol.CurrentProtocol != 2223 {
		t.Fatal("wrong native protocol")
	}
	for _, adapter := range wireAdapters2223() {
		t.Run(adapter.Ver(), func(t *testing.T) {
			for _, listener := range []bool{true, false} {
				for id := range adapter.Packets(listener) {
					if id >= packet.IDClientboundMatchmakingState {
						t.Fatalf("preview ID %d in historical pool", id)
					}
				}
			}
			if packets := adapter.ConvertFromLatest(&packet.ClientboundMatchmakingState{}, nil); len(packets) != 0 {
				t.Fatal("preview matchmaking leaked into historical client")
			}
		})
	}
}

func TestDeprecatedStackActionsAcrossHistoricalAdapters(t *testing.T) {
	for _, adapter := range wireAdapters2223() {
		if adapter.ID() <= 486 {
			continue
		}
		t.Run(adapter.Ver(), func(t *testing.T) {
			request := &packet.ItemStackRequest{Requests: []protocol.ItemStackRequest{{RequestID: -1, Actions: []protocol.StackRequestAction{&protocol.CraftNonImplementedStackRequestAction{}, &protocol.CraftResultsDeprecatedStackRequestAction{TimesCrafted: 2}}}}}
			packets := adapter.ConvertFromLatest(request, nil)
			if len(packets) != 1 {
				t.Fatal("wire request unexpectedly dropped")
			}
			var payload bytes.Buffer
			packets[0].Marshal(adapter.NewWriter(&payload, 0))
			decoded := adapter.Packets(true)[packet.IDItemStackRequest]()
			decoded.Marshal(adapter.NewReader(&payload, 0, true))
			if payload.Len() != 0 {
				t.Fatalf("%d unread action bytes", payload.Len())
			}
			latest := adapter.ConvertToLatest(decoded, nil)
			if len(latest) != 1 {
				t.Fatal("decoded request unexpectedly dropped")
			}
			actions := latest[0].(*packet.ItemStackRequest).Requests[0].Actions
			if _, ok := actions[0].(*protocol.CraftNonImplementedStackRequestAction); !ok {
				t.Fatal("nonimplemented action changed into preview reserved action")
			}
			if _, ok := actions[1].(*protocol.CraftResultsDeprecatedStackRequestAction); !ok {
				t.Fatal("deprecated result action changed")
			}
		})
	}
}
