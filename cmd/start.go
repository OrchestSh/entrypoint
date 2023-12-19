package cmd

import (
	"os"

	"entrypoint/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts supervisor",
	Long: `Will start supervisor in consequence all the configured processes that are marked for autostart.

If the env var AUTOSTART is set to false no process should be started, only supervisor itself (and the entrypoint of course)`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Setting up Odoo")
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("error retrieving file flag: %v", err)
		}
		useDockerSecrets, err := cmd.Flags().GetBool("docker-secrets")
		if err != nil {
			if err != nil {
				log.Fatalf("error retrieving docker-secrets flag: %v", err)
			}
		}
		if err := utils.Odoo(file, useDockerSecrets); err != nil {
			log.Fatalf("Error setting up Odoo: %s", err.Error())
		}

		err = utils.RunAndLogCmdAs("supervisord -c /etc/supervisor/supervisord.conf", "", nil)
		if err != nil {
			log.Errorf("Error starting supervisor: %s", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	startCmd.Flags().String("file", "", "describes a file to read key=value pairs to include in the configuration")
	startCmd.Flags().Bool("docker-secrets", false, "indicates if secrets at /run/secrets should be included in the values to update configuration")

	rootCmd.AddCommand(startCmd)
}
