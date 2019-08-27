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

package synopsysctl

import (
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// serveUICmd edits Synopsys resources
var serveUICmd = &cobra.Command{
	Use:   "serve-ui",
	Short: "Starts a service running the User Interface and listens for events",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Start Running the ember User Interface on localhost
		log.Debug("Starting User Interface's Ember Front End...")
		r := mux.NewRouter()
		// Serve static files for ember
		log.Debug("Serving static files for ember from /dist")
		static := "../../operator-ui-ember/dist/"
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(static)))

		// Route for base route
		log.Debug("Handling base / route")
		entry := "../../operator-ui-ember/dist/index.html"
		r.PathPrefix("/").HandlerFunc(IndexHandler(entry))

		log.Debug("listening and serving at 8081")
		port := "8081"
		srv := &http.Server{
			Handler: handlers.LoggingHandler(os.Stdout, r),
			Addr:    "localhost:" + port,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())

		// Start Running a backend server that listens for input from the User Interface
		log.Debug("Starting User Interface's Back End")

		return nil
	},
}

func IndexHandler(entry string) func(w http.ResponseWriter, r *http.Request) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, entry)
	}

	return http.HandlerFunc(fn)
}

func init() {
	rootCmd.AddCommand(serveUICmd)
}
