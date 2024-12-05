package cmd

import (
	"os"

	"entrypoint/utils"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a command as the given user",
	Long: `This is not a replacement for sudo, is just a wrapper to configure odoo and then run a command
as the given user.

The command must be between single quotes:

# /entrypoint run --user odoo '/home/odoo/instance/odoo/odoo-bin -c /home/odoo/.odoorc -u all --stop-after-init'
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Setting up Odoo")
		if err := utils.Odoo(); err != nil {
			log.Errorf("Error setting up Odoo: %s", err.Error())
			os.Exit(1)
		}
		user, err := cmd.Flags().GetString("user")
		if err != nil {
			log.Errorf("Error getting user name: %s", err.Error())
			os.Exit(1)
		}
		log.Debugf("Running command as user: %s", user)
		if len(args) == 0 {
			log.Error("Error reading the args, at leas one is required")
			os.Exit(1)
		}
		err = utils.RunAndLogCmdAs(args[0], user, nil)
		if err != nil {
			log.Errorf("Error executing the command: %s", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.PersistentFlags().String("user", "", "User that will execute the command")
	_ = runCmd.MarkFlagRequired("user")
}
