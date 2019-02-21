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

package apps

import (
	"database/sql"
	"fmt"

	// This is required to access the Postgres database
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

// Database will provide the postgres database configuration
type Database struct {
	Connection *sql.DB
}

// NewDatabase will create a database connection and provide the connection instance
func NewDatabase(host string, name string, user string, password string, driverName string) (*Database, error) {
	// Note that sslmode=disable is required it does not mean that the connection
	// is unencrypted. All connections via the proxy are completely encrypted.
	log.Debug("attempting to open database connection")
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable connect_timeout=10", host, name, user, password)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	log.Debug("connected to database")
	return &Database{Connection: db}, nil
}

// ExecuteStatements will list of statements
func (d *Database) ExecuteStatements(statements []string) []error {
	var errs []error
	for _, statement := range statements {
		_, err := d.Connection.Exec(statement)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// CloseDatabaseConnection will close the database connection
func (d *Database) CloseDatabaseConnection() error {
	return d.Connection.Close()
}
