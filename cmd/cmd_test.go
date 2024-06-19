package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func assertErrorNil(err error) bool {
	return err == nil
}
func TestRootCmd(t *testing.T) {
	resourceDataJSONPath := "../test/fixtures/service1-raw.json"
	resourceDataJSONBytes, err := os.ReadFile(resourceDataJSONPath)
	resourceDataJSON := string(resourceDataJSONBytes)
	if err != nil {
		t.Errorf("error readin test data file %s: %v", resourceDataJSONPath, err)
	}
	resourceDataYAMLPath := "../test/fixtures/service1-raw.yaml"
	resourceDataYAMLBytes, err := os.ReadFile(resourceDataYAMLPath)
	resourceDataYAML := string(resourceDataYAMLBytes)
	if err != nil {
		t.Errorf("error readin test data file %s: %v", resourceDataYAMLPath, err)
	}

	testcases := []struct {
		args        []string
		stdin       string
		assertError func(err error) bool
		expOut      string
	}{
		{
			args:        []string{},
			stdin:       "",
			assertError: assertErrorNil,
			expOut:      "",
		},
		{
			args:        []string{},
			stdin:       resourceDataJSON,
			assertError: assertErrorNil,
			expOut:      "apiVersion",
		},
		{
			args:        []string{},
			stdin:       resourceDataYAML,
			assertError: assertErrorNil,
			expOut:      "apiVersion",
		},
		{
			args:        []string{"-f", "-"},
			stdin:       resourceDataJSON,
			assertError: assertErrorNil,
			expOut:      "apiVersion",
		},
		{
			args:  []string{"-f", "/nogood"},
			stdin: "",
			assertError: func(err error) bool {
				_, ok := err.(*os.PathError)
				return ok
			},
			expOut: "",
		},
		{
			args:        []string{"-f", resourceDataJSONPath},
			stdin:       "",
			assertError: assertErrorNil,
			expOut:      "apiVersion",
		},
		{
			args:        []string{"-f", resourceDataYAMLPath},
			stdin:       "",
			assertError: assertErrorNil,
			expOut:      "apiVersion",
		},
	}

	for _, tc := range testcases {
		rootCmd.SetArgs(tc.args)
		if tc.stdin != "" {
			rootCmd.SetIn(bytes.NewReader([]byte(tc.stdin)))
		}
		cmdout := new(bytes.Buffer)
		cmderr := new(bytes.Buffer)
		rootCmd.SetOut(cmdout)
		rootCmd.SetErr(cmderr)
		rootCmd.ParseFlags(tc.args)
		resErr := rootCmd.RunE(rootCmd, tc.args)
		resStdout, err := io.ReadAll(cmdout)
		if err != nil {
			t.Errorf("error reading command output: %v", err)
		}
		resStderr, err := io.ReadAll(cmderr)
		if err != nil {
			t.Errorf("error reading command error: %v\ntest case: %v", err, tc)
		}
		if tc.assertError != nil && !tc.assertError(resErr) {
			t.Errorf("error assertion: have: %#v\ntest case: %v", resErr, tc)
		}
		if !strings.Contains(string(resStdout), tc.expOut) {
			t.Errorf("stdout assertion: have: %s\nwant: %s\ntest case: %v", string(resStdout), tc.expOut, tc)
		}
		if len(resStderr) > 0 {
			t.Errorf("stderr not empty: %s\ntest case: %v", string(resStderr), tc)
		}
	}
}

func TestGetCmd(t *testing.T) {
	kubectl = "../test/kubectl-stub"
	testcases := []struct {
		args        []string
		assertError func(err error) bool
		expOut      string
		expErr      string
	}{
		{
			args: []string{""},
			assertError: func(err error) bool {
				return strings.HasPrefix(err.Error(), "error invoking kubectl")
			},
			expOut: "",
			expErr: "",
		},
		{
			args:        []string{"pods"},
			assertError: assertErrorNil,
			expOut:      "apiVersion",
			expErr:      "",
		},
		{
			args:        []string{"pods", "mypod"},
			assertError: assertErrorNil,
			expOut:      "apiVersion",
			expErr:      "",
		},
		{
			args:        []string{"pods", "mypod", "-o", "yaml"},
			assertError: assertErrorNil,
			expOut:      "apiVersion",
			expErr:      "",
		},
		{
			args:        []string{"pods", "mypod", "-o", "json"},
			assertError: assertErrorNil,
			expOut:      "apiVersion",
			expErr:      "",
		},
	}

	for _, tc := range testcases {
		rootCmd.SetArgs(tc.args)
		cmdout := new(bytes.Buffer)
		cmderr := new(bytes.Buffer)
		rootCmd.SetOut(cmdout)
		rootCmd.SetErr(cmderr)
		rootCmd.ParseFlags(tc.args)
		resErr := getCmd.RunE(getCmd, tc.args)
		resStdout, err := io.ReadAll(cmdout)
		if err != nil {
			t.Errorf("error reading command output: %v", err)
		}
		resStderr, err := io.ReadAll(cmderr)
		if err != nil {
			t.Errorf("error reading command error: %v\ntest case: %v", err, tc)
		}
		if tc.assertError != nil && !tc.assertError(resErr) {
			t.Errorf("error assertion: have: %#v\ntest case: %v", resErr, tc)
		}
		if !strings.Contains(string(resStdout), tc.expOut) {
			t.Errorf("stdout assertion: have: %s\nwant: %s\ntest case: %v", string(resStdout), tc.expOut, tc)
		}
		if len(resStderr) > 0 {
			t.Errorf("stderr not empty: %s\ntest case: %v", string(resStderr), tc)
		}
	}
}
