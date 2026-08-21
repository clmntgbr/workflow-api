package variable

// SeedContextWithStaticVariables copies base and sets static variable values.
// Keys already present in base are left unchanged (caller context wins).
func SeedContextWithStaticVariables(base map[string]any, variables []VariableView) map[string]any {
	out := make(map[string]any, len(base)+len(variables))
	for key, value := range base {
		out[key] = value
	}
	for _, variable := range variables {
		if variable.Kind != KindStatic {
			continue
		}
		if _, exists := out[variable.Key]; exists {
			continue
		}
		out[variable.Key] = variable.Value
	}
	return out
}
