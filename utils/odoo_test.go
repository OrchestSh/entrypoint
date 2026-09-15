package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

func TestFilterStrings(t *testing.T) {
	executed := false
	_ = FilterStrings([]string{}, func([]string) map[string]string {
		executed = true
		return nil
	})
	if !assert.True(t, executed) {
		t.Errorf("The function FilterStrings was not called.")
	}
}

func TestGetMainRepoPath(t *testing.T) {
	values := []struct {
		mainRepo string
		expected string
	}{
		{"extra_addons/customer", "/home/odoo/instance/extra_addons/customer"},
		{"", "/home/odoo/instance/odoo"},
	}
	for _, v := range values {
		t.Setenv("MAIN_REPO_PATH", v.mainRepo)
		assert.Equal(t, v.expected, GetMainRepoPath(),
			"MAIN_REPO_PATH must be resolved under the instance directory.")
	}
}

func TestGetMainRepoBranch(t *testing.T) {
	repoPath := t.TempDir()
	if err := exec.Command("git", "init", "-b", "18.0-dev1", repoPath).Run(); err != nil {
		t.Skipf("git is not available: %s", err)
	}
	assert.Equal(t, "18.0-dev1", GetMainRepoBranch(repoPath),
		"The branch of the repository must be reported.")
	assert.Equal(t, "", GetMainRepoBranch(t.TempDir()),
		"A directory without a repository must report no branch.")
}

func TestGetImageTag(t *testing.T) {
	t.Setenv("MAIN_APP", "customer")
	t.Setenv("VERSION", "19.0")
	repoPath := t.TempDir()
	if err := exec.Command("git", "init", "-b", "19.0", repoPath).Run(); err != nil {
		t.Skipf("git is not available: %s", err)
	}
	if err := exec.Command("git", "-C", repoPath, "-c", "user.email=a@b", "-c", "user.name=a",
		"commit", "--allow-empty", "-m", "empty").Run(); err != nil {
		t.Fatal(err)
	}
	commit := RunGit(repoPath, "rev-parse", "--short", "HEAD")
	assert.Equal(t, "customer-19.0-"+commit, GetImageTag(repoPath),
		"The tag must be built from MAIN_APP, VERSION and the commit.")
	assert.Equal(t, "", GetImageTag(t.TempDir()),
		"Without a repository there is no tag to report.")
	t.Setenv("MAIN_APP", "")
	assert.Equal(t, "", GetImageTag(repoPath),
		"Without MAIN_APP there is no tag to report.")
}

func TestGetSentryEnvironment(t *testing.T) {
	// The main repository does not exist while the tests run, so the branch lookup finds nothing
	// and the environment falls back to VERSION.
	values := []struct {
		instanceType string
		version      string
		expected     string
	}{
		{"production", "19.0", "production-19.0"},
		{"production", "18.0", "production-18.0"},
		{"develop", "", "develop"},
	}
	for _, v := range values {
		t.Setenv("VERSION", v.version)
		assert.Equal(t, v.expected, GetSentryEnvironment(v.instanceType),
			"The environment must join the stage and the version.")
	}
}

