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

package synopsysctl

import (
	"bytes"
	"io"
	"testing"
)

func TestAskYesNoWithDefault(t *testing.T) {
	type args struct {
		question  func() error
		answer    string
		alwaysYes bool
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "alwaysYes",
			args: args{
				question: func() error {
					return nil
				},
				alwaysYes: true,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "no",
			args: args{
				question: func() error {
					return nil
				},
				answer:    "no",
				alwaysYes: false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "yes",
			args: args{
				question: func() error {
					return nil
				},
				answer:    "yes",
				alwaysYes: false,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "No",
			args: args{
				question: func() error {
					return nil
				},
				answer:    "No",
				alwaysYes: false,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "yEs",
			args: args{
				question: func() error {
					return nil
				},
				answer:    "yEs",
				alwaysYes: false,
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			if !tt.args.alwaysYes {

				if _, err := io.WriteString(&buf, tt.args.answer); err != nil {
					t.Error(err)
				}
			}
			got, err := AskYesNoWithDefault(tt.args.question, tt.args.alwaysYes, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("AskYesNoWithDefault() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AskYesNoWithDefault() got = %v, want %v", got, tt.want)
			}
		})
	}
}
