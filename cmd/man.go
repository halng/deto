package cmd

/*
Copyright © 2024 Hal Ng <haonguyentan2001@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

import (
	"fmt"
	"github.com/halng/deto/pkg"
	"github.com/halng/deto/tui"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

// ManCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "A Package Manager for Developers",
	Long: `Deto is a package manager for developers. It helps you to install, remove, list, and set default versions of development tools.
Available candidates: java, go
Available actions: install, remove, list, default
Let's combine them to use deto. 
For example: deto man
There will be a prompt to ask you to choose the candidate and action type. You just need to follow the instructions.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		tui.Clear()

		actionType, err := cmd.Flags().GetString("action")
		if err != nil {
			fmt.Println("There was an error getting the action type.", err.Error())
			os.Exit(1)
		}

		if actionType == "" {
			options := []string{
				"install",
				"remove",
				"list",
				"default",
			}

			actionType = tui.GetChoice(options)
		}
		tui.Clear()

		candidate, err := cmd.Flags().GetString("candidate")
		if err != nil {
			fmt.Println("There was an error getting the candidate type.", err.Error())
			os.Exit(1)
		}
		if candidate == "" {
			candidate = tui.Input("Enter the candidate name: ")
		}

		tui.Clear()

		os := runtime.GOOS
		arch := runtime.GOARCH

		var man = pkg.Man{
			Candidate:       candidate,
			ActionType:      actionType,
			OperatingSystem: os,
			Architecture:    arch,
		}

		man.Handler()
	},
}

func init() {
	rootCmd.AddCommand(manCmd)
	manCmd.Flags().StringP("action", "a", "", "Action name. [install|remove|list|default]")
	manCmd.Flags().StringP("candidate", "c", "", "Candidate name")
}
