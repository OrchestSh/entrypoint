package utils

import (
	"os"
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
	os.Unsetenv("ODOO_CONFIG_FILE")
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
