package utils

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{"extra_addons/ircodoo", "/home/odoo/instance/extra_addons/ircodoo"},
		{"", "/home/odoo/instance/odoo"},
	}
	for _, v := range values {
		t.Setenv("MAIN_REPO_PATH", v.mainRepo)
		res := GetMainRepoPath()
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
}

func TestGetMainRepoBranch(t *testing.T) {
	repoPath := t.TempDir()
	if err := exec.Command("git", "init", "-b", "18.0-dev1-moy", repoPath).Run(); err != nil {
		t.Skipf("git is not available: %s", err)
	}
	if !assert.Equal(t, "18.0-dev1-moy", GetMainRepoBranch(repoPath)) {
		t.Errorf("The branch of the repository was not reported.")
	}
	if !assert.Equal(t, "", GetMainRepoBranch(t.TempDir())) {
		t.Errorf("A directory without a repository must report no branch.")
	}
}

func TestGetImageTag(t *testing.T) {
	t.Setenv("MAIN_APP", "islamicrelief")
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
	if !assert.Equal(t, "islamicrelief-19.0-"+commit, GetImageTag(repoPath)) {
		t.Errorf("The dist must be shaped like the tag of the image.")
	}
	if !assert.Equal(t, "", GetImageTag(t.TempDir())) {
		t.Errorf("Without a repository there is no tag to report.")
	}
	t.Setenv("MAIN_APP", "")
	if !assert.Equal(t, "", GetImageTag(repoPath)) {
		t.Errorf("Without MAIN_APP there is no tag to report.")
	}
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
		res := GetSentryEnvironment(v.instanceType)
		if !assert.Equal(t, v.expected, res) {
			t.Errorf("Got: %+v, expected: %+v", res, v.expected)
		}
	}
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
