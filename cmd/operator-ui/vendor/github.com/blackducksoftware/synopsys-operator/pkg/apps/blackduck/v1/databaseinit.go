/*
Copyright (C) 2019 Synopsys, Inc.
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package blackduck

import (
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/database"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
)

// InitDatabase will init the database
func InitDatabase(createHub *v1.BlackduckSpec, adminPassword string, userPassword string, postgresPassword string) error {
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", createHub.Namespace)
	// postgres database
	err := database.ExecDBStatements(hostName, "postgres", "postgres", postgresPassword, "postgres", []string{
		fmt.Sprintf("ALTER USER blackduck WITH password '%s';", adminPassword),
		"GRANT blackduck TO postgres;",
		"CREATE DATABASE bds_hub owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;",
		"CREATE DATABASE bds_hub_report owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;",
		"CREATE DATABASE bdio owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;",
		"ALTER USER blackduck WITH NOCREATEDB SUPERUSER NOREPLICATION BYPASSRLS;",
		"CREATE USER blackduck_user WITH NOCREATEDB NOSUPERUSER NOREPLICATION NOBYPASSRLS;",
		fmt.Sprintf("ALTER USER blackduck_user WITH password '%s';", userPassword),
		"CREATE USER blackduck_reporter;",
		"CREATE USER blackduck_replication REPLICATION CONNECTION LIMIT 5;",
	}, true)
	if err != nil {
		return err
	}

	// bds_hub database
	err = database.ExecDBStatements(hostName, "bds_hub", "postgres", postgresPassword, "postgres", []string{
		"CREATE EXTENSION pgcrypto;",
		"CREATE SCHEMA st AUTHORIZATION blackduck;",
		"GRANT USAGE ON SCHEMA st TO blackduck_user;",
		"GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA st TO blackduck_user;",
		"GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA st to blackduck_user;",
		"ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;",
		"ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;",
		"ALTER DATABASE bds_hub SET standard_conforming_strings TO OFF;",
	}, false)
	if err != nil {
		return err
	}

	// bds_hub_report database
	err = database.ExecDBStatements(hostName, "bds_hub_report", "postgres", postgresPassword, "postgres", []string{
		"GRANT SELECT ON ALL TABLES IN SCHEMA public TO blackduck_reporter;",
		"ALTER DEFAULT PRIVILEGES FOR ROLE blackduck IN SCHEMA public GRANT SELECT ON TABLES TO blackduck_reporter;",
		"GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA public TO blackduck_user;",
		"GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public to blackduck_user;",
		"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;",
		"ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;",
		"ALTER DATABASE bds_hub_report SET standard_conforming_strings TO OFF;",
	}, false)
	if err != nil {
		return err
	}

	// bdio database
	err = database.ExecDBStatements(hostName, "bdio", "postgres", postgresPassword, "postgres", []string{
		"GRANT ALL PRIVILEGES ON DATABASE bdio TO blackduck_user;",
		"ALTER DATABASE bdio SET standard_conforming_strings TO ON;",
	}, false)
	if err != nil {
		return err
	}

	// System parameters
	err = database.ExecDBStatements(hostName, "postgres", "postgres", postgresPassword, "postgres", []string{
		"ALTER SYSTEM SET autovacuum TO 'on';",
		"ALTER SYSTEM SET autovacuum_max_workers TO '20';",
		"ALTER SYSTEM SET autovacuum_vacuum_cost_limit TO '2000';",
		"ALTER SYSTEM SET autovacuum_vacuum_cost_delay TO '10ms';",
		"ALTER SYSTEM SET checkpoint_completion_target TO '0.8';",
		"ALTER SYSTEM SET max_wal_size TO '8GB';",
		"ALTER SYSTEM SET checkpoint_timeout TO '30min';",
		"ALTER SYSTEM SET constraint_exclusion TO 'partition';",
		"ALTER SYSTEM SET default_statistics_target TO '100';",
		"ALTER SYSTEM SET effective_cache_size TO '256MB';",
		"ALTER SYSTEM SET escape_string_warning TO 'off';",
		"ALTER SYSTEM SET log_destination TO 'stderr';",
		"ALTER SYSTEM SET log_directory TO 'pg_log';",
		"ALTER SYSTEM SET log_filename TO 'postgresql_%a.log';",
		"ALTER SYSTEM SET log_line_prefix TO '%m %p ';",
		"ALTER SYSTEM SET log_rotation_age TO '1440';",
		"ALTER SYSTEM SET log_truncate_on_rotation TO 'on';",
		"ALTER SYSTEM SET logging_collector TO 'on';",
		"ALTER SYSTEM SET maintenance_work_mem TO '32MB';",
		"ALTER SYSTEM SET max_connections TO '300';",
		"ALTER SYSTEM SET max_locks_per_transaction TO '256';",
		"ALTER SYSTEM SET random_page_cost TO '4.0';",
		"ALTER SYSTEM SET shared_buffers TO '1024MB';",
		"ALTER SYSTEM SET standard_conforming_strings TO 'off';",
		"ALTER SYSTEM SET temp_buffers TO '16MB';",
		"ALTER SYSTEM SET work_mem TO '32MB';",
	}, false)
	if err != nil {
		return err
	}

	return nil
}
