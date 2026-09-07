package event

import (
	"strings"

	"github.com/google/uuid"
)

func DeterministicID(parts ...string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(strings.Join(parts, ":"))).String()
}
