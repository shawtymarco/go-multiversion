package v1_26_45

import "github.com/sandertv/gophertunnel/minecraft/protocol"

type wireIO struct {
	protocol.IO
	reading bool
	runtime *runtimeData
}

type wireReader struct{ *wireIO }
type wireWriter struct{ *wireIO }

func newWireIO(base protocol.IO, reading bool, runtime ...*runtimeData) *wireIO {
	var data *runtimeData
	if len(runtime) != 0 {
		data = runtime[0]
	}
	return &wireIO{IO: base, reading: reading, runtime: data}
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
	panic("protocol 2169 packet used without its reader or writer")
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

func (io *wireIO) PlayerInventoryAction(value *protocol.UseItemTransactionData) {
	marshalPlayerInventoryAction(io, value)
}

func (io *wireIO) StackRequestItem(value *protocol.StackRequestItem) {
	if io.runtime == nil || io.runtime.blocks == nil || value.BlockRuntimeID <= 0 {
		io.IO.StackRequestItem(value)
		return
	}
	if io.reading {
		io.IO.StackRequestItem(value)
		mapped, ok := io.runtime.blocks.TargetToNative(uint32(value.BlockRuntimeID))
		if !ok {
			io.InvalidValue(value.BlockRuntimeID, "stack request block runtime ID", "unknown protocol 2169 block")
			return
		}
		value.BlockRuntimeID = int32(mapped)
		return
	}
	cloned := *value
	mapped, _, _ := io.runtime.blocks.MapNative(uint32(value.BlockRuntimeID))
	cloned.BlockRuntimeID = int32(mapped)
	io.IO.StackRequestItem(&cloned)
}

// ShapeData preserves the protocol-2169 TextShape layout, which predates the
// optional line-gap field added to the protocol-2192 native model.
func (io *wireIO) ShapeData(value *protocol.ShapeData) {
	var kind uint32
	if io.reading {
		io.Varuint32(&kind)
		switch kind {
		case protocol.ShapeDataLast:
			*value = &protocol.LastShape{}
		case protocol.ShapeDataArrow:
			*value = &protocol.ArrowShape{}
		case protocol.ShapeDataText:
			*value = &protocol.TextShape{}
		case protocol.ShapeDataBox:
			*value = &protocol.BoxShape{}
		case protocol.ShapeDataLine:
			*value = &protocol.LineShape{}
		case protocol.ShapeDataSphere:
			*value = &protocol.SphereShape{}
		case protocol.ShapeDataCylinder:
			*value = &protocol.CylinderShape{}
		case protocol.ShapeDataPyramid:
			*value = &protocol.PyramidShape{}
		case protocol.ShapeDataEllipsoid:
			*value = &protocol.EllipsoidShape{}
		case protocol.ShapeDataCone:
			*value = &protocol.ConeShape{}
		default:
			io.UnknownEnumOption(kind, "debug shape data type")
			return
		}
	} else {
		switch (*value).(type) {
		case *protocol.LastShape:
			kind = protocol.ShapeDataLast
		case *protocol.ArrowShape:
			kind = protocol.ShapeDataArrow
		case *protocol.TextShape:
			kind = protocol.ShapeDataText
		case *protocol.BoxShape:
			kind = protocol.ShapeDataBox
		case *protocol.LineShape:
			kind = protocol.ShapeDataLine
		case *protocol.SphereShape:
			kind = protocol.ShapeDataSphere
		case *protocol.CylinderShape:
			kind = protocol.ShapeDataCylinder
		case *protocol.PyramidShape:
			kind = protocol.ShapeDataPyramid
		case *protocol.EllipsoidShape:
			kind = protocol.ShapeDataEllipsoid
		case *protocol.ConeShape:
			kind = protocol.ShapeDataCone
		default:
			io.UnknownEnumOption(*value, "debug shape data type")
			return
		}
		io.Varuint32(&kind)
	}
	if text, ok := (*value).(*protocol.TextShape); ok {
		marshalTextShape(io, text)
		return
	}
	(*value).Marshal(io)
}

func doubleOptionalFunc[T any](io *wireIO, value *protocol.Optional[T], marshal func(*T)) {
	present := true
	io.Bool(&present)
	if !present {
		*value = protocol.Optional[T]{}
		return
	}
	protocol.OptionalFunc(io, value, marshal)
}
