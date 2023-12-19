package cou

import (
	"os"

	"entrypoint/utils"
	log "github.com/sirupsen/logrus"
)

// RunCou will execute click-odoo-update to update the database, if you want to save the logs pass stream=false else the
// logs will be streamed to the stdout
func RunCou(configFile, db_name string, stream bool) error {
	cmd := "click-odoo-update -c " + configFile
	if db_name != "" {
		cmd += " -d " + db_name
	}
	if !stream {
		cmd += " --logfile /tmp/deployvlogs/cou_update.log"
	}
	pythonPath := "/home/odoo/instance/odoo:" + os.Getenv("PYTHONPATH")
	log.Debugf("Running command: %s", cmd)
	err := utils.RunAndLogCmdAs(cmd, "odoo", []string{pythonPath})
	return err
}
