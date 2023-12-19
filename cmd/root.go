package cmd

import (
	"fmt"
	"os"

	easy "github.com/ruiztulio/logrus-easy-formatter"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "entrypoint",
	Short: "Entry point for container with Odoo",
	Long: `This command was designed to be used in the containers running Odoo, if the start subcommand is used
will do the best to update the odoo configuration file with the env vars`,
	PersistentPreRun: setup,
}

func setup(_ *cobra.Command, _ []string) {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "%time% %lvl% %msg%\n",
	})
	if os.Getenv("DEBUG_ENTRYPOINT") != "" {
		log.SetLevel(log.DebugLevel)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}
