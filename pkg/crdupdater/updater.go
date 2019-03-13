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

package crdupdater

import (
	"github.com/juju/errors"
)

// AddOrRemoveComponents consist of methods to add or remove the components for update events
type AddOrRemoveComponents interface {
	get(name string) (interface{}, error)
	add() error
	list() (interface{}, error)
	delete(name string) error
	remove() error
}

// Updater handles in updating the components
type Updater struct {
	updaters []AddOrRemoveComponents
}

// NewUpdater will create the specification that is used for updating the components
func NewUpdater() *Updater {
	updater := Updater{
		updaters: make([]AddOrRemoveComponents, 0),
	}
	return &updater
}

// AddUpdater will add the updater to the list
func (u *Updater) AddUpdater(updater AddOrRemoveComponents) {
	u.updaters = append(u.updaters, updater)
}

// Update add or remove the components
func (u *Updater) Update() error {
	for _, updater := range u.updaters {
		err := updater.add()
		if err != nil {
			return errors.Annotatef(err, "add components:")
		}
		err = updater.remove()
		if err != nil {
			return errors.Annotatef(err, "remove components:")
		}
	}
	return nil
}
