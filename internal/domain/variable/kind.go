package variable

import "fmt"

type Kind string

const (
	KindExtracted Kind = "extracted"
	KindStatic    Kind = "static"
)

func (k Kind) Valid() bool {
	switch k {
	case KindExtracted, KindStatic:
		return true
	default:
		return false
	}
}

func ParseKind(value string) (Kind, error) {
	k := Kind(value)
	if !k.Valid() {
		return "", fmt.Errorf("invalid variable kind %q", value)
	}
	return k, nil
}
