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
	"github.com/halng/deto/tui"
	"github.com/spf13/cobra"
	"os"
)

// ManCmd represents the man command
var manCmd = &cobra.Command{
	Use:   "man",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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

		fmt.Printf("Action: %s for candidate: %s", actionType, candidate)

	},
}

func init() {
	rootCmd.AddCommand(manCmd)
	manCmd.Flags().StringP("action", "a", "", "Action name. [install|remove|list|default]")
	manCmd.Flags().StringP("candidate", "c", "", "Candidate name")
}
