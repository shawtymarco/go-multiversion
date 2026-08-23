package mapping

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// StateKey returns a deterministic key for a block identifier and all of its
// typed state properties. Property names and NBT scalar types are significant.
func StateKey(name string, properties map[string]any) (string, error) {
	keys := make([]string, 0, len(properties))
	for key := range properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	data := make([]byte, 0, len(name)+len(keys)*12)
	data = appendString(data, name)
	for _, key := range keys {
		data = appendString(data, key)
		switch value := properties[key].(type) {
		case bool:
			// Dragonfly block implementations expose bit states as bool while
			// the BDS registry stores the same NBT tag as uint8. Both encode as
			// one network byte and are intentionally equivalent here.
			data = append(data, 'B')
			if value {
				data = append(data, 1)
			} else {
				data = append(data, 0)
			}
		case uint8:
			data = append(data, 'B', value)
		case int8:
			data = append(data, 'B', byte(value))
		case int16:
			data = append(data, 's')
			data = binary.LittleEndian.AppendUint16(data, uint16(value))
		case uint16:
			data = append(data, 'S')
			data = binary.LittleEndian.AppendUint16(data, value)
		case int32:
			data = append(data, 'i')
			data = binary.LittleEndian.AppendUint32(data, uint32(value))
		case uint32:
			data = append(data, 'I')
			data = binary.LittleEndian.AppendUint32(data, value)
		case string:
			data = append(data, 't')
			data = appendString(data, value)
		default:
			return "", fmt.Errorf("block state %s property %s has unsupported type %T", name, key, value)
		}
	}
	return string(data), nil
}

func appendString(data []byte, value string) []byte {
	data = binary.LittleEndian.AppendUint32(data, uint32(len(value)))
	return append(data, value...)
}
