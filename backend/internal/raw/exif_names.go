package raw

import "fmt"

func enrichTagNames(res *Result) {
	if res.Tags == nil {
		return
	}
	named := map[string]any{}
	for k, v := range res.Tags {
		named[k] = v
		var id uint16
		if _, err := fmt.Sscanf(k, "0x%X", &id); err == nil {
			if n := TagName(id); n != "" {
				named[n] = v
			}
		}
	}
	res.Tags = named
}
