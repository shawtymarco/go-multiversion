package proxy

import (
	"reflect"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// cloneValue preserves typed NBT and copies mutable packet contents. Optional's
// private payload is handled through its public API for mutable packet types.
func cloneValue(v reflect.Value) reflect.Value {
	if !v.IsValid() {
		return v
	}
	if v.CanInterface() {
		switch x := v.Interface().(type) {
		case protocol.InputFlags:
			y := protocol.NewInputFlags(x.Len())
			for i := 0; i < x.Len(); i++ {
				if x.Load(i) {
					y.Set(i)
				}
			}
			return reflect.ValueOf(y)
		case protocol.Optional[[]byte]:
			return reflect.ValueOf(cloneOption(x))
		case protocol.Optional[[]protocol.PlayerBlockAction]:
			return reflect.ValueOf(cloneOption(x))
		case protocol.Optional[protocol.UseItemTransactionData]:
			return reflect.ValueOf(cloneOption(x))
		case protocol.Optional[protocol.ItemStackRequest]:
			return reflect.ValueOf(cloneOption(x))
		case protocol.Optional[protocol.ItemInstance]:
			return reflect.ValueOf(cloneOption(x))
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type().Elem())
		out.Elem().Set(cloneValue(v.Elem()))
		return out
	case reflect.Interface:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.New(v.Type()).Elem()
		out.Set(cloneValue(v.Elem()))
		return out
	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		switch v.Type().Elem().Kind() {
		case reflect.Uint8, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.String:
			reflect.Copy(out, v)
			return out
		}
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(cloneValue(v.Index(i)))
		}
		return out
	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		out := reflect.MakeMapWithSize(v.Type(), v.Len())
		it := v.MapRange()
		for it.Next() {
			out.SetMapIndex(it.Key(), cloneValue(it.Value()))
		}
		return out
	case reflect.Struct:
		out := reflect.New(v.Type()).Elem()
		out.Set(v)
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				out.Field(i).Set(cloneValue(v.Field(i)))
			}
		}
		return out
	case reflect.Array:
		out := reflect.New(v.Type()).Elem()
		for i := 0; i < v.Len(); i++ {
			out.Index(i).Set(cloneValue(v.Index(i)))
		}
		return out
	default:
		return v
	}
}
func cloneOption[T any](v protocol.Optional[T]) protocol.Optional[T] {
	x, ok := v.Value()
	if !ok {
		return protocol.Optional[T]{}
	}
	return protocol.Option(cloneValue(reflect.ValueOf(x)).Interface().(T))
}
func clonePacket(pk packet.Packet) packet.Packet {
	return cloneValue(reflect.ValueOf(pk)).Interface().(packet.Packet)
}

// visitPacket traverses only exported wire data and mutable Optional payloads.
func visitPacket(value reflect.Value, visit func(reflect.Value, string) error, name string) error {
	if !value.IsValid() {
		return nil
	}
	if value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		return visitPacket(value.Elem(), visit, name)
	}
	if err := visit(value, name); err != nil {
		return err
	}
	if value.CanInterface() && value.CanSet() {
		switch x := value.Interface().(type) {
		case protocol.Optional[protocol.UseItemTransactionData]:
			n, err := visitOption(x, visit)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(n))
			return nil
		case protocol.Optional[protocol.ItemStackRequest]:
			n, err := visitOption(x, visit)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(n))
			return nil
		case protocol.Optional[protocol.ItemInstance]:
			n, err := visitOption(x, visit)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(n))
			return nil
		}
	}
	switch value.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).IsExported() {
				if value.Type() == reflect.TypeOf(protocol.ItemStack{}) && value.Type().Field(i).Name == "BlockRuntimeID" && value.Field(i).Int() == 0 {
					continue
				}
				if err := visitPacket(value.Field(i), visit, value.Type().Field(i).Name); err != nil {
					return err
				}
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if err := visitPacket(value.Index(i), visit, name); err != nil {
				return err
			}
		}
	}
	return nil
}
func visitOption[T any](o protocol.Optional[T], visit func(reflect.Value, string) error) (protocol.Optional[T], error) {
	x, ok := o.Value()
	if !ok {
		return o, nil
	}
	if err := visitPacket(reflect.ValueOf(&x).Elem(), visit, ""); err != nil {
		return o, err
	}
	return protocol.Option(x), nil
}
