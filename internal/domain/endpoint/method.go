package endpoint

import (
	"fmt"
	"strings"
)

type Method string

const (
	MethodGET     Method = "GET"
	MethodPOST    Method = "POST"
	MethodPUT     Method = "PUT"
	MethodPATCH   Method = "PATCH"
	MethodDELETE  Method = "DELETE"
	MethodHEAD    Method = "HEAD"
	MethodOPTIONS Method = "OPTIONS"
)

func (m Method) Valid() bool {
	switch m {
	case MethodGET, MethodPOST, MethodPUT, MethodPATCH, MethodDELETE, MethodHEAD, MethodOPTIONS:
		return true
	default:
		return false
	}
}

func ParseMethod(value string) (Method, error) {
	m := Method(strings.ToUpper(strings.TrimSpace(value)))
	if !m.Valid() {
		return "", fmt.Errorf("invalid endpoint method %q", value)
	}
	return m, nil
}
