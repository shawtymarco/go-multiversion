package packetconv

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"testing"
)

func TestPreviewMemoryCategoriesUseVersionLocalIDs(t *testing.T) {
	if got := MemoryCategory(2193, protocol.MemoryCategoryBlockTickingQueues, true); got != 6 {
		t.Fatalf("native block queues mapped to %d, want historical 6", got)
	}
	if got := MemoryCategory(2193, 6, false); got != protocol.MemoryCategoryBlockTickingQueues {
		t.Fatalf("historical block queues mapped to %d", got)
	}
	if got := MemoryCategory(2193, 5, false); got != protocol.MemoryCategoryUnknown {
		t.Fatalf("removed Balancer category mapped to %d", got)
	}
	if got := MemoryCategory(2193, protocol.MemoryCategoryExecutable, true); got != 0 {
		t.Fatalf("preview-only Executable category mapped to %d", got)
	}
}