func TestSentryReadsOwnSection(t *testing.T) {
	loader := filepath.Join(t.TempDir(), "config.py")
	if err := os.WriteFile(loader, []byte("self._log(logging.WARNING, \"unknown option %r\")\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assert.True(t, SentryReadsOwnSection(loader),
		"The marker in the config loader means the module reads its own section.")

	older := filepath.Join(t.TempDir(), "config.py")
	if err := os.WriteFile(older, []byte("def _load_file_options(self, rcfile):\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	assert.False(t, SentryReadsOwnSection(older),
		"Without the marker the keys belong in [options] alone.")
	assert.False(t, SentryReadsOwnSection(filepath.Join(t.TempDir(), "missing.py")),
		"An unreadable loader must not add the section.")
}

func TestMoveSentryKeysToOwnSection(t *testing.T) {
	config := ini.Empty()
	options := config.Section("options")
	options.Key("sentry_enabled").SetValue("True")
	options.Key("sentry_odoo_dir").SetValue("/home/odoo/instance/extra_addons/customer")
	options.Key("server_wide_modules").SetValue("web,queue_job,sentry")
	options.Key("db_host").SetValue("localhost")

	MoveSentryKeysToOwnSection(config)

	sentry := config.Section("sentry")
	assert.Equal(t, "True", sentry.Key("sentry_enabled").Value(),
		"Every sentry_ key must reach the section the module reads.")
	assert.Equal(t, "/home/odoo/instance/extra_addons/customer", sentry.Key("sentry_odoo_dir").Value(),
		"Every sentry_ key must reach the section the module reads.")
	assert.False(t, options.HasKey("sentry_enabled"),
		"A key Odoo does not know must not stay in [options], it warns on every start.")
	assert.False(t, options.HasKey("sentry_odoo_dir"),
		"A key Odoo does not know must not stay in [options], it warns on every start.")
	assert.Equal(t, "web,queue_job,sentry", options.Key("server_wide_modules").Value(),
		"server_wide_modules is an Odoo option and stays in [options].")
	assert.Equal(t, "localhost", options.Key("db_host").Value(),
		"Only the sentry_ prefix moves.")
	assert.False(t, sentry.HasKey("db_host"), "Only the sentry_ prefix moves.")
}

func TestSentryValue(t *testing.T) {
	config := ini.Empty()
	config.Section("options").Key("sentry_enabled").SetValue("True")
	assert.Equal(t, "True", SentryValue(config, "sentry_enabled"),
		"A key that is still in [options] must be found there.")

	MoveSentryKeysToOwnSection(config)
	assert.Equal(t, "True", SentryValue(config, "sentry_enabled"),
		"A key already moved to [sentry] must be found there, which is the state of every start "+
			"after the first one.")
	assert.Equal(t, "", SentryValue(config, "sentry_dsn"),
		"A key that lives in neither section reads empty.")
}

func TestUpdateSentryAfterTheKeysMoved(t *testing.T) {
	// What a restart looks like: the entry point reloads its own output, so sentry_enabled is in
	// [sentry] and a variable added since then landed in [options] as a new key.
	config := ini.Empty()
	config.Section("sentry").Key("sentry_enabled").SetValue("True")
	config.Section("options").Key("sentry_dsn").SetValue("https://key@sentry.example.com/1")

	UpdateSentry(config, "production")

	assert.Equal(t, GetMainRepoPath(), SentryValue(config, "sentry_odoo_dir"),
		"The block must run even though sentry_enabled is no longer in [options].")
	assert.Equal(t, GetSentryEnvironment("production"), SentryValue(config, "sentry_environment"),
		"The environment must be recomputed on every start, it carries the branch.")
	assert.Equal(t, "https://key@sentry.example.com/1", SentryValue(config, "sentry_dsn"),
		"A variable that arrived after the first start must still be reachable.")
}

func TestSplitEnvVars(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "111", "odoorc_var": "123", "CAPS_VAR": "wer", "ODOORC_INCAPS": "caps"},
		},
	}
	for _, v := range values {
		res := SplitEnvVars(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestDefaultConverter(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "111", "odoorc_var": "123", "CAPS_VAR": "wer", "ODOORC_INCAPS": "caps"},
		},
	}
	for _, v := range values {
		res := DefaultConverter(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestOdoorcConverter(t *testing.T) {
	values := []struct {
		input    []string
		expected map[string]string
	}{
		{
			[]string{"var=111", "odoorc_var=123", "CAPS_VAR=wer", "ODOORC_INCAPS=caps"},
			map[string]string{"var": "123", "incaps": "caps"},
		},
	}
	for _, v := range values {
		res := OdoorcConverter(v.input)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestGetOdooUser(t *testing.T) {
	res := GetOdooUser()
	assert.Equal(t, "odoo", res)
}

func TestGetConfigFile(t *testing.T) {
	vr, err := GetValueReader()
	if err != nil {
		t.Error(err)
	}
	res := GetConfigFile(vr)
	assert.Equal(t, "/home/odoo/.odoorc", res)
	err = os.Setenv("ODOO_CONFIG_FILE", "/etc/odoo.conf")
	assert.NoError(t, err)
	res = GetConfigFile(vr)
	assert.Equal(t, "/etc/odoo.conf", res)
	err = os.Unsetenv("ODOO_CONFIG_FILE")
	assert.NoError(t, err)
}

func TestGetInstanceType(t *testing.T) {
	vr, err := GetValueReader()
	if err != nil {
		t.Error(err)
	}

	_, err = GetInstanceType(vr)
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and/or ODOO_STAGE 'must' be defined and match")

	err = os.Setenv("INSTANCE_TYPE", "test")
	assert.NoError(t, err)
	res, err := GetInstanceType(vr)
	assert.NoError(t, err)
	assert.Equal(t, "test", res)

	err = os.Setenv("ODOO_STAGE", "dev")
	assert.NoError(t, err)
	_, err = GetInstanceType(vr)
	assert.Errorf(t, err, "cannot determine the instance type, env vars INSTANCE_TYPE and ODOO_STAGE 'must' match")
}

//func TestUpdateSentry(t *testing.T) {
//	values := []struct{
//		input map[string]string
//		instanceType string
//		expected map[string]string
//	}{
//		{
//			map[string]string{"sentry_enabled": "true"},
//			"develop",
//			map[string]string{"sentry_enabled": "true", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "develop"},
//		},
//		{
//			map[string]string{"sentry_enabled": "false"},
//			"test",
//			map[string]string{"sentry_enabled": "false"},
//		},
//		{
//			map[string]string{"sentry_enabled": "True"},
//			"production",
//			map[string]string{"sentry_enabled": "True", "sentry_odoo_dir": "/home/odoo/instance/odoo", "sentry_environment": "production"},
//		},
//		{
//			map[string]string{"not_sentry": "True"},
//			"production",
//			map[string]string{"not_sentry": "True"},
//		},
//	}
//	for _, v := range values {
//		UpdateSentry(v.input, v.instanceType)
//		if !assert.Equal(t, v.expected, v.input) {
//			t.Errorf("Got: %+v, expected: %+v", v.input, v.expected)
//		}
//	}
//}
