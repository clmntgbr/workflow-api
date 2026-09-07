package runexport

import (
	"regexp"
	"strings"
	"time"
)

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func ExportFilename(workflowName string, from, to time.Time) string {
	slug := slugify(workflowName)
	return slug + "-runs-" + from.UTC().Format("20060102") + "-" + to.UTC().Format("20060102") + ".xlsx"
}

func slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonSlug.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "workflow"
	}
	if len(s) > 60 {
		s = strings.Trim(s[:60], "-")
	}
	return s
}
