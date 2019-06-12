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
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion SHELL",
	Short: "Output shell completion code for the specified shell (bash or zsh)",
	Long: `Output shell completion code for the specified shell (bash or zsh).
	The shell code must be evaluated to provide interactive
	completion of synopsysctl commands.  This can be done by sourcing it from
	the .bash_profile.
	Detailed instructions on how to do this are available here:
	https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion
	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2`,
	Example: `
	# Installing bash completion on macOS using homebrew
	## If running Bash 3.2 included with macOS
			brew install bash-completion
	## or, if running Bash 4.1+
			brew install bash-completion@2
	## You may need add the completion to your completion directory
			synopsysctl completion bash > $(brew --prefix)/etc/bash_completion.d/synopsysctl
	# Installing bash completion on Linux
	## If bash-completion is not installed on Linux, please install the 'bash-completion' package
	## via your distribution's package manager.
	## Load the synopsysctl completion code for bash into the current shell
			source <(synopsysctl completion bash)
	## Write bash completion code to a file and source if from .bash_profile
			synopsysctl completion bash > ~/.synopsysctl/completion.bash.inc
			printf "
				# synopsysctl shell completion
				source '$HOME/.synopsysctl/completion.bash.inc'
				" >> $HOME/.bash_profile
			source $HOME/.bash_profile
	# Load the synopsysctl completion code for zsh[1] into the current shell
			source <(synopsysctl completion zsh)
	# Set the synopsysctl completion code for zsh[1] to autoload on startup
			synopsysctl completion zsh > "${fpath[1]}/_synopsysctl"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("shell not specified")
		}
		if len(args) > 1 {
			return fmt.Errorf("too many arguments. Expected only the shell type")
		}
		shell := args[0]
		if strings.EqualFold(shell, "bash") {
			rootCmd.GenBashCompletion(os.Stdout)
		} else if strings.EqualFold(shell, "zsh") {
			rootCmd.GenZshCompletion(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
