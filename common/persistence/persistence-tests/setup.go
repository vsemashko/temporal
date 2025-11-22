package persistencetests

import (
	"go.temporal.io/server/common/config"
	"go.temporal.io/server/common/persistence/sql/sqlplugin/mysql"
	"go.temporal.io/server/common/persistence/sql/sqlplugin/postgresql"
	"go.temporal.io/server/common/persistence/sql/sqlplugin/sqlite"
	"go.temporal.io/server/temporal/environment"
)

const (
	// ============================================================================
	// SECURITY WARNING: TEST CREDENTIALS ONLY - DO NOT USE IN PRODUCTION
	// ============================================================================
	// These credentials are ONLY for automated testing and development.
	// They represent WEAK credentials that should NEVER be used in production.
	// In production, use strong, randomly-generated passwords from a secrets
	// management system (e.g., HashiCorp Vault, AWS Secrets Manager).
	// ============================================================================

	testMySQLUser      = "temporal" // TEST ONLY - DO NOT USE IN PRODUCTION
	testMySQLPassword  = "temporal" // TEST ONLY - WEAK PASSWORD - DO NOT USE IN PRODUCTION
	testMySQLSchemaDir = "schema/mysql/v8"

	testPostgreSQLUser      = "temporal" // TEST ONLY - DO NOT USE IN PRODUCTION
	testPostgreSQLPassword  = "temporal" // TEST ONLY - WEAK PASSWORD - DO NOT USE IN PRODUCTION
	testPostgreSQLSchemaDir = "schema/postgresql/v12"

	testSQLiteUser      = "" // TEST ONLY
	testSQLitePassword  = "" // TEST ONLY
	testSQLiteMode      = "memory"
	testSQLiteCache     = "private"
	testSQLiteSchemaDir = "schema/sqlite/v3" // specify if mode is not "memory"
)

// GetMySQLTestClusterOption return test options
func GetMySQLTestClusterOption() *TestBaseOptions {
	return &TestBaseOptions{
		SQLDBPluginName: mysql.PluginName,
		DBUsername:      testMySQLUser,
		DBPassword:      testMySQLPassword,
		DBHost:          environment.GetMySQLAddress(),
		DBPort:          environment.GetMySQLPort(),
		SchemaDir:       testMySQLSchemaDir,
		StoreType:       config.StoreTypeSQL,
	}
}

// GetPostgreSQLTestClusterOption return test options
func GetPostgreSQLTestClusterOption() *TestBaseOptions {
	return &TestBaseOptions{
		SQLDBPluginName: postgresql.PluginName,
		DBUsername:      testPostgreSQLUser,
		DBPassword:      testPostgreSQLPassword,
		DBHost:          environment.GetPostgreSQLAddress(),
		DBPort:          environment.GetPostgreSQLPort(),
		SchemaDir:       testPostgreSQLSchemaDir,
		StoreType:       config.StoreTypeSQL,
	}
}

// GetPostgreSQLPGXTestClusterOption return test options
func GetPostgreSQLPGXTestClusterOption() *TestBaseOptions {
	return &TestBaseOptions{
		SQLDBPluginName: postgresql.PluginNamePGX,
		DBUsername:      testPostgreSQLUser,
		DBPassword:      testPostgreSQLPassword,
		DBHost:          environment.GetPostgreSQLAddress(),
		DBPort:          environment.GetPostgreSQLPort(),
		SchemaDir:       testPostgreSQLSchemaDir,
		StoreType:       config.StoreTypeSQL,
	}
}

// GetSQLiteFileTestClusterOption return test options
func GetSQLiteFileTestClusterOption() *TestBaseOptions {
	return &TestBaseOptions{
		SQLDBPluginName:   sqlite.PluginName,
		DBUsername:        testSQLiteUser,
		DBPassword:        testSQLitePassword,
		DBHost:            environment.GetLocalhostIP(),
		DBPort:            0,
		SchemaDir:         testSQLiteSchemaDir,
		StoreType:         config.StoreTypeSQL,
		ConnectAttributes: map[string]string{"cache": testSQLiteCache},
	}
}

// GetSQLiteMemoryTestClusterOption return test options
func GetSQLiteMemoryTestClusterOption() *TestBaseOptions {
	return &TestBaseOptions{
		SQLDBPluginName:   sqlite.PluginName,
		DBUsername:        testSQLiteUser,
		DBPassword:        testSQLitePassword,
		DBHost:            environment.GetLocalhostIP(),
		DBPort:            0,
		SchemaDir:         "",
		StoreType:         config.StoreTypeSQL,
		ConnectAttributes: map[string]string{"mode": testSQLiteMode, "cache": testSQLiteCache},
	}
}
