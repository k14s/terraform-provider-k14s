package kapp

import (
	"bytes"
	"fmt"
	"io"
	goexec "os/exec"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/logger"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

type ResourceData interface {
	Get(key string) interface{}
	GetOk(key string) (interface{}, bool)
}

type SettableResourceData interface {
	ResourceData
	Set(key string, val interface{}) error
}

var _ ResourceData = &schema.ResourceData{}

type Kapp struct {
	data   SettableResourceData
	logger logger.Logger
}

func (t *Kapp) Deploy() (string, string, error) {
	args, stdin, err := t.addDeployArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building deploy args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	if err != nil {
		stderrStr := stderrBs.String()
		return "", stderrStr, fmt.Errorf("Executing kapp: %s (stderr: %s)", err, stderrStr)
	}

	return stdoutBs.String(), "", nil
}

func (t *Kapp) Diff() (string, string, error) {
	args, stdin, err := t.addDeployArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building deploy args: %s", err)
	}

	// TODO currently diff run leaves app record behind
	args = append(args, []string{"--diff-run", "--diff-exit-status"}...)

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	stderrStr := stderrBs.String()

	if err == nil {
		return "", stderrStr, fmt.Errorf("Executing kapp: Expected "+
			"non-0 exit code (stderr: %s)", err, stderrStr)
	}

	if exitError, ok := err.(*goexec.ExitError); ok {
		switch exitError.ExitCode() {
		case 2: // no changes
			t.logger.Debug("no changes found")
			return "", "", nil

		case 3: // pending changes
			t.logger.Debug("pending changes found")
			return "", "", t.setDiff(stdoutBs.String())

		default:
			return "", stderrStr, fmt.Errorf("Executing kapp: Expected specific "+
				"exit error, but was %s (stderr: %s)", err, stderrStr)
		}
	}

	return "", stderrStr, fmt.Errorf("Executing kapp: Expected exit error, "+
		"but was %s (stderr: %s)", err, stderrStr)
}

func (t *Kapp) setDiff(stdout string) error {
	err := t.data.Set(schemaClusterDriftDetectedKey, true)
	if err != nil {
		return fmt.Errorf("Updating revision key: %s", err)
	}

	err = t.data.Set(schemaChangeDiffKey, stdout)
	if err != nil {
		return fmt.Errorf("Updating last deploy key: %s", err)
	}

	return nil
}

func (t *Kapp) Delete() (string, string, error) {
	args, stdin, err := t.addDeleteArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building delete args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	if err != nil {
		stderrStr := stderrBs.String()
		return "", stderrStr, fmt.Errorf("Executing kapp: %s (stderr: %s)", err, stderrStr)
	}

	return stdoutBs.String(), "", nil
}

func (t *Kapp) addDeployArgs() ([]string, io.Reader, error) {
	args := []string{
		"deploy",
		"-a", t.data.Get(schemaAppKey).(string),
		"-n", t.data.Get(schemaNamespaceKey).(string),
		"--yes",
		"--tty",
	}

	var stdin io.Reader

	diffChanges, exists := t.data.GetOk(schemaDiffChangesKey)
	if exists && diffChanges.(bool) {
		args = append(args, "--diff-changes")
	}

	diffContext, exists := t.data.GetOk(schemaDiffContextKey)
	if exists {
		args = append(args, fmt.Sprintf("--diff-context=%d", diffContext.(int)))
	}

	config := t.data.Get(schemaConfigYAMLKey).(string)
	if len(config) > 0 {
		args = append(args, "-f-")

		config, err := schemamisc.Heredoc{config}.StripIndent()
		if err != nil {
			return nil, nil, fmt.Errorf("Formatting %s: %s", schemaConfigYAMLKey, err)
		}

		stdin = bytes.NewReader([]byte(config))
	}

	files := t.data.Get(schemaFilesKey).([]interface{})
	if len(files) > 0 {
		for _, file := range files {
			args = append(args, "--file="+file.(string))
		}
	}

	return args, stdin, nil
}

func (t *Kapp) addDeleteArgs() ([]string, io.Reader, error) {
	args := []string{
		"delete",
		"-a", t.data.Get(schemaAppKey).(string),
		"-n", t.data.Get(schemaNamespaceKey).(string),
		"--yes",
		"--tty",
	}
	return args, nil, nil
}