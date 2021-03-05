package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

type Kubeconfig struct {
	kappRaw []interface{}
}

func NewKubeconfig(d *schema.ResourceData) Kubeconfig {
	// Do not save off resource data as it should not be accessed
	// beyond this method call (ConfigureFunc is parent)
	val, ok := d.Get(schemaKappKey).([]interface{})
	if !ok {
		val = []interface{}{}
	}
	return Kubeconfig{val}
}

func (c Kubeconfig) AsString() (string, string, error) {
	if len(c.kappRaw) == 0 {
		return "", "", fmt.Errorf("Expected non-empty provider config (key '%s')", schemaKappKey)
	}

	kapp := c.kappRaw[0].(map[string]interface{})

	kubeconfigYAML := kapp[schemaKappKubeconfigYAMLKey].(string)
	kubeconfigRaw := kapp[schemaKappKubeconfigKey].([]interface{})

	if len(kubeconfigYAML) > 0 {
		if len(kubeconfigRaw) != 0 {
			return "", "", fmt.Errorf("Expected empty config (key '%s')", schemaKappKubeconfigKey)
		}

		kubeconfigYAML, err := schemamisc.Heredoc{kubeconfigYAML}.StripIndent()
		if err != nil {
			return "", "", fmt.Errorf("Formatting %s: %s", schemaKappKubeconfigYAMLKey, err)
		}

		return kubeconfigYAML, "", nil
	}

	if len(kubeconfigRaw) != 1 {
		return "", "", fmt.Errorf("Expected non-empty config (key '%s')", schemaKappKubeconfigKey)
	}

	kubeconfig := kubeconfigRaw[0].(map[string]interface{})

	fromEnv := kubeconfig[schemaKappKubeconfigFromEnvKey].(bool)
	context := kubeconfig[schemaKappKubeconfigContextKey].(string)
	connInfo := KubeconfigConnInfo{
		Server:     kubeconfig[schemaKappKubeconfigServerKey].(string),
		Username:   kubeconfig[schemaKappKubeconfigUsernameKey].(string),
		Password:   kubeconfig[schemaKappKubeconfigPasswordKey].(string),
		CACert:     kubeconfig[schemaKappKubeconfigCACertKey].(string),
		ClientCert: kubeconfig[schemaKappKubeconfigClientCertKey].(string),
		ClientKey:  kubeconfig[schemaKappKubeconfigClientKeyKey].(string),
		Token:      kubeconfig[schemaKappKubeconfigTokenKey].(string),
	}

	connInfo.StripIndentInCerts()

	if fromEnv {
		if connInfo.HasValues() {
			return "", "", fmt.Errorf("Expected key '%s' to be the only key configured",
				schemaKappKubeconfigFromEnvKey)
		}
		return "", context, nil
	}

	if len(context) > 0 {
		return "", "", fmt.Errorf("Expected key '%s' to be configured with key '%s'",
			schemaKappKubeconfigContextKey, schemaKappKubeconfigFromEnvKey)
	}

	return c.buildConfig(connInfo)
}

func (Kubeconfig) buildConfig(connInfo KubeconfigConnInfo) (string, string, error) {
	clusterConfig := map[string]interface{}{
		"server": connInfo.Server,
	}

	if len(connInfo.CACert) > 0 {
		clusterConfig["certificate-authority-data"] = base64.StdEncoding.EncodeToString([]byte(connInfo.CACert))
	}

	userConfig := map[string]interface{}{}

	if len(connInfo.Username) > 0 {
		userConfig["username"] = connInfo.Username
	}
	if len(connInfo.Password) > 0 {
		userConfig["password"] = connInfo.Password
	}
	if len(connInfo.ClientCert) > 0 {
		userConfig["client-certificate-data"] = base64.StdEncoding.EncodeToString([]byte(connInfo.ClientCert))
	}
	if len(connInfo.ClientKey) > 0 {
		userConfig["client-key-data"] = base64.StdEncoding.EncodeToString([]byte(connInfo.ClientKey))
	}
	if len(connInfo.Token) > 0 {
		userConfig["token"] = connInfo.Token
	}

	const (
		clusterName = "kapp-cluster"
		userName    = "kapp-user"
		contextName = "kapp-context"
	)

	mergedConfig := map[string]interface{}{
		"apiVersion":      "v1",
		"kind":            "Config",
		"current-context": contextName,
		"clusters": []interface{}{
			map[string]interface{}{
				"name":    clusterName,
				"cluster": clusterConfig,
			},
		},
		"contexts": []interface{}{
			map[string]interface{}{
				"name": contextName,
				"context": map[string]interface{}{
					"cluster": clusterName,
					"user":    userName,
				},
			},
		},
		"users": []interface{}{
			map[string]interface{}{
				"name": userName,
				"user": userConfig,
			},
		},
	}

	mergedConfigBs, err := json.Marshal(mergedConfig)
	if err != nil {
		return "", "", fmt.Errorf("Marshaling kubeconfig: %s", err)
	}

	return string(mergedConfigBs), "", nil
}

type KubeconfigConnInfo struct {
	Server     string
	Username   string
	Password   string
	CACert     string
	ClientCert string
	ClientKey  string
	Token      string
}

func (a KubeconfigConnInfo) HasValues() bool {
	return len(a.Server+a.Username+a.Password+a.CACert+a.ClientCert+a.ClientKey+a.Token) > 0
}

func (a *KubeconfigConnInfo) StripIndentInCerts() error {
	var err error

	a.CACert, err = schemamisc.Heredoc{a.CACert}.StripIndent()
	if err != nil {
		return fmt.Errorf("Formatting CA certificate: %s", err)
	}

	a.ClientCert, err = schemamisc.Heredoc{a.ClientCert}.StripIndent()
	if err != nil {
		return fmt.Errorf("Formatting client certificate: %s", err)
	}

	a.ClientKey, err = schemamisc.Heredoc{a.ClientKey}.StripIndent()
	if err != nil {
		return fmt.Errorf("Formatting client key: %s", err)
	}

	return nil
}
