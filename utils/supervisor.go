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

// UpdateSupervisor will look for all the services configuration in the supervisor conf.d directory and
// update the configuration accordingly applying the setter function.
func UpdateSupervisor(confPath string, setter func(content []byte, w io.Writer) error) error {

	files, err := os.ReadDir(confPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".conf") {
			continue
		}
		log.Infof("Checking file: %s", f.Name())
		cfgFile := path.Join(confPath, f.Name())
		if err2 := setConfig(cfgFile, setter); err2 != nil {
			return err2
		}
	}
	return nil
}

func setConfig(cfgFile string, setter func(content []byte, w io.Writer) error) error {
	content, err := os.ReadFile(filepath.Clean(cfgFile))
	if err != nil {
		return err
	}
	w, err := os.Create(filepath.Clean(cfgFile))
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil {
			log.Errorf("Failed to close file descriptor: %v", err)
		}
	}()
	err = setter(content, w)
	if err != nil {
		return err
	}
	return nil
}

// SetAutostart will enable autostart in a particular file, but won't touch the default section neither supervisor one
// only in the services section
func SetAutostart(content []byte, w io.Writer) error {
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

// SetStout will enable stdout_logfile and stderr_logfile in a particular configuration file so all the output will be
// sent to the standard output instead of a file as usually is set
func SetStout(content []byte, w io.Writer) error {
	cfg, err := ini.Load(content)
	if err != nil {
		return err
	}
	for _, section := range cfg.Sections() {
		if section.Name() == "supervisord" || section.Name() == "DEFAULT" {
			continue
		}
		section.Key("stdout_logfile").SetValue("/dev/fd/1")
		section.Key("stderr_logfile").SetValue("/dev/fd/1")
		section.Key("stdout_logfile_maxbytes").SetValue("0")
		section.Key("stderr_logfile_maxbytes").SetValue("0")
	}
	if _, err := cfg.WriteTo(w); err != nil {
		return err
	}
	return nil
}
