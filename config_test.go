package config

import (
	"testing"
	"reflect"
)

type TestConfig struct {
	Address                  string
	Project                  string
	FirestoreDatabaseID      string
	FirebaseAdminCredentials string
	FirebaseAPIKey           string
}

func TestParse(t *testing.T) {
	cfg := &TestConfig{} 

	err := Parse("config-test.json", cfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = Dump(cfg, "config-dump.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	newCfg := &TestConfig{}

	err = Parse("config-dump.json", newCfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(cfg, newCfg) {
		t.Fatalf("mismatch")
	}
}
