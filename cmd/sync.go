package cmd

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"strings"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronizes local changes with changes on origin or specified branch",
	Long: `sync
sync origin master
sync local-branch
`,
	Run: func(cmd *cobra.Command, args []string) {
		RunInTerminalWithColor("git", []string{"fetch"})

		// if possibly squashed
		if IsDiverged() {
			RunInTerminalWithColor("git", []string{"status", "-sb", "--untracked-files=no"})

			ans := ""
			prompt := &survey.Select{
				Message: "Branch is diverged from origin/upstream – handle by...",
				Options: []string{
					"Rebase on origin/upstream",
					"Force (destructive) push to origin/" + CurrentBranch(),
					"Cancel"},
			}
			survey.AskOne(prompt, &ans)
			if strings.HasPrefix(ans, "Rebase") {
				RunInTerminalWithColor("git", []string{"pull", "-r"})
			} else if strings.HasPrefix(ans, "Force") {
				RunInTerminalWithColor("git", []string{"push", "--force-with-lease"})
			} else {
				fmt.Println("Canceling..")
			}
			return
		}
		if !CloudBranchExists() {
			RunInTerminalWithColor("git", []string{"push", "--set-upstream", "origin", CurrentBranch()})
			save("")
			RunInTerminalWithColor("git", []string{"push"})
			return
		}
		save("")
		if !CloudBranchExists() {
			RunInTerminalWithColor("git", []string{"push", "--set-upstream", "origin", CurrentBranch()})
		}
		if IsAheadOfCurrent() {
			RunInTerminalWithColor("git", []string{"push"})
		} else {
			RunInTerminalWithColor("git", []string{"pull", "-r"})
			if len(args) > 0 {
				RunInTerminalWithColor("git", append([]string{"pull", "-r"}, args...))
			}
			RunInTerminalWithColor("git", []string{"push"})
		}

		// After syncing with current branch and user wants to sync with another branch

		if CurrentBranch() == "master" && len(args) == 1 && strings.HasSuffix(args[0], "master") {
			yes := AskConfirm("Squash & merge this branch into master?")

			if yes {
				RunInTerminalWithColor("git", []string{"merge", "--squash"})
				return
			}
			fmt.Printf("Cancelling...")
			//RunInTerminalWithColor("git", []string{"stash", "pop"}) deprecated switch stashing
			return
		}

		if len(args) == 1 {
			branch := args[0]
			refreshOnBranch(branch)
		}

	},
	//Args: cobra.MaximumNArgs(1),
}

func init() {
	ShellCmd.AddCommand(syncCmd)
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
