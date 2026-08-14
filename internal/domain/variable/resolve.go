package variable

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var placeholderPattern = regexp.MustCompile(`\{\{([a-zA-Z0-9_-]+)\}\}`)

type MissingVariableError struct {
	Key string
}

func (e MissingVariableError) Error() string {
	return fmt.Sprintf("variable %s not found", e.Key)
}

func CollectReferencedKeys(
	url string,
	headers map[string]string,
	query map[string]string,
	body map[string]any,
) []string {
	seen := map[string]struct{}{}
	add := func(key string) {
		if key == "" {
			return
		}
		seen[key] = struct{}{}
	}

	for _, match := range placeholderPattern.FindAllStringSubmatch(url, -1) {
		add(match[1])
	}
	for _, value := range headers {
		for _, match := range placeholderPattern.FindAllStringSubmatch(value, -1) {
			add(match[1])
		}
	}
	for _, value := range query {
		for _, match := range placeholderPattern.FindAllStringSubmatch(value, -1) {
			add(match[1])
		}
	}
	collectBodyRefsKey(body, add)

	out := make([]string, 0, len(seen))
	for key := range seen {
		out = append(out, key)
	}
	return out
}

func collectBodyRefsKey(value any, add func(string)) {
	switch typed := value.(type) {
	case map[string]any:
		if raw, ok := typed["$var"]; ok && len(typed) == 1 {
			switch keyRaw := raw.(type) {
			case string:
				add(keyRaw)
			}
			return
		}
		for _, child := range typed {
			collectBodyRefsKey(child, add)
		}
	case []any:
		for _, child := range typed {
			collectBodyRefsKey(child, add)
		}
	}
}

func ResolveTemplates(
	url string,
	headers map[string]string,
	query map[string]string,
	body map[string]any,
	context map[string]any,
) (string, map[string]string, map[string]string, map[string]any, error) {
	resolvedURL, err := resolveString(url, context)
	if err != nil {
		return "", nil, nil, nil, err
	}
	resolvedHeaders := make(map[string]string, len(headers))
	for key, value := range headers {
		resolved, err := resolveString(value, context)
		if err != nil {
			return "", nil, nil, nil, err
		}
		resolvedHeaders[key] = resolved
	}
	resolvedQuery := make(map[string]string, len(query))
	for key, value := range query {
		resolved, err := resolveString(value, context)
		if err != nil {
			return "", nil, nil, nil, err
		}
		resolvedQuery[key] = resolved
	}
	resolvedBody, err := resolveBody(body, context)
	if err != nil {
		return "", nil, nil, nil, err
	}
	bodyMap, _ := resolvedBody.(map[string]any)
	if bodyMap == nil {
		bodyMap = map[string]any{}
	}
	return resolvedURL, resolvedHeaders, resolvedQuery, bodyMap, nil
}

func resolveString(input string, context map[string]any) (string, error) {
	var missing *MissingVariableError
	out := placeholderPattern.ReplaceAllStringFunc(input, func(match string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(match, "{{"), "}}")
		value, ok := context[key]
		if !ok {
			missing = &MissingVariableError{Key: key}
			return match
		}
		return stringifyValue(value)
	})
	if missing != nil {
		return "", missing
	}
	return out, nil
}

func resolveBody(value any, context map[string]any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		if raw, ok := typed["$var"]; ok && len(typed) == 1 {
			keyStr, ok := raw.(string)
			if !ok {
				return nil, ErrInvalidRef
			}
			value, ok := context[keyStr]
			if !ok {
				return nil, &MissingVariableError{Key: keyStr}
			}
			return value, nil
		}
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			resolved, err := resolveBody(child, context)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			resolved, err := resolveBody(child, context)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return value, nil
	}
}

func stringifyValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64, float32, int, int64, int32, bool:
		return fmt.Sprint(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}
