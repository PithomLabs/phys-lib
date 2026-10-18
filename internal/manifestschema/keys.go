package manifestschema

// topLevelKeys decodes a manifest's top-level JSON object keys without
// interpreting any value. It deliberately does not use DisallowUnknownFields:
// detecting unknown keys is this package's job, reported as conformance
// violations rather than a decode error.

import (
	"encoding/json"
	"fmt"
)

func topLevelKeys(raw []byte) ([]string, error) {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("manifest is not a JSON object: %w", err)
	}
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys, nil
}
