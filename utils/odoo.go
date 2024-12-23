package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
	"gopkg.in/ini.v1"
)

type envConverter func([]string) map[string]string

// GetOdooUser is for future use, so far we do not plan to use other user that odoo to execute the instance
func GetOdooUser() string {
	user := os.Getenv("ODOO_USER")
	if user == "" {
		user = "odoo"
	}
	return user
}

// GetConfigFile will return the odoo config file path that will be used by default
func GetConfigFile(vr valueReader) string {
	fileEnv := vr.readValue("ODOO_CONFIG_FILE")
	if fileEnv != "" {
		return fileEnv
	}
	return "/home/odoo/.odoorc"
}

// GetInstanceType will use by default INSTANCE_TYPE env var because is the one we have been using in DeployV
// for over 5 years, but as Odoo added a similar one we use it too. It is important to notice that the values must match
// for example "updates" and "staging" matchm because are the same stage, but different name
func GetInstanceType(vr valueReader) (string, error) {
	it := vr.readValue("INSTANCE_TYPE")
	ost := vr.readValue("ODOO_STAGE")
	switch {
	case ost == "" && it == "":
		return "", fmt.Errorf("cannot determine the instance type, env vars INSTANCE_TYPE and/or ODOO_STAGE 'must' be defined and match")
	case ost == "":
		return it, nil
	case it == "production" && ost == "production":
		return "production", nil
	case it == "updates" && ost == "staging":
		return "updates", nil
	case it == "develop" && ost == "dev":
		return "develop", nil
	case it == "test" && ost == "staging":
		return "test", nil
	}
	return "", fmt.Errorf("cannot determine the instance type, env vars INSTANCE_TYPE and ODOO_STAGE 'must' match, got: 'INSTANCE_TYPE=%s' and 'ODOO_STAGE=%s'", it, ost)
}

// DefaultConverter receives a slice of strings with all the env vars and returns a mapping with the keys and values
// of those env vars.
func DefaultConverter(list []string) map[string]string {
	return SplitEnvVars(list)
}

// OdoorcConverter receives a slice of strings and filters them by the 'odoorc_' prefix because these are the variables that
// will be replaced in the configuration file, returns them as a map of strings where the key is the key of the
// configuration.
func OdoorcConverter(list []string) map[string]string {
	env_list := SplitEnvVars(list)
	return OdoorcMapConverter(env_list)
}

func OdoorcMapConverter(m map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range m {
		if strings.HasPrefix(strings.ToLower(k), "odoorc_") {
			key := strings.TrimPrefix(strings.ToLower(k), "odoorc_")
			res[key] = v
		}
	}
	return res
}

// SplitEnvVars receives a slice of strings and creates a mapping of strings based on the values in the slice separated
// by `=`. This is used to split the environment variables into a mapping of key: value.
func SplitEnvVars(list []string) map[string]string {
	res := make(map[string]string)
	for _, v := range list {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) < 2 {
			continue
		}
		res[parts[0]] = parts[1]
	}
	return res
}

// FilterStrings receives a slice of strings and filters them using the envConverter provided based on a specific criteria.
func FilterStrings(list []string, converter envConverter) map[string]string {
	return converter(list)
}

// UpdateOdooConfig saves the ini object and updates the addons paths
func UpdateOdooConfig(config *ini.File, vr valueReader) error {
	cfgFile := GetConfigFile(vr)
	if err := config.SaveTo(cfgFile); err != nil {
		return err
	}
	return nil
}

// UpdateSentry check if sentry is enabled in such case adds/updates the values in the ini condiguration file
// setting the environment and the odoo instance path
func UpdateSentry(config *ini.File, instanceType string) {
	if !config.Section("options").HasKey("sentry_enabled") {
		return
	}
	sentryStr := config.Section("options").Key("sentry_enabled").Value()
	isEnabled, err := strconv.ParseBool(sentryStr)
	if err != nil {
		return
	}
	if isEnabled {
		config.Section("options").Key("sentry_odoo_dir").SetValue("/home/odoo/instance/odoo")
		config.Section("options").Key("sentry_environment").SetValue(instanceType)
	}
}

// SetupWorker will update the configuration to match the desired type of container, for example:
// if you wish to run a cron only container set the containerType parameter to cron and this func will disable the
// longpolling and the xmlrpc service
func SetupWorker(config *ini.File, containerType, version string) {
	switch strings.ToLower(containerType) {
	case "worker":
		config.Section("options").Key("odoorc_http_enable").SetValue("True")
		config.Section("options").Key("max_cron_threads").SetValue("0")
		config.Section("options").Key("workers").SetValue("0")
		config.Section("options").Key("xmlrpcs").SetValue("False")

	case "cron":
		config.Section("options").Key("odoorc_http_enable").SetValue("False")
		config.Section("options").Key("max_cron_threads").SetValue("1")
		config.Section("options").Key("workers").SetValue("-1")
		config.Section("options").Key("xmlrpcs").SetValue("False")
		config.Section("options").Key("xmlrpc").SetValue("False")
		if semver.Compare("v"+version, "v16.0") == -1 {
			config.Section("options").Key("workers").SetValue("0")
		}

	case "chat":
		config.Section("options").Key("odoorc_http_enable").SetValue("False")
		config.Section("options").Key("max_cron_threads").SetValue("0")
		config.Section("options").Key("workers").SetValue("2")
		config.Section("options").Key("xmlrpcs").SetValue("False")
	}
}

