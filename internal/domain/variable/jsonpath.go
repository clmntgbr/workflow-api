package variable

import (
	"strings"
	"unicode"
)

// needsBracketNotation reports whether a JSON object key must be accessed with
// bracket notation in PaesslerAG/jsonpath (e.g. hydra:member, @context).
func needsBracketNotation(key string) bool {
	if key == "" {
		return true
	}
	for i, r := range key {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return true
			}
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return true
		}
	}
	return false
}

func appendPathSegment(currentPath, key string) string {
	if needsBracketNotation(key) {
		return currentPath + `["` + escapeJSONPathKey(key) + `"]`
	}
	if currentPath == "$" {
		return "$." + key
	}
	return currentPath + "." + key
}

func escapeJSONPathKey(key string) string {
	return strings.ReplaceAll(key, `"`, `\"`)
}

// NormalizeJSONPath converts dot-notation paths with special keys into bracket
// notation understood by PaesslerAG/jsonpath.
//
// Example: $.hydra:member[0].id -> $["hydra:member"][0].id
func NormalizeJSONPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "$") {
		return path
	}
	if strings.Contains(path, `["`) {
		return path
	}

	var b strings.Builder
	b.WriteByte('$')
	i := 1
	for i < len(path) {
		switch path[i] {
		case '.':
			i++
			if i >= len(path) {
				break
			}
			start := i
			for i < len(path) && path[i] != '.' && path[i] != '[' {
				i++
			}
			key := path[start:i]
			if needsBracketNotation(key) {
				b.WriteString(`["`)
				b.WriteString(escapeJSONPathKey(key))
				b.WriteString(`"]`)
			} else {
				b.WriteByte('.')
				b.WriteString(key)
			}
		case '[':
			end := strings.IndexByte(path[i:], ']')
			if end < 0 {
				b.WriteString(path[i:])
				return b.String()
			}
			b.WriteString(path[i : i+end+1])
			i += end + 1
		default:
			b.WriteByte(path[i])
			i++
		}
	}
	return b.String()
}
