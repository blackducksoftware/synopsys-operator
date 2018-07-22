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

package webservice

import (
	"github.com/blackducksoftware/perceptor-protoform/pkg/hub"
	"github.com/blackducksoftware/perceptor-protoform/pkg/model"
	gin "github.com/gin-gonic/gin"

	log "github.com/sirupsen/logrus"
)

// all the core logic is here...

func SetupHTTPServer() {
	go func() {
		// data, err := ioutil.ReadFile("/public/index.html")
		// Set the router as the default one shipped with Gin
		router := gin.Default()

		// prints debug stuff out.
		router.Use(GinRequestLogger())

		router.POST("/hub", func(c *gin.Context) {
			log.Debug("create hub request")
			request := &model.CreateHubRequest{}
			if err := c.BindJSON(request); err != nil {
				log.Debugf("Fatal failure binding the incoming request ! %v", c.Request)
			}

			// log.Debugf("[begin] Attempting to get hub now... chekcing if %v in %v", request.Namespace, cmc.GetModel())

			// if _, ok := cmc.GetModel().Hubs[request.Namespace]; ok {
			// 	c.JSON(500, fmt.Errorf("{\"message\":\"namespace %s already in use\"}", request.Namespace))
			// 	return
			// }

			log.Debug("...Attempting to get hub now [done]")

			createHub := &model.CreateHub{
				Namespace:        request.Namespace,
				DockerRegistry:   request.DockerRegistry,
				DockerRepo:       request.DockerRepo,
				HubVersion:       request.HubVersion,
				Flavor:           request.Flavor,
				AdminPassword:    request.AdminPassword,
				UserPassword:     request.UserPassword,
				PostgresPassword: request.PostgresPassword,
				IsRandomPassword: request.IsRandomPassword,
			}

			log.Debug("making a possibly blocking call to create hub now !!!")
			go func() {
				hubCreater := hub.NewHubCreater()
				hubCreater.CreateHub(createHub)
			}()

			c.JSON(200, "\"message\": \"Succeeded\"")
		})

		router.DELETE("/hub", func(c *gin.Context) {
			var request *model.DeleteHubRequest = &model.DeleteHubRequest{}

			c.BindJSON(request)
			log.Debugf("delete hub request %v", request.Namespace)

			// This is on the event loop.
			go func() {
				hubCreater := hub.NewHubCreater()
				hubCreater.DeleteHub(request)
			}()

			c.JSON(200, "\"message\": \"Succeeded\"")
		})

		// Start and run the server - blocking call, obviously :)
		router.Run(":80")
	}()
}
