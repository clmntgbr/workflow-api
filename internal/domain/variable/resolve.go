package variable

import (
	"encoding/json"
	"fmt"
	"regexp"

	"go-api/internal/domain/httpquery"
)

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_-]+)\s*\}\}`)

type MissingVariableError struct {
	Key string
}

func (e MissingVariableError) Error() string {
	return fmt.Sprintf("variable %s not found", e.Key)
}

func CollectReferencedKeys(
	url string,
	headers map[string]string,
	query httpquery.Params,
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
		for _, item := range value.Strings() {
			for _, match := range placeholderPattern.FindAllStringSubmatch(item, -1) {
				add(match[1])
			}
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
		for _, child := range typed {
			collectBodyRefsKey(child, add)
		}
	case []any:
		for _, child := range typed {
			collectBodyRefsKey(child, add)
		}
	case string:
		for _, match := range placeholderPattern.FindAllStringSubmatch(typed, -1) {
			add(match[1])
		}
	}
}

func ResolveTemplates(
	url string,
	headers map[string]string,
	query httpquery.Params,
	body map[string]any,
	context map[string]any,
) (string, map[string]string, httpquery.Params, map[string]any, error) {
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
	resolvedQuery := make(httpquery.Params, len(query))
	for key, value := range query {
		items := value.Strings()
		resolvedItems := make([]string, len(items))
		for i, item := range items {
			resolved, err := resolveString(item, context)
			if err != nil {
				return "", nil, nil, nil, err
			}
			resolvedItems[i] = resolved
		}
		resolvedQuery[key] = httpquery.Multi(resolvedItems)
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
		sub := placeholderPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := sub[1]
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
	case string:
		return resolveString(typed, context)
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
