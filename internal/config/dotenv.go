package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// loadEnv reads literal values; it never executes shell commands or expands variables.
// Existing environment variables take precedence.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for line := 1; scanner.Scan(); line++ {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || strings.HasPrefix(s, "#") {
			continue
		}
		s = strings.TrimPrefix(s, "export ")
		key, value, ok := strings.Cut(s, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if !ok || key == "" {
			return fmt.Errorf("invalid .env assignment at line %d", line)
		}
		for i, c := range key {
			if !(c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && c >= '0' && c <= '9') {
				return fmt.Errorf("invalid .env variable at line %d", line)
			}
		}
		if len(value) > 0 && (value[0] == '\'' || value[0] == '"') {
			quote := value[0]
			end := strings.IndexByte(value[1:], quote)
			if end < 0 {
				return fmt.Errorf("unclosed .env quote at line %d", line)
			}
			end++
			tail := strings.TrimSpace(value[end+1:])
			if tail != "" && !strings.HasPrefix(tail, "#") {
				return fmt.Errorf("invalid .env value at line %d", line)
			}
			value = value[1:end]
		} else {
			if i := strings.Index(value, " #"); i >= 0 {
				value = strings.TrimSpace(value[:i])
			}
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("cannot set .env variable at line %d", line)
			}
		}
	}
	return scanner.Err()
}
