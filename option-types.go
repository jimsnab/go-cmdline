package cmdline

import (
	"fmt"
	"path/filepath"
	"strconv"
)

type OptionTypes interface {
	StringToAttributes(typeName string, spec string) *OptionTypeAttributes
	MakeValue(typeIndex int, inputValue string) (interface{}, error)
	NewList(typeIndex int) (interface{}, error)
	AppendList(typeIndex int, list interface{}, inputValue string) (interface{}, error)
}

type OptionTypeAttributes struct {
	Index        int
	DefaultValue interface{}
}

type argType int

const (
	argTypeBool argType = iota
	argTypeInt
	argTypeFloat64
	argTypeString
	argTypePath
)

type defaultOptionTypes struct {
}

func newDefaultOptionTypes() *defaultOptionTypes {
	return &defaultOptionTypes{}
}

func (dot *defaultOptionTypes) StringToAttributes(typeName string, spec string) *OptionTypeAttributes {
	switch typeName {
	case "bool":
		return &OptionTypeAttributes{Index: int(argTypeBool), DefaultValue: bool(false)}
	case "int":
		return &OptionTypeAttributes{Index: int(argTypeInt), DefaultValue: int(0)}
	case "float64":
		return &OptionTypeAttributes{Index: int(argTypeFloat64), DefaultValue: float64(0)}
	case "string":
		return &OptionTypeAttributes{Index: int(argTypeString), DefaultValue: ""}
	case "path":
		return &OptionTypeAttributes{Index: int(argTypePath), DefaultValue: ""}
	default:
		panic(fmt.Errorf("%svalid arg type %s in %s", basePanic, typeName, spec))
	}
}

func (dot *defaultOptionTypes) MakeValue(typeIndex int, inputValue string) (interface{}, error) {
	var result interface{}
	var err error

	switch argType(typeIndex) {
	case argTypeBool:
		result, err = strconv.ParseBool(inputValue)

	case argTypeInt:
		result, err = strconv.Atoi(inputValue)

	case argTypeFloat64:
		result, err = strconv.ParseFloat(inputValue, 64)

	case argTypeString:
		result = inputValue
		err = nil

	case argTypePath:
		result, err = filepath.Abs(inputValue)

	default:
		panic(fmt.Errorf("invalid arg type index"))
	}

	return result, err
}

func (dot *defaultOptionTypes) NewList(typeIndex int) (interface{}, error) {
	switch argType(typeIndex) {
	case argTypeBool:
		return []bool{}, nil

	case argTypeInt:
		return []int{}, nil

	case argTypeFloat64:
		return []float64{}, nil

	case argTypeString:
		return []string{}, nil

	case argTypePath:
		return []string{}, nil

	default:
		panic(fmt.Errorf("invalid arg type index"))
	}
}

func (dot *defaultOptionTypes) AppendList(typeIndex int, list interface{}, inputValue string) (interface{}, error) {
	value, err := dot.MakeValue(typeIndex, inputValue)
	if err != nil {
		return nil, err
	}

	switch argType(typeIndex) {
	case argTypeBool:
		list = append(list.([]bool), value.(bool))

	case argTypeInt:
		list = append(list.([]int), value.(int))

	case argTypeFloat64:
		list = append(list.([]float64), value.(float64))

	case argTypeString:
		list = append(list.([]string), value.(string))

	case argTypePath:
		list = append(list.([]string), value.(string))
	}

	return list, nil
}
