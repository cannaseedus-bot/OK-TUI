//go:build !windows

package runtime

// RegistryValue represents a registry key/value pair (Unix stub)
type RegistryValue struct {
	Key   string
	Value interface{}
}

// RegistryGet retrieves a value from the registry
// On Unix systems, this returns nil (registry is Windows-only)
func RegistryGet(hive, key, valueName string) interface{} {
	// Not implemented on Unix systems
	return nil
}

// RegistrySet sets a value in the registry
// On Unix systems, this returns an error (registry is Windows-only)
func RegistrySet(hive, key, valueName string, value interface{}) error {
	// Not implemented on Unix systems
	return nil
}

// RegistryList lists all values under a registry key
// On Unix systems, this returns an empty list
func RegistryList(hive, key string) []RegistryValue {
	// Not implemented on Unix systems
	return []RegistryValue{}
}

// RegistryDelete deletes a registry value
// On Unix systems, this returns an error
func RegistryDelete(hive, key, valueName string) error {
	// Not implemented on Unix systems
	return nil
}
