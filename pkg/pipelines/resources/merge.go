package resources

// Resources is a mapping of filename to k8s resource suitable for marshaling to
// YAML.
type Resources map[string]interface{}

// Merge accepts to sets of resources, and merges them overwriting resources in
// the "to" set with resources from the "from" set.
func Merge(from, to Resources) Resources {
	merged := Resources{}
	for k, v := range to {
		merged[k] = v
	}
	for k, v := range from {
		merged[k] = v
	}
	return merged
}
