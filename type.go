package gosmi

/*
#cgo LDFLAGS: -lsmi
#include <stdlib.h>
#include <smi.h>
*/
import "C"

import (
	"encoding/binary"

	"github.com/sleepinggenius2/gosmi/types"
)

type Enum struct {
	BaseType types.BaseType
	Values   []NamedNumber
}

type NamedNumber struct {
	Name  string
	Value interface{}
}

type Range struct {
	BaseType types.BaseType
	MinValue interface{}
	MaxValue interface{}
}

type Type struct {
	SmiType     *C.struct_SmiType `json:"-"`
	BaseType    types.BaseType
	Decl        types.Decl
	Description string
	Enum        *Enum
	Format      string
	Name        string
	Ranges      []Range
	Reference   string
	Status      types.Status
	Units       string
}

func (t *Type) getEnum() {
	if t.BaseType == types.BaseTypeUnknown || !(t.BaseType == types.BaseTypeEnum || t.BaseType == types.BaseTypeBits) {
		return
	}

	smiNamedNumber := C.smiGetFirstNamedNumber(t.SmiType)
	if smiNamedNumber == nil {
		return
	}

	enum := Enum{
		BaseType: types.BaseType(smiNamedNumber.value.basetype),
	}
	for ; smiNamedNumber != nil; smiNamedNumber = C.smiGetNextNamedNumber(smiNamedNumber) {
		namedNumber := NamedNumber{
			Name:  C.GoString(smiNamedNumber.name),
			Value: convertValue(smiNamedNumber.value),
		}
		enum.Values = append(enum.Values, namedNumber)
	}
	t.Enum = &enum
	return
}

func (t Type) GetModule() (module Module) {
	smiModule := C.smiGetTypeModule(t.SmiType)
	return CreateModule(smiModule)
}

func (t *Type) getRanges() {
	if t.BaseType == types.BaseTypeUnknown {
		return
	}

	ranges := make([]Range, 0)
	for smiRange := C.smiGetFirstRange(t.SmiType); smiRange != nil; smiRange = C.smiGetNextRange(smiRange) {
		r := Range{
			BaseType: types.BaseType(smiRange.minValue.basetype),
			MinValue: convertValue(smiRange.minValue),
			MaxValue: convertValue(smiRange.maxValue),
		}
		ranges = append(ranges, r)
	}
	t.Ranges = ranges
}

func CreateType(smiType *C.struct_SmiType) (outType Type) {
	if smiType == nil {
		return
	}

	outType.SmiType = smiType
	outType.BaseType = types.BaseType(smiType.basetype)

	if smiType.name == nil {
		smiType = C.smiGetParentType(smiType)
	}

	outType.Decl = types.Decl(smiType.decl)
	outType.Description = C.GoString(smiType.description)
	outType.Format = C.GoString(smiType.format)
	outType.Name = C.GoString(smiType.name)
	outType.Reference = C.GoString(smiType.reference)
	outType.Status = types.Status(smiType.status)
	outType.Units = C.GoString(smiType.units)

	outType.getEnum()
	outType.getRanges()

	return
}

func CreateTypeFromNode(smiNode *C.struct_SmiNode) (outType *Type) {
	smiType := C.smiGetNodeType(smiNode)

	if smiType == nil {
		return
	}

	tempType := CreateType(smiType)
	outType = &tempType

	if format := C.GoString(smiNode.format); format != "" {
		outType.Format = format
	}
	if units := C.GoString(smiNode.units); units != "" {
		outType.Units = units
	}

	return
}

func convertValue(value C.struct_SmiValue) (outValue interface{}) {
	switch types.BaseType(value.basetype) {
	case types.BaseTypeInteger32:
		tempValue := binary.LittleEndian.Uint32(value.value[:4])
		outValue = int32(tempValue)
	case types.BaseTypeUnsigned32:
		outValue = binary.LittleEndian.Uint32(value.value[:4])
	case types.BaseTypeInteger64:
		tempValue := binary.LittleEndian.Uint64(value.value[:8])
		outValue = int64(tempValue)
	case types.BaseTypeUnsigned64:
		outValue = binary.LittleEndian.Uint64(value.value[:8])
	}
	return
}