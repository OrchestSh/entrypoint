package utils

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// UpdateAutostart will look for all the services configuration in the supervisor conf.d directory and
// update autostart. If autostart is true nothing will change, if false all values will be set to false
func UpdateAutostart(autostart bool, confPath string) error {
	// if autostart is true we do nothing because should use the default configuration
	if autostart {
		return nil
	}

	files, err := os.ReadDir(confPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".conf") {
			continue
		}
		log.Infof("Checking file: %s", f.Name())
		cfgFile := filepath.Clean(path.Join(confPath, f.Name()))
		content, err := os.ReadFile(cfgFile)
		if err != nil {
			return err
		}
		w, err := os.Create(cfgFile)
		if err != nil {
			return err
		}
		defer w.Close()
		err = setAutostart(content, w)
		if err != nil {
			return err
		}
	}
	return nil
}

// setAutostart will enable autostart in a particular file, but won't touch the default section neither supervisor one
// only in the services section
func setAutostart(content []byte, w io.Writer) error {
	cfg, err := ini.Load(content)
	if err != nil {
		return err
	}
	sections := cfg.Sections()
	for _, section := range sections {
		if section.Name() != "supervisord" && section.Name() != "DEFAULT" {
			section.Key("autostart").SetValue("false")
		}
	}
	if _, err := cfg.WriteTo(w); err != nil {
		return err
	}
	return nil
}
