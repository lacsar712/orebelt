package config

import "testing"

func TestDefaultValidate(t *testing.T) {
	cfg := Default()
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRejectsSmallRing(t *testing.T) {
	cfg := Default()
	cfg.RingCapacity = 4
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestFromEnv(t *testing.T) {
	t.Setenv("OREBELT_LISTEN", ":9090")
	cfg := FromEnv(Default())
	if cfg.ListenAddr != ":9090" {
		t.Fatalf("got %s", cfg.ListenAddr)
	}
}

func TestParseArgs(t *testing.T) {
	cfg, err := ParseArgs([]string{"-listen", ":7070"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddr != ":7070" {
		t.Fatal(cfg.ListenAddr)
	}
}
