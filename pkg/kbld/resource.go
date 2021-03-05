package kbld

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/logger"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
)

type Resource struct {
	logger logger.Logger
}

func NewResource(logger logger.Logger) *schema.Resource {
	res := Resource{logger}
	return &schema.Resource{Read: res.Read, Schema: resourceScheme}
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
