/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package update

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "applies all the secrets in input-file.  does not delete old secrets. Prompts on empty secrets.",
	Long: ` 
	devsecrets update --all | --name <name> --verbose --input-file dev-secrets.json
    
    `,
	Run: func(cmd *cobra.Command, args []string) {
		OnUpdate()
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// repoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// repoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
