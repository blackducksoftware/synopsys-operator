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
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	// "os"
	// "time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Port to serve on
var serverPort = "8081"

// serveUICmd edits Synopsys resources
var serveUICmd = &cobra.Command{
	Use:   "serve-ui",
	Short: "Starts a service running the User Interface and listens for events",
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Debug("Starting User Interface Server...")

		// Start Running a backend server that listens for input from the User Interface
		router := mux.NewRouter()

		// api route - deploy Polaris
		router.HandleFunc("/api/deploy_polaris", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Handler Deploy Polaris: %q\n", html.EscapeString(r.URL.Path))
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Data from Polaris Body: %s\n", reqBody)
		})

		// api route - deploy Black Duck
		router.HandleFunc("/api/deploy_black_duck", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Handler Deploy Black Duck: %q\n", html.EscapeString(r.URL.Path))
			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Data from Black Duck Body: %s\n", reqBody)
		})

		// Serve files for UI
		spa := spaHandler{
			staticPath: "../../operator-ui-ember/dist",
			indexPath:  "../../operator-ui-ember/dist/index.html",
		}
		router.PathPrefix("/").Handler(spa)

		fmt.Printf("==================================\n")
		fmt.Printf("Serving at: http://localhost:%s\n", serverPort)
		fmt.Printf("api:\n  - /api/deploy_polaris\n  - /api/deploy_black_duck\n")
		fmt.Printf("==================================\n")
		fmt.Printf("\n")

		// Serving the server
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", serverPort), handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "Content-type", "Authorization"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(router)))

		return nil
	},
}

// Copied From: https://github.com/gorilla/mux#serving-single-page-applications
// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// Copied From: https://github.com/gorilla/mux#serving-single-page-applications
// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

func init() {
	rootCmd.AddCommand(serveUICmd)
	serveUICmd.Flags().StringVarP(&serverPort, "port", "p", serverPort, "Port to listen for UI requests")
}
