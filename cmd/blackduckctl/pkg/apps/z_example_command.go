package apps

// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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

import (
	"fmt"

	"github.com/spf13/cobra"
)

var wtfCommand = &cobra.Command{
	Use:   "example",
	Short: "blackduckctl example command",
	Long:  `This is an example of how to add a command to blackduckctl.  Interactive support has to be hardwired.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Example command  called ! Good job.")
		if wtf, _ := cmd.Flags().GetBool("wtf"); wtf == true {
			fmt.Println("\t You wanted me to complain ? ok. WTF !")
		}
	},
}

// implementing init is important ! thats how cobra knows to bind your 'app' to top level command.
func init() {
	RootCmd.AddCommand(wtfCommand)

	// specific flags for your app, add them here...
	wtfCommand.Flags().BoolP("wtf", "w", false, "print 'wtf' as part of the example of how to issue subcommands.")
}
