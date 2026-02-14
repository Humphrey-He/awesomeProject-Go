package interface_internals

import (
	"fmt"
	"reflect"
)

// EmptyInterfaceModel models runtime "eface" conceptually:
// type pointer + data pointer.
type EmptyInterfaceModel struct {
	Type string
	Data string
}

// NonEmptyInterfaceModel models runtime "iface" conceptually:
// itab pointer (type + method table) + data pointer.
type NonEmptyInterfaceModel struct {
	Interface string
	Concrete  string
	MethodSet []string
}

// Speaker is a non-empty interface used in demos.
type Speaker interface {
	Speak() string
}

type Person struct {
	Name string
}

func (p Person) Speak() string {
	return "hi, " + p.Name
}

// DescribeInterface reports dynamic type/value characteristics.
func DescribeInterface(v any) string {
	if v == nil {
		return "interface=nil"
	}
	rv := reflect.ValueOf(v)
	rt := reflect.TypeOf(v)
	return fmt.Sprintf("dynamic_type=%s, kind=%s, value=%v", rt.String(), rv.Kind(), v)
}

// NilPitfall shows: typed nil pointer assigned to interface is NOT nil interface.
func NilPitfall() (isNil bool, describe string) {
	var p *Person = nil
	var s Speaker = p
	return s == nil, DescribeInterface(s)
}

// TypeAssertDemo demonstrates safe assertion and type switch.
func TypeAssertDemo(v any) (string, bool) {
	if p, ok := v.(Person); ok {
		return "person:" + p.Name, true
	}

	switch x := v.(type) {
	case string:
		return "string:" + x, true
	case int:
		return fmt.Sprintf("int:%d", x), true
	default:
		return "unknown", false
	}
}

func BuildEmptyModel(v any) EmptyInterfaceModel {
	if v == nil {
		return EmptyInterfaceModel{Type: "<nil>", Data: "<nil>"}
	}
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	return EmptyInterfaceModel{
		Type: rt.String(),
		Data: fmt.Sprintf("%v", rv.Interface()),
	}
}

func BuildNonEmptyModel(s Speaker) NonEmptyInterfaceModel {
	if s == nil {
		return NonEmptyInterfaceModel{
			Interface: "Speaker",
			Concrete:  "<nil>",
			MethodSet: []string{"Speak() string"},
		}
	}
	rt := reflect.TypeOf(s)
	return NonEmptyInterfaceModel{
		Interface: "Speaker",
		Concrete:  rt.String(),
		MethodSet: []string{"Speak() string"},
	}
}
