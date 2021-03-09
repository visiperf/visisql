package visisql

import (
	"sort"

	"github.com/huandu/go-sqlbuilder"
)

func extractMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func extractMapValues(m map[string]interface{}, keys []string) []interface{} {
	vals := make([]interface{}, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, m[k])
	}

	return vals
}

func assignMap(m map[string]interface{}, keys []string, builder *sqlbuilder.UpdateBuilder) []string {
	assignements := make([]string, 0, len(keys))
	for _, k := range keys {
		assignements = append(assignements, builder.Assign(k, m[k]))
	}

	return assignements
}
