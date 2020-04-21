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

// NewAppStartCommand returns an Opsani CLI command for starting the app
func NewAppStartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewAPIClientFromConfig()
			if resp, err := client.StartApp(); err == nil {
				return PrettyPrintJSONResponse(resp)
			} else {
				return err
			}
		},
	}
}

// NewAppStopCommand returns an Opsani CLI command for stopping the app
func NewAppStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewAPIClientFromConfig()
			resp, err := client.StopApp()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

// NewAppRestartCommand returns an Opsani CLI command for restarting the app
func NewAppRestartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewAPIClientFromConfig()
			resp, err := client.RestartApp()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}

// NewAppStatusCommand returns an Opsani CLI command for retrieving status on the app
func NewAppStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check app status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := NewAPIClientFromConfig()
			resp, err := client.GetAppStatus()
			if err != nil {
				return err
			}
			return PrettyPrintJSONResponse(resp)
		},
	}
}
