package httpquery

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrInvalidURL   = errors.New("invalid url")
	ErrInvalidQuery = errors.New("invalid query params")
)

// Params maps query keys to a string or a list of strings.
// JSON shape: Record<string, string | string[]>
type Params map[string]Value

// Value holds one or more query values for a single key.
type Value struct {
	items []string
}

func Single(s string) Value {
	return Value{items: []string{s}}
}

func Multi(values []string) Value {
	return Value{items: append([]string(nil), values...)}
}

func (v Value) Strings() []string {
	return append([]string(nil), v.items...)
}

func (v Value) Len() int {
	return len(v.items)
}

func (v Value) IsMulti() bool {
	return len(v.items) > 1
}

func (v Value) MarshalJSON() ([]byte, error) {
	if len(v.items) == 0 {
		return json.Marshal("")
	}
	if len(v.items) == 1 {
		return json.Marshal(v.items[0])
	}
	return json.Marshal(v.items)
}

func (v *Value) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		v.items = []string{""}
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		v.items = []string{asString}
		return nil
	}

	var asStrings []string
	if err := json.Unmarshal(data, &asStrings); err == nil {
		v.items = asStrings
		return nil
	}

	var asAny []any
	if err := json.Unmarshal(data, &asAny); err == nil {
		items := make([]string, 0, len(asAny))
		for _, item := range asAny {
			s, err := coerceString(item)
			if err != nil {
				return fmt.Errorf("%w: array value must be string-compatible", ErrInvalidQuery)
			}
			items = append(items, s)
		}
		v.items = items
		return nil
	}

	var asNumber json.Number
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	if err := dec.Decode(&asNumber); err == nil {
		v.items = []string{asNumber.String()}
		return nil
	}

	var asBool bool
	if err := json.Unmarshal(data, &asBool); err == nil {
		v.items = []string{strconv.FormatBool(asBool)}
		return nil
	}

	return fmt.Errorf("%w: value must be string or string[]", ErrInvalidQuery)
}

// Normalize converts a loosely typed map into Params.
// Accepts string, []string, []any, numbers and bools.
func Normalize(raw map[string]any) (Params, error) {
	if raw == nil {
		return Params{}, nil
	}
	out := make(Params, len(raw))
	for key, value := range raw {
		normalized, err := normalizeValue(value)
		if err != nil {
			return nil, fmt.Errorf("%w: key %q", err, key)
		}
		out[key] = normalized
	}
	return out, nil
}

func normalizeValue(value any) (Value, error) {
	switch typed := value.(type) {
	case nil:
		return Single(""), nil
	case Value:
		return Multi(typed.Strings()), nil
	case string:
		return Single(typed), nil
	case []string:
		return Multi(typed), nil
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			s, err := coerceString(item)
			if err != nil {
				return Value{}, ErrInvalidQuery
			}
			items = append(items, s)
		}
		return Multi(items), nil
	default:
		s, err := coerceString(typed)
		if err != nil {
			return Value{}, ErrInvalidQuery
		}
		return Single(s), nil
	}
}

func coerceString(value any) (string, error) {
	switch typed := value.(type) {
	case nil:
		return "", nil
	case string:
		return typed, nil
	case json.Number:
		return typed.String(), nil
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(typed), nil
	case int64:
		return strconv.FormatInt(typed, 10), nil
	case int32:
		return strconv.FormatInt(int64(typed), 10), nil
	case bool:
		return strconv.FormatBool(typed), nil
	default:
		return "", ErrInvalidQuery
	}
}

// SplitURLAndQuery extracts query params from a URL, preserving duplicates.
// Returns the URL without query/fragment.
// The base URL keeps the original path characters (e.g. {{order-id}}) instead of
// re-encoding them via url.URL.String().
func SplitURLAndQuery(rawURL string) (string, Params, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", nil, ErrInvalidURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", nil, ErrInvalidURL
	}

	params := Params{}
	if parsed.RawQuery != "" {
		values, err := url.ParseQuery(parsed.RawQuery)
		if err != nil {
			return "", nil, fmt.Errorf("%w: %v", ErrInvalidURL, err)
		}
		for key, list := range values {
			if len(list) == 1 {
				params[key] = Single(list[0])
				continue
			}
			params[key] = Multi(list)
		}
	}

	baseURL := stripQueryAndFragment(rawURL)
	if baseURL == "" {
		return "", nil, ErrInvalidURL
	}
	return baseURL, params, nil
}

func stripQueryAndFragment(rawURL string) string {
	base := rawURL
	if idx := strings.IndexByte(base, '?'); idx >= 0 {
		base = base[:idx]
	}
	if idx := strings.IndexByte(base, '#'); idx >= 0 {
		base = base[:idx]
	}
	return strings.TrimSpace(base)
}

// Merge combines URL-extracted params with body params.
// Body wins on key conflict. Body arrays replace entirely.
func Merge(fromURL, fromBody Params) Params {
	out := Clone(fromURL)
	if out == nil {
		out = Params{}
	}
	for key, value := range fromBody {
		out[key] = Multi(value.Strings())
	}
	return out
}

// ResolveURLAndQuery splits url query, normalizes body query, merges (body wins).
func ResolveURLAndQuery(rawURL string, bodyQuery Params) (string, Params, error) {
	baseURL, fromURL, err := SplitURLAndQuery(rawURL)
	if err != nil {
		return "", nil, err
	}
	if bodyQuery == nil {
		bodyQuery = Params{}
	}
	return baseURL, Merge(fromURL, bodyQuery), nil
}

// BuildQueryString rebuilds a query string, repeating keys for multi-values.
// Example: groups[]=a&groups[]=b (never groups[]=a,b).
func BuildQueryString(params Params) string {
	if len(params) == 0 {
		return ""
	}

	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	first := true
	for _, key := range keys {
		for _, value := range params[key].items {
			if !first {
				b.WriteByte('&')
			}
			first = false
			b.WriteString(encodeQueryKey(key))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(value))
		}
	}
	return b.String()
}

// encodeQueryKey keeps [] literal (PHP/API Platform style: groups[]=a).
func encodeQueryKey(key string) string {
	escaped := url.QueryEscape(key)
	escaped = strings.ReplaceAll(escaped, "%5B", "[")
	escaped = strings.ReplaceAll(escaped, "%5D", "]")
	return escaped
}

func Clone(params Params) Params {
	if params == nil {
		return Params{}
	}
	out := make(Params, len(params))
	for key, value := range params {
		out[key] = Multi(value.Strings())
	}
	return out
}

func Empty() Params {
	return Params{}
}
