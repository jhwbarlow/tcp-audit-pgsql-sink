package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	hostEnvVar     = "PGHOST"
	portEnvVar     = "PGPORT"
	dbEnvVar       = "PGDATABASE"
	userEnvVar     = "PGUSER"
	passwordEnvVar = "PGPASSWORD"
)

// ConfigGetter is an interface which describes objects which provide
// a datastore connection string based upon some configuration source.
type configGetter interface {
	config() (string, error)
}

// EnvVarConfigGetter provides a PostgreSQL database connection string
// from configuration provided in the standard PostgreSQL client environment
// variables.
// See https://www.postgresql.org/docs/current/libpq-envars.html
type envVarConfigGetter struct{}

// Config returns a PostgreSQL database connection string based upon values
// provided in environment variables.
func (cg *envVarConfigGetter) config() (string, error) {
	host := os.Getenv(hostEnvVar)
	if host == "" {
		return "", fmt.Errorf("environment variable %s not set", hostEnvVar)
	}

	var port uint16 = 5432
	if os.Getenv(portEnvVar) != "" {
		portVar, err := strconv.ParseUint(os.Getenv(portEnvVar), 10, 16)
		if err != nil {
			return "", fmt.Errorf("environment variable %s has invalid value", portEnvVar)
		}
		port = uint16(portVar)
	}

	db := os.Getenv(dbEnvVar)
	if db == "" {
		return "", fmt.Errorf("environment variable %s not set", dbEnvVar)
	}

	user := os.Getenv(userEnvVar)
	if user == "" {
		return "", fmt.Errorf("environment variable %s not set", userEnvVar)
	}

	password := os.Getenv(passwordEnvVar)
	if password == "" {
		return "", fmt.Errorf("environment variable %s not set", passwordEnvVar)
	}

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		user,
		password,
		host,
		port,
		db)
	return connStr, nil
}
