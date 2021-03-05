package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	schemaKappKey               = "kapp"
	schemaKappKubeconfigKey     = "kubeconfig"
	schemaKappKubeconfigYAMLKey = "kubeconfig_yaml"

	schemaKappKubeconfigFromEnvKey    = "from_env"
	schemaKappKubeconfigContextKey    = "context"
	schemaKappKubeconfigServerKey     = "server"
	schemaKappKubeconfigUsernameKey   = "username"
	schemaKappKubeconfigPasswordKey   = "password"
	schemaKappKubeconfigCACertKey     = "ca_cert"
	schemaKappKubeconfigClientCertKey = "client_cert"
	schemaKappKubeconfigClientKeyKey  = "client_key"
	schemaKappKubeconfigTokenKey      = "token"
)

var (
	resourceSchema = map[string]*schema.Schema{
		schemaKappKey: {
			Type:        schema.TypeList,
			Description: "Kapp options",
			Optional:    true,
			MinItems:    0,
			MaxItems:    1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					schemaKappKubeconfigKey: {
						Type:        schema.TypeList,
						Description: "kubeconfig used by kapp",
						Optional:    true,
						MinItems:    0,
						MaxItems:    1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								schemaKappKubeconfigFromEnvKey: {
									Type:        schema.TypeBool,
									Description: "Pull configuration from environment (typically found in ~/.kube/config or via $KUBECONFIG)",
									Optional:    true,
								},
								schemaKappKubeconfigContextKey: {
									Type:        schema.TypeString,
									Description: "Use particular context",
									Optional:    true,
								},
								schemaKappKubeconfigServerKey: {
									Type:        schema.TypeString,
									Description: "Address of API server",
									Optional:    true,
								},
								schemaKappKubeconfigUsernameKey: {
									Type:        schema.TypeString,
									Description: "Username",
									Optional:    true,
								},
								schemaKappKubeconfigPasswordKey: {
									Type:        schema.TypeString,
									Description: "Password",
									Optional:    true,
								},
								schemaKappKubeconfigCACertKey: {
									Type:        schema.TypeString,
									Description: "CA certificate in PEM format",
									Optional:    true,
								},
								schemaKappKubeconfigClientCertKey: {
									Type:        schema.TypeString,
									Description: "Client certificate in PEM format",
									Optional:    true,
								},
								schemaKappKubeconfigClientKeyKey: {
									Type:        schema.TypeString,
									Description: "Client key in PEM format",
									Optional:    true,
								},
								schemaKappKubeconfigTokenKey: {
									Type:        schema.TypeString,
									Description: "Auth token",
									Optional:    true,
								},
							},
						},
					},
					schemaKappKubeconfigYAMLKey: {
						Type:        schema.TypeString,
						Description: "kubeconfig as YAML",
						Optional:    true,
					},
				},
			},
		},
	}
)
