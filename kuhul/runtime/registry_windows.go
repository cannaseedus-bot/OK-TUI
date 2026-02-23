//go:build windows

package runtime

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// RegistryValue represents a registry key/value pair
type RegistryValue struct {
	Key   string
	Value interface{}
}

// GetRegistryHive converts a hive name string to a registry hive constant
func GetRegistryHive(hiveName string) registry.Key {
	hiveStr := strings.ToUpper(hiveName)
	switch hiveStr {
	case "HKCR", "HKEY_CLASSES_ROOT":
		return registry.CLASSES_ROOT
	case "HKCU", "HKEY_CURRENT_USER":
		return registry.CURRENT_USER
	case "HKLM", "HKEY_LOCAL_MACHINE":
		return registry.LOCAL_MACHINE
	case "HKU", "HKEY_USERS":
		return registry.USERS
	case "HKCC", "HKEY_CURRENT_CONFIG":
		return registry.CURRENT_CONFIG
	case "HKPD", "HKEY_PERFORMANCE_DATA":
		return registry.PERFORMANCE_DATA
	default:
		// Default to HKEY_CURRENT_USER
		return registry.CURRENT_USER
	}
}

// RegistryGet retrieves a value from the Windows registry
func RegistryGet(hive, keyPath, valueName string) interface{} {
	hiveKey := GetRegistryHive(hive)

	// Open the registry key
	key, err := registry.OpenKey(hiveKey, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer key.Close()

	// Get the value
	value, _, err := key.GetStringValue(valueName)
	if err != nil {
		// Try as DWORD if string fails
		intValue, _, err := key.GetIntegerValue(valueName)
		if err != nil {
			return nil
		}
		return intValue
	}

	return value
}

// RegistrySet sets a value in the Windows registry
func RegistrySet(hive, keyPath, valueName string, value interface{}) error {
	hiveKey := GetRegistryHive(hive)

	// Open or create the registry key
	key, err := registry.OpenKey(hiveKey, keyPath, registry.SET_VALUE)
	if err != nil {
		// Try to create if it doesn't exist
		key, _, err = registry.CreateKey(hiveKey, keyPath, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("failed to open/create registry key: %w", err)
		}
	}
	defer key.Close()

	// Set the value based on its type
	switch v := value.(type) {
	case string:
		return key.SetStringValue(valueName, v)
	case int, int32, int64, float64:
		intVal := int64(0)
		switch val := v.(type) {
		case int:
			intVal = int64(val)
		case int32:
			intVal = int64(val)
		case int64:
			intVal = val
		case float64:
			intVal = int64(val)
		}
		return key.SetDWordValue(valueName, uint32(intVal))
	default:
		// Convert to string
		return key.SetStringValue(valueName, fmt.Sprintf("%v", value))
	}
}

// RegistryList lists all values under a registry key
func RegistryList(hive, keyPath string) []RegistryValue {
	result := []RegistryValue{}

	hiveKey := GetRegistryHive(hive)

	// Open the registry key
	key, err := registry.OpenKey(hiveKey, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return result
	}
	defer key.Close()

	// Enumerate values
	valueNames, err := key.ReadValueNames(-1)
	if err != nil {
		return result
	}

	for _, valueName := range valueNames {
		value := RegistryGet(hive, keyPath, valueName)
		result = append(result, RegistryValue{
			Key:   valueName,
			Value: value,
		})
	}

	return result
}

// RegistryDelete deletes a registry value
func RegistryDelete(hive, keyPath, valueName string) error {
	hiveKey := GetRegistryHive(hive)

	// Open the registry key
	key, err := registry.OpenKey(hiveKey, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Delete the value
	return key.DeleteValue(valueName)
}

// RegistryListSubkeys lists all subkeys under a registry key
func RegistryListSubkeys(hive, keyPath string) []string {
	result := []string{}

	hiveKey := GetRegistryHive(hive)

	// Open the registry key
	key, err := registry.OpenKey(hiveKey, keyPath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return result
	}
	defer key.Close()

	// Enumerate subkeys
	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return result
	}

	return subkeys
}
