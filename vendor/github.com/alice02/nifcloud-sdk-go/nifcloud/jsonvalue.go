package nifcloud

// JSONValue is a representation of a grab bag type that will be marshaled
// into a json string. This type can be used just like any other map.
//
//	Example:
//
//	values := nifcloud.JSONValue{
//		"Foo": "Bar",
//	}
//	values["Baz"] = "Qux"
type JSONValue map[string]interface{}
