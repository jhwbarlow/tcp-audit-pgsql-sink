package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func bootstrapEnv(host, port, db, user, password string) error {
	if err := os.Setenv(hostEnvVar, host); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", hostEnvVar, err)
	}

	if err := os.Setenv(portEnvVar, port); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", portEnvVar, err)
	}

	if err := os.Setenv(dbEnvVar, db); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", dbEnvVar, err)
	}

	if err := os.Setenv(userEnvVar, user); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", userEnvVar, err)
	}

	if err := os.Setenv(passwordEnvVar, password); err != nil {
		return fmt.Errorf("unable to set env var %q: %w", passwordEnvVar, err)
	}

	return nil
}

func destroyEnv() {
	os.Unsetenv(hostEnvVar)
	os.Unsetenv(portEnvVar)
	os.Unsetenv(dbEnvVar)
	os.Unsetenv(userEnvVar)
	os.Unsetenv(passwordEnvVar)
}

func TestGetConfigFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("mock-host",
		"7337",
		"mock-database",
		"mock-user",
		"mock-password"); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	connStr, err := configGetter.config()
	if err != nil {
		t.Errorf("expected nil error, got %q (of type %T)", err, err)
	}

	expectedConnStr := "postgresql://mock-user:mock-password@mock-host:7337/mock-database"
	if connStr != expectedConnStr {
		t.Errorf("expected conn string %q, got %q", expectedConnStr, connStr)
	}
}

func TestGetConfigErrorNoHostFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("",
		"7337",
		"mock-database",
		"mock-user",
		"mock-password"); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	_, err := configGetter.config()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !strings.Contains(err.Error(), hostEnvVar) {
		t.Errorf("expected error to contain env var name %q, but did not", hostEnvVar)
	}
}

func TestGetConfigErrorBadPortFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("mock-host",
		"NaN",
		"mock-database",
		"mock-user",
		"mock-password"); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	_, err := configGetter.config()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !strings.Contains(err.Error(), portEnvVar) {
		t.Errorf("expected error to contain env var name %q, but did not", portEnvVar)
	}
}

func TestGetConfigErrorNoDBFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("mock-host",
		"7337",
		"",
		"mock-user",
		"mock-password"); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	_, err := configGetter.config()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !strings.Contains(err.Error(), dbEnvVar) {
		t.Errorf("expected error to contain env var name %q, but did not", dbEnvVar)
	}
}

func TestGetConfigErrorNoUserFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("mock-host",
		"7337",
		"mock-database",
		"",
		"mock-password"); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	_, err := configGetter.config()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !strings.Contains(err.Error(), userEnvVar) {
		t.Errorf("expected error to contain env var name %q, but did not", userEnvVar)
	}
}

func TestGetConfigErrorNoPasswordFromEnv(t *testing.T) {
	defer destroyEnv()
	if err := bootstrapEnv("mock-host",
		"7337",
		"mock-database",
		"mock-user",
		""); err != nil {
		t.Fatalf("test bootstrapping: unable to set environment: %v", err)
	}

	configGetter := new(envVarConfigGetter)
	_, err := configGetter.config()
	if err == nil {
		t.Error("expected error, got nil")
	}

	t.Logf("got error %q (of type %T)", err, err)

	if !strings.Contains(err.Error(), passwordEnvVar) {
		t.Errorf("expected error to contain env var name %q, but did not", passwordEnvVar)
	}
}
