/*
 * Copyright (C) 2019 Synopsys, Inc.
 *
 *  Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 *  with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 *  under the License.
 */

package utils

import (
	"testing"

	// log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestRandomStringGenerator will test the random string generator
func TestRandomStringGenerator(t *testing.T) {
	rand, err := GetRandomString(32)
	// log.Infof("random string: %s", rand)
	if err != nil {
		t.Errorf("unable to get the random string due to %+v", err)
	}
	assert.Equal(t, 32, len(rand), "random string length not matching")
}
