package v1_16_100

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type wireIO struct {
	protocol.IO
	reading bool
	runtime *runtimeData
}

type wireReader struct{ *wireIO }
type wireWriter struct{ *wireIO }

func newWireIO(io protocol.IO, reading bool, runtime *runtimeData) *wireIO {
	return &wireIO{IO: io, reading: reading, runtime: runtime}
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
	reading := false
	if inherited, ok := io.(interface{ LegacyWireReading() bool }); ok {
		reading = inherited.LegacyWireReading()
	}
	return newWireIO(io, reading, nil)
}

func (io *wireIO) LegacyWireReading() bool { return io.reading }

// UBlockPos preserves the unsigned Y coordinate used by protocol 419 before
// BlockPos switched to a signed Varint32 Y value.
func (io *wireIO) UBlockPos(x *protocol.BlockPos) {
	io.Varint32(&x[0])
	y := uint32(x[1])
	io.Varuint32(&y)
	x[1] = int32(y)
	io.Varint32(&x[2])
}

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

func (io *wireIO) Item(x *protocol.ItemStack)                    { marshalItem(io, x) }
func (io *wireIO) ItemInstance(x *protocol.ItemInstance)         { marshalItemInstance(io, x) }
func (io *wireIO) StackRequestItem(x *protocol.StackRequestItem) { marshalStackRequestItem(io, x) }
func (io *wireIO) ItemDescriptorCount(x *protocol.ItemDescriptorCount) {
	marshalItemDescriptorCount(io, x)
}
func (io *wireIO) StackRequestAction(x *protocol.StackRequestAction) {
	marshalStackRequestAction(io, x)
}
func (io *wireIO) MaterialReducer(x *protocol.MaterialReducer) { marshalMaterialReducer(io, x) }
func (io *wireIO) TransactionDataType(x *protocol.InventoryTransactionData) {
	marshalTransactionDataType(io, x)
}
func (io *wireIO) PlayerInventoryAction(x *protocol.UseItemTransactionData) {
	marshalPlayerInventoryAction(io, x)
}
func (io *wireIO) EntityMetadata(x *protocol.EntityMetadata) { marshalEntityMetadata(io, x) }
