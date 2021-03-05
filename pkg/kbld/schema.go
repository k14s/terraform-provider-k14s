package kbld

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const (
	schemaFilesKey      = "files"
	schemaConfigYAMLKey = "config_yaml"
	schemaResultKey     = "result"
	schemaDebugLogsKey  = "debug_logs"
)

var (
	resourceScheme = map[string]*schema.Schema{
		schemaFilesKey: {
			Type:        schema.TypeList,
			Description: "Files",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		schemaConfigYAMLKey: {
			Type:        schema.TypeString,
			Description: "Configuration as YAML",
			Optional:    true,
			Sensitive:   true,
		},
		schemaResultKey: {
			Type:        schema.TypeString,
			Description: "Result",
			Computed:    true,
			Sensitive:   true,
		},
		schemaDebugLogsKey: {
			Type:        schema.TypeBool,
			Description: "Enable debug logging",
			Optional:    true,
			Default:     false,
		},
	}
)
