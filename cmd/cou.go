package cmd

import (
	"entrypoint/pkg/cou"
	"entrypoint/utils"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// couCmd represents the cou command
var couCmd = &cobra.Command{
	Use:   "cou",
	Short: "Update the given database using click-odoo-update",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Setting up Odoo")
		if err := utils.Odoo("", false); err != nil {
			log.Errorf("Error setting up Odoo: %s\n", err.Error())
			os.Exit(1)
		}

		filename := utils.GetConfigFile()
		db_name, err := cmd.Flags().GetString("db_name")
		if err != nil {
			log.Errorf("Error getting database name: %s", err.Error())
			os.Exit(1)
		}

		err = cou.RunCou(filename, db_name, false)
		if err != nil {
			log.Errorf("Error running click-odoo-update: %s", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(couCmd)
	couCmd.PersistentFlags().String("db_name", "", "Database name to be used in the update")
	_ = couCmd.MarkFlagRequired("db_name")
}
