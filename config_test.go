package config

import (
	"context"
	"os"
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

func TestParseFile(t *testing.T) {
	cfg := &TestConfig{} 

	err := Parse(context.Background(), "config-test.json", cfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = Dump(cfg, "config-dump.json")
	if err != nil {
		t.Fatalf("%v", err)
	}

	newCfg := &TestConfig{}

	err = Parse(context.Background(), "config-dump.json", newCfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(cfg, newCfg) {
		t.Fatalf("mismatch")
	}
}

func TestParseSecret(t *testing.T) {
	cfg := &TestConfig{} 

	err := Parse(context.Background(), "config-test.json", cfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	data, err := Serialize(cfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	project := os.Getenv("PROJECT")

	version, err := SaveSecret(context.Background(), project, "test-secret",  data)
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("version %s", version)

	newCfg := &TestConfig{}

	path := SecretPath(project, "test-secret", "latest")

	err = Parse(context.Background(), path, newCfg)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if !reflect.DeepEqual(cfg, newCfg) {
		t.Fatalf("mismatch")
	}
}
