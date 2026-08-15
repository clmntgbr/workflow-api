package variable

import (
	"strconv"
	"strings"
)

const maxArrayElements = 5

func ExtractResponsePaths(data any, query string) []string {
	return extractPaths(data, "$", strings.ToLower(strings.TrimSpace(query)))
}

func extractPaths(data any, currentPath string, queryLower string) []string {
	var paths []string

	switch typed := data.(type) {
	case map[string]any:
		for key, value := range typed {
			newPath := currentPath + "." + key
			if matchesPathQuery(newPath, key, queryLower) {
				paths = append(paths, newPath)
			}
			paths = append(paths, extractPaths(value, newPath, queryLower)...)
		}
	case []any:
		maxElements := len(typed)
		if maxElements > maxArrayElements {
			maxElements = maxArrayElements
		}
		for i := 0; i < maxElements; i++ {
			indexPath := currentPath + "[" + strconv.Itoa(i) + "]"
			indexPathLower := strings.ToLower(indexPath)
			shouldRecurse := queryLower == "" ||
				strings.Contains(indexPathLower, queryLower) ||
				strings.HasPrefix(queryLower, indexPathLower) ||
				strings.HasPrefix(queryLower, strings.TrimPrefix(indexPathLower, "$."))

			if !shouldRecurse {
				continue
			}

			switch typed[i].(type) {
			case map[string]any, []any:
				paths = append(paths, extractPaths(typed[i], indexPath, queryLower)...)
			default:
				if matchesPathQuery(indexPath, "", queryLower) {
					paths = append(paths, indexPath)
				}
			}
		}
	}

	return paths
}

func matchesPathQuery(path, key, queryLower string) bool {
	if queryLower == "" {
		return true
	}
	pathLower := strings.ToLower(path)
	keyLower := strings.ToLower(key)
	return strings.Contains(pathLower, queryLower) ||
		(keyLower != "" && strings.Contains(keyLower, queryLower))
}
