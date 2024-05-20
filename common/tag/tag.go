package tag

import "strings"

func ParseTagSettings(tag string, sep string) map[string]string {
	settings := make(map[string]string)
	names := strings.Split(tag, sep)

	for i := 0; i < len(names); i++ {
		values := strings.Split(names[i], ":")
		k := strings.TrimSpace(values[0])
		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}
