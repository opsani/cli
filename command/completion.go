// Copyright 2020 Opsani
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package command

import (
	"errors"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/opsani/cli/internal/cobrafish"
	"github.com/spf13/cobra"
)

// NewCompletionCommand returns a new Opsani CLI cmpletion command instance
func NewCompletionCommand(baseCmd *BaseCommand) *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for Opsani CLI commands.

		The output of this command will be computer code and is meant to be saved to a
		file or immediately evaluated by an interactive shell.

		For example, for bash you could add this to your '~/.bash_profile':

			eval "$(gh completion -s bash)"

		When installing Opsani CLI through a package manager, however, it's possible that
		no additional shell configuration is necessary to gain completion support. For
		Homebrew, see <https://docs.brew.sh/Shell-Completion>
		 `,
		PersistentPreRunE: nil,
		RunE: func(cmd *cobra.Command, args []string) error {
			shellType, err := cmd.Flags().GetString("shell")
			if err != nil {
				return err
			}

			if shellType == "" {
				out := cmd.OutOrStdout()
				isTTY := false
				if outFile, isFile := out.(*os.File); isFile {
					isTTY = IsTerminal(outFile)
				}

				if isTTY {
					return errors.New("error: the value for `--shell` is required\nsee `opsani help completion` for more information")
				}
				shellType = "bash"
			}

			switch shellType {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "powershell":
				return cmd.Root().GenPowerShellCompletion(cmd.OutOrStdout())
			case "fish":
				return cobrafish.GenCompletion(cmd.Root(), cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell type %q", shellType)
			}
		},
	}

	completionCmd.Flags().StringP("shell", "s", "", "Shell type: {bash|zsh|fish|powershell}")

	return completionCmd
}

// IsTerminal reports whether the file descriptor is connected to a terminal
func IsTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
