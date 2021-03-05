package kapp

import (
	"bytes"
	"fmt"
	"io"
	"os"
	goexec "os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

type Kubeconfig interface {
	AsString() (string, string, error)
}

var _ ResourceData = &schema.ResourceData{}

type Kapp struct {
	data       SettableResourceData
	kubeconfig Kubeconfig
	logger     logger.Logger
}

func (t *Kapp) Deploy() (string, string, error) {
	args, env, stdin, err := t.addDeployArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building deploy args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Env = append(os.Environ(), env...)
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
	args, env, stdin, err := t.addDeployArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building deploy args: %s", err)
	}

	// TODO currently diff run leaves app record behind
	args = append(args, []string{"--diff-run", "--diff-exit-status"}...)

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Env = append(os.Environ(), env...)
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

			err := t.data.Set(schemaClusterDriftDetectedKey, true)
			if err != nil {
				return "", "", fmt.Errorf("Updating revision key: %s", err)
			}

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
	err := t.data.Set(schemaChangeDiffKey, stdout)
	if err != nil {
		return fmt.Errorf("Updating %s key: %s", schemaChangeDiffKey, err)
	}
	return nil
}

func (t *Kapp) Delete() (string, string, error) {
	args, env, stdin, err := t.addDeleteArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building delete args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kapp", args...)
	cmd.Stdin = stdin
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	if err != nil {
		stderrStr := stderrBs.String()
		return "", stderrStr, fmt.Errorf("Executing kapp: %s (stderr: %s)", err, stderrStr)
	}

	return stdoutBs.String(), "", nil
}

func (t *Kapp) addDeployArgs() ([]string, []string, io.Reader, error) {
	args := []string{
		"deploy",
		"-a", t.data.Get(schemaAppKey).(string),
		"-n", t.data.Get(schemaNamespaceKey).(string),
		"--yes",
		"--tty",
	}

	env := []string{}

	kubeconfig, kubeconfigContext, err := t.kubeconfig.AsString()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Building kubeconfig: %s", err)
	}

	if len(kubeconfigContext) > 0 {
		args = append(args, "--kubeconfig-context", kubeconfigContext)
	}

	env = append(env, "KAPP_KUBECONFIG_YAML="+kubeconfig)

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
			return nil, nil, nil, fmt.Errorf("Formatting %s: %s", schemaConfigYAMLKey, err)
		}

		stdin = bytes.NewReader([]byte(config))
	}

	files := t.data.Get(schemaFilesKey).([]interface{})
	if len(files) > 0 {
		for _, file := range files {
			args = append(args, "--file="+file.(string))
		}
	}

	deployOptsRaw := t.data.Get(schemaDeployKey).([]interface{})
	if len(deployOptsRaw) > 0 {
		deployOpts := deployOptsRaw[0].(map[string]interface{})
		if rawOpts, ok := deployOpts[schemaRawOptionsKey]; ok {
			for _, rawOpt := range rawOpts.([]interface{}) {
				args = append(args, rawOpt.(string))
			}
		}
	}

	return args, env, stdin, nil
}

func (t *Kapp) addDeleteArgs() ([]string, []string, io.Reader, error) {
	args := []string{
		"delete",
		"-a", t.data.Get(schemaAppKey).(string),
		"-n", t.data.Get(schemaNamespaceKey).(string),
		"--yes",
		"--tty",
	}

	env := []string{}

	kubeconfig, kubeconfigContext, err := t.kubeconfig.AsString()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Building kubeconfig: %s", err)
	}

	if len(kubeconfigContext) > 0 {
		args = append(args, "--kubeconfig-context", kubeconfigContext)
	}

	env = append(env, "KAPP_KUBECONFIG_YAML="+kubeconfig)

	deleteOptsRaw := t.data.Get(schemaDeleteKey).([]interface{})
	if len(deleteOptsRaw) > 0 {
		deleteOpts := deleteOptsRaw[0].(map[string]interface{})
		if rawOpts, ok := deleteOpts[schemaRawOptionsKey]; ok {
			for _, rawOpt := range rawOpts.([]interface{}) {
				args = append(args, rawOpt.(string))
			}
		}
	}

	return args, env, nil, nil
}
