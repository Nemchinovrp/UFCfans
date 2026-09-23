package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnv(t *testing.T) {
	t.Setenv("UFCFANS_EXISTING", "keep")
	const key = "UFCFANS_TEST_LITERAL"
	os.Unsetenv(key)
	t.Cleanup(func() { os.Unsetenv(key) })
	p := filepath.Join(t.TempDir(), ".env")
	os.WriteFile(p, []byte("# comment\nexport UFCFANS_TEST_LITERAL='$(no-execution) # literal' # comment\nUFCFANS_EXISTING=replace\n"), 0600)
	if err := loadEnv(p); err != nil {
		t.Fatal(err)
	}
	if os.Getenv(key) != "$(no-execution) # literal" || os.Getenv("UFCFANS_EXISTING") != "keep" {
		t.Fatal("incorrect dotenv precedence or literal parsing")
	}
	os.WriteFile(p, []byte("SECRET='unclosed"), 0600)
	if err := loadEnv(p); err == nil {
		t.Fatal("expected syntax error")
	}
}
