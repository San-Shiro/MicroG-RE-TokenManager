// Package proto provides protobuf encoding/decoding without external dependencies.
// Adapted from gphotos-go's schema-driven encoder.
package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
)

// WireType represents protobuf wire types.
type WireType int

const (
	WireVarint  WireType = 0
	WireFixed64 WireType = 1
	WireBytes   WireType = 2
	WireFixed32 WireType = 5
)

// FieldType for schema-driven encoding.
type FieldType string

const (
	TypeInt     FieldType = "int"
	TypeString  FieldType = "string"
	TypeBytes   FieldType = "bytes"
	TypeMessage FieldType = "message"
	TypeBool    FieldType = "bool"
)

// FieldDef defines how a field should be encoded.
type FieldDef struct {
	Type       FieldType
	Repeated   bool
	MessageDef map[string]FieldDef
}

// MessageType is a schema for protobuf encoding.
type MessageType map[string]FieldDef

// Encode encodes a message map using a schema into protobuf bytes.
func Encode(message map[string]interface{}, schema MessageType) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeMessage(&buf, message, schema); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeRaw encodes without schema (auto-detect types).
func EncodeRaw(message map[int]interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeRawMessage(&buf, message); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeMessage(buf *bytes.Buffer, msg map[string]interface{}, schema MessageType) error {
	// Sort field numbers for deterministic output
	keys := make([]string, 0, len(msg))
	for k := range msg {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, _ := strconv.Atoi(keys[i])
		b, _ := strconv.Atoi(keys[j])
		return a < b
	})

	for _, key := range keys {
		value := msg[key]
		fieldNum, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("invalid field number: %s", key)
		}

		fieldDef, hasSchema := schema[key]
		if !hasSchema {
			// Auto-detect type
			if err := encodeAutoField(buf, fieldNum, value); err != nil {
				return err
			}
			continue
		}

		if fieldDef.Repeated {
			items, ok := value.([]interface{})
			if !ok {
				// Single item as repeated
				items = []interface{}{value}
			}
			for _, item := range items {
				if err := encodeTypedField(buf, fieldNum, item, fieldDef); err != nil {
					return err
				}
			}
		} else {
			if err := encodeTypedField(buf, fieldNum, value, fieldDef); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeTypedField(buf *bytes.Buffer, fieldNum int, value interface{}, def FieldDef) error {
	switch def.Type {
	case TypeInt:
		v, err := toInt64(value)
		if err != nil {
			return fmt.Errorf("field %d: %w", fieldNum, err)
		}
		writeTag(buf, fieldNum, WireVarint)
		writeVarint(buf, uint64(v))

	case TypeBool:
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("field %d: expected bool", fieldNum)
		}
		writeTag(buf, fieldNum, WireVarint)
		if v {
			writeVarint(buf, 1)
		} else {
			writeVarint(buf, 0)
		}

	case TypeString:
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %d: expected string, got %T", fieldNum, value)
		}
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(len(s)))
		buf.WriteString(s)

	case TypeBytes:
		var b []byte
		switch v := value.(type) {
		case []byte:
			b = v
		case string:
			b = []byte(v)
		default:
			return fmt.Errorf("field %d: expected bytes", fieldNum)
		}
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(len(b)))
		buf.Write(b)

	case TypeMessage:
		msgMap, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field %d: expected map[string]interface{} for message, got %T", fieldNum, value)
		}
		var nested bytes.Buffer
		schema := MessageType{}
		if def.MessageDef != nil {
			for k, v := range def.MessageDef {
				schema[k] = v
			}
		}
		if err := encodeMessage(&nested, msgMap, schema); err != nil {
			return fmt.Errorf("field %d: %w", fieldNum, err)
		}
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(nested.Len()))
		buf.Write(nested.Bytes())

	default:
		return fmt.Errorf("field %d: unknown type %s", fieldNum, def.Type)
	}
	return nil
}

func encodeAutoField(buf *bytes.Buffer, fieldNum int, value interface{}) error {
	switch v := value.(type) {
	case int:
		writeTag(buf, fieldNum, WireVarint)
		writeVarint(buf, uint64(v))
	case int32:
		writeTag(buf, fieldNum, WireVarint)
		writeVarint(buf, uint64(v))
	case int64:
		writeTag(buf, fieldNum, WireVarint)
		writeVarint(buf, uint64(v))
	case uint64:
		writeTag(buf, fieldNum, WireVarint)
		writeVarint(buf, v)
	case bool:
		writeTag(buf, fieldNum, WireVarint)
		if v {
			writeVarint(buf, 1)
		} else {
			writeVarint(buf, 0)
		}
	case string:
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(len(v)))
		buf.WriteString(v)
	case []byte:
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(len(v)))
		buf.Write(v)
	case map[string]interface{}:
		var nested bytes.Buffer
		if err := encodeMessage(&nested, v, MessageType{}); err != nil {
			return err
		}
		writeTag(buf, fieldNum, WireBytes)
		writeVarint(buf, uint64(nested.Len()))
		buf.Write(nested.Bytes())
	default:
		return fmt.Errorf("field %d: unsupported auto type %T", fieldNum, value)
	}
	return nil
}

func encodeRawMessage(buf *bytes.Buffer, msg map[int]interface{}) error {
	keys := make([]int, 0, len(msg))
	for k := range msg {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, fieldNum := range keys {
		if err := encodeAutoField(buf, fieldNum, msg[fieldNum]); err != nil {
			return err
		}
	}
	return nil
}

func writeTag(buf *bytes.Buffer, fieldNum int, wireType WireType) {
	writeVarint(buf, uint64(fieldNum<<3)|uint64(wireType))
}

func writeVarint(buf *bytes.Buffer, value uint64) {
	var varint [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(varint[:], value)
	buf.Write(varint[:n])
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint64:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}
