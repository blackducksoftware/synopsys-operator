/*
 * Copyright (C) 2020 Synopsys, Inc.
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

package polarisreporting

// SetHelmValueInMap adds the finalValue into the valueMapPointer at the location specified
// by the keyList
// valueMapPointer - a map for helm values (maps are pointers in Golang)
//  - it is used to track the current map being updated
// keyList - an ordered list of keys that lead to the location in the valueMapPointer to place the finalValue
// finalValue - the value to set in the map
func SetHelmValueInMap(valueMapPointer map[string]interface{}, keyList []string, finalValue interface{}) {
	for i, currKey := range keyList {
		if i == (len(keyList) - 1) { // at the last key -> set the value
			valueMapPointer[currKey] = finalValue
			return
		}
		if nextMap, _ := valueMapPointer[currKey]; nextMap != nil { // key is in map -> go to next map
			valueMapPointer = nextMap.(map[string]interface{})
		} else { // key is not in the map -> add the key and next key; go to next map
			nextKey := keyList[i+1]
			valueMapPointer[currKey] = map[string]interface{}{nextKey: nil}
			nextMap := valueMapPointer[currKey].(map[string]interface{})
			valueMapPointer = nextMap
		}
	}
}
