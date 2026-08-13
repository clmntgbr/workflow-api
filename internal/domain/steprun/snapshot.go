package steprun

type ResponseSnapshot struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    any               `json:"body"`
}

func (s ResponseSnapshot) Normalized() ResponseSnapshot {
	headers := s.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	return ResponseSnapshot{
		Status:  s.Status,
		Headers: headers,
		Body:    s.Body,
	}
}

func normalizeStringMap(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}
	return values
}

func normalizeAnyMap(values map[string]any) map[string]any {
	if values == nil {
		return map[string]any{}
	}
	return values
}
