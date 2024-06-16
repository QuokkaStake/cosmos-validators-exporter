package config

type Queries map[string]bool

func (q Queries) Enabled(query string) bool {
	if value, ok := q[query]; !ok {
		return true // all queries are enabled by default
	} else {
		return value
	}
}
