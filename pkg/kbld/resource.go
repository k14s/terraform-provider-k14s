package kbld

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/logger"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

const (
	schemaFilesKey      = "files"
	schemaConfigYAMLKey = "config_yaml"
	schemaResultKey     = "result"
	schemaDebugLogsKey  = "debug_logs"
)

type Resource struct {
	logger logger.Logger
}

func NewResource(logger logger.Logger) *schema.Resource {
	res := Resource{logger}

	return &schema.Resource{
		Read: res.Read,
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func (r Resource) Read(d *schema.ResourceData, meta interface{}) error {
	var logger logger.Logger = logger.NewNoop()

	if d.Get(schemaDebugLogsKey).(bool) {
		logger = r.logger.WithLabel("read")
		logger.Debug("started")
	}

	stdout, _, err := (&Kbld{d}).Template()
	if err != nil {
		return err
	}

	d.Set(schemaResultKey, stdout)
	d.SetId(schemamisc.SHA256Sum(stdout))

	logger.Debug("id=%s", d.Id())

	return nil
}
