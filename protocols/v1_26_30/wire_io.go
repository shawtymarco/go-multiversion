package v1_26_30

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// wireIO keeps current primitive operations while replacing compound
// operations whose protocol-1001 layout differs from native.
type wireIO struct {
	protocol.IO
	reading bool
	runtime *runtimeData
}

type wireReader struct{ *wireIO }
type wireWriter struct{ *wireIO }

func newWireIO(io protocol.IO, reading bool, runtime ...*runtimeData) *wireIO {
	var data *runtimeData
	if len(runtime) != 0 {
		data = runtime[0]
	}
	return &wireIO{IO: io, reading: reading, runtime: data}
}

func asWireIO(io protocol.IO) *wireIO {
	switch legacy := io.(type) {
	case *wireIO:
		return legacy
	case *wireReader:
		return legacy.wireIO
	case *wireWriter:
		return legacy.wireIO
	}
	return newWireIO(io, false)
}

// directional preserves the concrete reader/writer marker required by the
// generic protocol slice helpers. Passing wireIO directly would hide
// SliceLength from readers and prevent those helpers from allocating values.
func (io *wireIO) directional() protocol.IO {
	if io.reading {
		return &wireReader{wireIO: io}
	}
	return &wireWriter{wireIO: io}
}

func (reader *wireReader) SliceLength(value uint32, max uint32) {
	if limited, ok := reader.IO.(interface{ SliceLength(uint32, uint32) }); ok {
		limited.SliceLength(value, max)
	}
}

func (io *wireIO) EntityMetadata(x *protocol.EntityMetadata) { marshalEntityMetadata(io, x) }
func (io *wireIO) Item(x *protocol.ItemStack)                { marshalItem(io, x) }
func (io *wireIO) ItemInstance(x *protocol.ItemInstance)     { marshalItemInstance(io, x) }
func (io *wireIO) StackRequestItem(x *protocol.StackRequestItem) {
	marshalStackRequestItem(io, x)
}
func (io *wireIO) ItemDescriptorCount(x *protocol.ItemDescriptorCount) {
	marshalItemDescriptorCount(io, x)
}
func (io *wireIO) StackRequestAction(x *protocol.StackRequestAction) {
	marshalStackRequestAction(io, x)
}
func (io *wireIO) MaterialReducer(x *protocol.MaterialReducer) {
	marshalMaterialReducer(io, x)
}
func (io *wireIO) TransactionDataType(x *protocol.InventoryTransactionData) {
	marshalTransactionDataType(io, x)
}
func (io *wireIO) PlayerInventoryAction(x *protocol.UseItemTransactionData) {
	marshalPlayerInventoryAction(io, x)
}
func (io *wireIO) GameRule(x *protocol.GameRule) { marshalGameRule(io, x, false) }
func (io *wireIO) AbilityValue(x *any)           { marshalAbilityValue(io, x) }
func (io *wireIO) ShapeData(x *protocol.ShapeData) {
	marshalShapeData(io, x)
}
