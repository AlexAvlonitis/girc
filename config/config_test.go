package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(writeConfig(t, "server: irc.example.com\nnick: gopher\nssl: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{Server: "irc.example.com", Port: 6697, Nick: "gopher", User: "gopher", RealName: "gopher", TLS: true}
	if cfg != want {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
	if got := cfg.Addr(); got != "irc.example.com:6697" {
		t.Errorf("Addr() = %q", got)
	}
}

func TestLoadInvalid(t *testing.T) {
	for name, content := range map[string]string{
		"missing server": "nick: gopher\n",
		"missing nick":   "server: irc.example.com\n",
		"bad port":       "server: s\nnick: n\nport: 70000\n",
		"bad yaml":       "server: [\n",
	} {
		if _, err := Load(writeConfig(t, content)); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if _, err := Load(filepath.Join(t.TempDir(), "missing.yaml")); err == nil {
		t.Error("missing file: expected error")
	}
}
