// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package synopsysctl

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ResourceCtl interface defines functions that other
// ctl types for resources should define
type ResourceCtl interface {
	GetSpec() struct{}               // returns spec for the resource
	SwitchSpec(string) error         // change the spec for the resource
	AddSpecFlags(cmd *cobra.Command) // Add flags for the resource spec
	SetFlags(f *pflag.Flag)          // update the spec with flags that changed
}
