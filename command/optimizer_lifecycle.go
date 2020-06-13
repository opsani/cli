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

import "github.com/spf13/cobra"

// NewOptimizerStartCommand returns an Opsani CLI command for starting the app
func NewOptimizerStartCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			if resp, err := client.StartApp(); err == nil {
				return PrettyPrintJSONResponse(resp)
			} else {
				return err
			}
		},
	}
}

// NewOptimizerStopCommand returns an Opsani CLI command for stopping the app
func NewOptimizerStopCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			resp, err := client.StopApp()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

// NewOptimizerRestartCommand returns an Opsani CLI command for restarting the app
func NewOptimizerRestartCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			resp, err := client.RestartApp()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

// NewOptimizerStatusCommand returns an Opsani CLI command for retrieving status on the app
func NewOptimizerStatusCommand(baseCmd *BaseCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check app status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := baseCmd.NewAPIClient()
			resp, err := client.GetAppStatus()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}
