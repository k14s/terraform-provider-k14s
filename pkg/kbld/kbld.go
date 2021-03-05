package kbld

import (
	"bytes"
	"fmt"
	"io"
	goexec "os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

type Kbld struct {
	data *schema.ResourceData
}

func (t *Kbld) Template() (string, string, error) {
	args, stdin, err := t.addArgs()
	if err != nil {
		return "", "", fmt.Errorf("Building args: %s", err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kbld", args...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs

	err = cmd.Run()
	if err != nil {
		stderrStr := stderrBs.String()
		return "", stderrStr, fmt.Errorf("Executing kbld: %s (stderr: %s)", err, stderrStr)
	}

	return stdoutBs.String(), "", nil
}

func (t *Kbld) addArgs() ([]string, io.Reader, error) {
	args := []string{}
	var stdin io.Reader

	files := t.data.Get(schemaFilesKey).([]interface{})
	if len(files) > 0 {
		for _, file := range files {
			args = append(args, "--file="+file.(string))
		}
	}

	values := t.data.Get(schemaConfigYAMLKey).(string)
	if len(values) > 0 {
		args = append(args, "-f-")

		values, err := schemamisc.Heredoc{values}.StripIndent()
		if err != nil {
			return nil, nil, fmt.Errorf("Formatting %s: %s", schemaConfigYAMLKey, err)
		}

		stdin = bytes.NewReader([]byte(values))
	}

	return args, stdin, nil
}
