package env

import "strings"

// ListToMap converts a list of strings in format KEY=VALUE into a map
func ListToMap(list []string) map[string]string {
	out := map[string]string{}
	for _, entry := range list {
		if len(entry) == 0 || entry[0] == '#' {
			continue
		}

		parts := strings.SplitN(entry, "=", 2)
		out[parts[0]] = strings.Trim(parts[1], "\"")
	}
	return out
}

// MapToList converts a map into a list of strings in format KEY=VALUE
func MapToList(envMap map[string]string) []string {
	out := []string{}
	for k, v := range envMap {
		out = append(out, k+"="+v)
	}
	return out
}
