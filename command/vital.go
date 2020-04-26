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
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

type vitalCommand struct {
	*BaseCommand
}

// NewVitalCommand returns a new instance of the vital command
func NewVitalCommand(baseCmd *BaseCommand) *cobra.Command {
	vitalCommand := vitalCommand{BaseCommand: baseCmd}
	cobraCmd := &cobra.Command{
		Use:               "vital",
		Short:             "Start optimizing",
		Args:              cobra.NoArgs,
		PersistentPreRunE: nil,
		RunE:              vitalCommand.RunVital,
	}

	return cobraCmd
}

func (vitalCommand *vitalCommand) RunVital(cobraCmd *cobra.Command, args []string) error {
	in := `# Opsani Vital
	
	This is a simple example of glamour!
	Check out the [other examples](https://github.com/charmbracelet/glamour/tree/master/examples).
	
	Bye!
	`

	out, _ := glamour.Render(in, "dark")
	fmt.Print(out)
	return nil
}