// UpdateFromVars will update the odoo configuration from env vars which should start with ODOORC_ prefix, if the exists
// the value  will be updated else the parameter will be added to the 'options' section only when appendNew == true,
// which is the default for Odoo.
// If you wish to add it to another section, add a file with only that section to the folder '/external_files/odoocfg'
func UpdateFromVars(config *ini.File, odooVars map[string]string, appendNew bool) {
	sections := config.Sections()
	for k, v := range odooVars {
		k = strings.ToLower(k)
		updated := false
		for _, section := range sections {
			if section.HasKey(k) {
				section.Key(k).SetValue(v)
				updated = true
				break
			}
		}
		// The key does not exist, and we want to force append, so we add it into the options section
		if !updated && appendNew {
			config.Section("options").Key(k).SetValue(v)
		}
	}
}

// SetDefaults takes care of important defaults:
// - Won't allow admin as default superuser password, a random string is generated
// - Won't allow to change the default ports because inside the container is not needed and will mess with the external
// - Disable logrotate since supervisor will handle that
func SetDefaults(config *ini.File, version string) {
	config.Section("options").DeleteKey("longpolling_port")
	config.Section("options").DeleteKey("xmlrpc_port")

	config.Section("options").Key("gevent_port").SetValue("8072")
	config.Section("options").Key("http_port").SetValue("8069")

	if semver.Compare("v"+version, "v16.0") == -1 {
		config.Section("options").Key("longpolling_port").SetValue("8072")
		config.Section("options").Key("xmlrpc_port").SetValue("8069")
	}

	config.Section("options").Key("logrorate").SetValue("False")
	if config.Section("options").Key("admin_passwd").Value() == "admin" ||
		config.Section("options").Key("admin_passwd").Value() == "" {
		config.Section("options").Key("admin_passwd").SetValue(RandStringRunes(64))
	}
}

// Odoo this func coordinates all the odoo configuration loading the config file, calling all the methods needed to
// update the configuration
func Odoo() error {
	log.Info("Preparing the configuration")
	vr, err := GetValueReader()
	if err != nil {
		return err
	}
	if err := prepareFiles(vr); err != nil {
		return err
	}

	log.Info("Setting up the config file")
	odooCfg, err := ini.Load(GetConfigFile(vr))
	if err != nil {
		log.Errorf("Error loading Odoo config: %s", err.Error())
		return err
	}
	store := vr.getDict()
	odooVersion, ok := store["VERSION"]
	if !ok {
		return errors.New("failed to get the Odoo version, please set the env var VERSION with the proper value")
	}

	UpdateFromVars(odooCfg, store, false)
	odooVars := OdoorcMapConverter(store)
	UpdateFromVars(odooCfg, odooVars, true)

	SetupWorker(odooCfg, vr.readValue("ORCHESTSH_CONTAINER_TYPE"), odooVersion)
	instanceType, err := GetInstanceType(vr)
	if err != nil {
		return err
	}
	log.Debugf("Instance type: %s", instanceType)
	UpdateSentry(odooCfg, instanceType)
	SetDefaults(odooCfg, odooVersion)
	if vr.readValue("AUTOSTART") != "" {
		autostart, err := strconv.ParseBool(vr.readValue("AUTOSTART"))
		if err != nil {
			log.Errorf("Error parsing AUTOSTART with value '%s': %s", vr.readValue("AUTOSTART"), err.Error())
			return err
		}
		log.Debugf("Autostart: %v", autostart)

		if autostart {
			if err := UpdateSupervisor("/etc/supervisor/conf.d", SetAutostart); err != nil {
				return err
			}
		}
	}

	log.Debugf("Orchestsh stdout: %v", vr.readValue("ORCHESTSH_STDOUT"))
	if vr.readValue("ORCHESTSH_STDOUT") != "" {
		stdout, err := strconv.ParseBool(vr.readValue("ORCHESTSH_STDOUT"))
		if err != nil {
			log.Errorf("Error parsing ORCHESTSH_STDOUT with value '%s': %s", vr.readValue("ORCHESTSH_STDOUT"), err.Error())
			return err
		}
		if stdout {
			if err := UpdateSupervisor("/etc/supervisor/conf.d", SetStout); err != nil {
				return err
			}
		}
		log.Debugf("Orchestsh stdout: %v", stdout)
	}

	log.Info("Saving new Odoo configuration")
	if err := UpdateOdooConfig(odooCfg, vr); err != nil {
		return err
	}

	return nil
}

func prepareFiles(vr valueReader) error {
	if err := appendFiles(GetConfigFile(vr), "/external_files/odoocfg"); err != nil {
		return err
	}

	fsPath := vr.readValue("CONFIGFILE_PATH")
	if fsPath == "" {
		fsPath = "/home/odoo/.local/share/Odoo/filestore"
	}

	if _, err := os.Stat(fsPath); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(fsPath, 0o777) // #nosec G301
			if err != nil {
				return err
			}
		}
	}

	cmds := []string{
		"chmod ugo+rwxt /tmp",
		"chmod ugo+rw /var/log/supervisor",
		fmt.Sprintf("chown odoo:odoo %s", filepath.Dir(fsPath)),
		fmt.Sprintf("chown odoo:odoo %s", fsPath),
		"chown -R odoo:odoo /home/odoo/.ssh",
	}

	for _, c := range cmds {
		log.Debugf("Running command: %s", c)
		if err := RunAndLogCmdAs(c, "", nil); err != nil {
			log.Errorf("Error running command: %s", err.Error())
			return err
		}
	}
	return nil
}
