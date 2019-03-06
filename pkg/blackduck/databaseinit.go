/*
Copyright (C) 2018 Synopsys, Inc.

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
	"database/sql"
	"fmt"
	"time"

	"github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	// This is required to access the Postgres database
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// InitDatabase will init the database
func InitDatabase(createHub *v1.BlackduckSpec, adminPassword string, userPassword string, postgresPassword string) error {
	databaseName := "postgres"
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", createHub.Namespace)
	postgresDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}
	execPostGresDBStatements(postgresDB, adminPassword, userPassword)

	databaseName = "bds_hub"
	bdsHubDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}
	execBdsHubDBStatements(bdsHubDB)
	bdsHubDB.Close()

	databaseName = "bds_hub_report"
	bdsHubReportDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}
	execBdsHubReportDBStatements(bdsHubReportDB)
	bdsHubReportDB.Close()

	databaseName = "bdio"
	bdioDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", postgresPassword, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}
	execBdioDBStatements(bdioDB)
	bdioDB.Close()

	execSystemDBStatements(postgresDB)
	postgresDB.Close()

	return nil
}

// OpenDatabaseConnection will open the database connection
func OpenDatabaseConnection(hostName string, dbName string, user string, password string, sqlType string) (*sql.DB, error) {
	// Note that sslmode=disable is required it does not mean that the connection
	// is unencrypted. All connections via the proxy are completely encrypted.
	log.Debug("attempting to open database connection")
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable connect_timeout=10", hostName, dbName, user, password)
	db, err := sql.Open(sqlType, dsn)
	//defer db.Close()
	if err == nil {
		log.Debug("connected to database ")
	}
	return db, err
}

func execPostGresDBStatementsClone(db *sql.DB, adminPassword string, userPassword string) error {
	var err error
	(func() {
		_, err = db.Exec(fmt.Sprintf("ALTER USER blackduck WITH password '%s';", adminPassword))
		if dispErr(err) {
			return
		}

		_, err = db.Exec(fmt.Sprintf("ALTER USER blackduck_user WITH password '%s';", userPassword))
		if dispErr(err) {
			return
		}

	})()

	db.Close()
	return err
}

func exec(db *sql.DB, statement string) error {
	_, err := db.Exec(statement)
	if err != nil {
		log.Errorf("unable to exec %s statment due to %+v", statement, err)
	}
	return err
}

func execPostGresDBStatements(db *sql.DB, adminPassword string, userPassword string) {
	for {
		log.Debug("executing SELECT 1")
		err := exec(db, "SELECT 1;")
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}
	exec(db, fmt.Sprintf("ALTER USER blackduck WITH password '%s';", adminPassword))
	exec(db, "GRANT blackduck TO postgres;")
	exec(db, "CREATE DATABASE bds_hub owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;")
	exec(db, "CREATE DATABASE bds_hub_report owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;")
	exec(db, "CREATE DATABASE bdio owner blackduck TEMPLATE template0 ENCODING SQL_ASCII;")
	exec(db, "CREATE USER blackduck_user;")
	exec(db, fmt.Sprintf("ALTER USER blackduck_user WITH password '%s';", userPassword))
	exec(db, "CREATE USER blackduck_reporter;")
	// db.Close()
}

func execBdsHubDBStatements(db *sql.DB) {
	exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "CREATE SCHEMA st AUTHORIZATION blackduck;")
	exec(db, "GRANT USAGE ON SCHEMA st TO blackduck_user;")
	exec(db, "GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA st TO blackduck_user;")
	exec(db, "GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA st to blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA st GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;")
	// exec(db, "ALTER DATABASE bds_hub SET tcp_keepalives_idle TO 600;")
	// exec(db, "ALTER DATABASE bds_hub SET tcp_keepalives_interval TO 30;")
	// exec(db, "ALTER DATABASE bds_hub SET tcp_keepalives_count TO 10;")
	// db.Close()
}

func execBdsHubReportDBStatements(db *sql.DB) {
	// exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "GRANT SELECT ON ALL TABLES IN SCHEMA public TO blackduck_reporter;")
	exec(db, "ALTER DEFAULT PRIVILEGES FOR ROLE blackduck IN SCHEMA public GRANT SELECT ON TABLES TO blackduck_reporter;")
	exec(db, "GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON ALL TABLES IN SCHEMA public TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, TRUNCATE, DELETE, REFERENCES ON TABLES TO blackduck_user;")
	exec(db, "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO blackduck_user;")
	// exec(db, "ALTER DATABASE bds_hub_report SET tcp_keepalives_idle TO 600;")
	// exec(db, "ALTER DATABASE bds_hub_report SET tcp_keepalives_interval TO 30;")
	// exec(db, "ALTER DATABASE bds_hub_report SET tcp_keepalives_count TO 10;")
	// db.Close()
}

func execBdioDBStatements(db *sql.DB) {
	// exec(db, "CREATE EXTENSION pgcrypto;")
	exec(db, "GRANT ALL PRIVILEGES ON DATABASE bdio TO blackduck_user;")
	exec(db, "ALTER DATABASE bdio SET standard_conforming_strings TO ON;")
	// exec(db, "ALTER DATABASE bdio SET tcp_keepalives_idle TO 600;")
	// exec(db, "ALTER DATABASE bdio SET tcp_keepalives_interval TO 30;")
	// exec(db, "ALTER DATABASE bdio SET tcp_keepalives_count TO 10;")
	// db.Close()
}

func execSystemDBStatements(db *sql.DB) {
	exec(db, "ALTER SYSTEM SET autovacuum TO 'on';")
	exec(db, "ALTER SYSTEM SET autovacuum_max_workers TO '20';")
	exec(db, "ALTER SYSTEM SET autovacuum_vacuum_cost_limit TO '2000';")
	exec(db, "ALTER SYSTEM SET autovacuum_vacuum_cost_delay TO '10ms';")
	exec(db, "ALTER SYSTEM SET checkpoint_completion_target TO '0.8';")
	exec(db, "ALTER SYSTEM SET max_wal_size TO '8GB';")
	exec(db, "ALTER SYSTEM SET checkpoint_timeout TO '30min';")
	exec(db, "ALTER SYSTEM SET constraint_exclusion TO 'partition';")
	exec(db, "ALTER SYSTEM SET default_statistics_target TO '100';")
	exec(db, "ALTER SYSTEM SET effective_cache_size TO '256MB';")
	exec(db, "ALTER SYSTEM SET escape_string_warning TO 'off';")
	exec(db, "ALTER SYSTEM SET log_destination TO 'stderr';")
	exec(db, "ALTER SYSTEM SET log_directory TO 'pg_log';")
	exec(db, "ALTER SYSTEM SET log_filename TO 'postgresql_%a.log';")
	exec(db, "ALTER SYSTEM SET log_line_prefix TO '%m %p ';")
	exec(db, "ALTER SYSTEM SET log_rotation_age TO '1440';")
	exec(db, "ALTER SYSTEM SET log_truncate_on_rotation TO 'on';")
	exec(db, "ALTER SYSTEM SET logging_collector TO 'on';")
	exec(db, "ALTER SYSTEM SET maintenance_work_mem TO '32MB';")
	exec(db, "ALTER SYSTEM SET max_connections TO '300';")
	exec(db, "ALTER SYSTEM SET max_locks_per_transaction TO '256';")
	exec(db, "ALTER SYSTEM SET random_page_cost TO '4.0';")
	exec(db, "ALTER SYSTEM SET shared_buffers TO '1024MB';")
	exec(db, "ALTER SYSTEM SET standard_conforming_strings TO 'off';")
	exec(db, "ALTER SYSTEM SET temp_buffers TO '16MB';")
	exec(db, "ALTER SYSTEM SET work_mem TO '32MB';")
	// db.Close()
}

func dispErr(err error) bool {
	return err != nil
}
