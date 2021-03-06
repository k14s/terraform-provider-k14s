package kapp

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware-tanzu/terraform-provider-carvel/pkg/logger"
	"github.com/vmware-tanzu/terraform-provider-carvel/pkg/schemamisc"
)

type Resource struct {
	logger logger.Logger
}

func NewResource(logger logger.Logger) *schema.Resource {
	res := Resource{logger}

	return &schema.Resource{
		Create:        res.Create,
		Read:          res.Read,
		Update:        res.Update,
		Delete:        res.Delete,
		CustomizeDiff: res.CustomizeDiff,
		Schema:        resourceSchema,
	}
}

func (r Resource) Create(d *schema.ResourceData, meta interface{}) error {
	logger, diffLogger := r.newLogger(d, meta, "create")

	d.SetId(r.id(d))

	r.clearDiff(d)
	defer r.clearDiff(d)

	ctx := meta.(schemamisc.Context)

	_, _, err := (&Kapp{d, ctx.Kubeconfig, diffLogger, logger}).Deploy()
	if err != nil {
		return fmt.Errorf("Creating %s: %s", r.id(d), err)
	}

	return nil
}

func (r Resource) Read(d *schema.ResourceData, meta interface{}) error {
	logger, diffLogger := r.newLogger(d, meta, "read")

	d.SetId(r.id(d))

	r.clearDiff(d)

	ctx := meta.(schemamisc.Context)

	// Updates revision to indicate change
	_, _, err := (&Kapp{d, ctx.Kubeconfig, diffLogger, logger}).Diff()
	if err != nil {
		r.logger.Error("Ignoring diffing error: %s", err)
		// TODO ignore diffing error since it might
		// be diffed against invalid old configuration
		// (eg Ownership error with previously set configuration).
		// return fmt.Errorf("Reading %s: %s", r.id(d), err)
	}

	return nil
}

func (r Resource) Update(d *schema.ResourceData, meta interface{}) error {
	logger, diffLogger := r.newLogger(d, meta, "update")

	// TODO do we need to set this?
	d.SetId(r.id(d))

	r.clearDiff(d)
	defer r.clearDiff(d)

	ctx := meta.(schemamisc.Context)

	_, _, err := (&Kapp{d, ctx.Kubeconfig, diffLogger, logger}).Deploy()
	if err != nil {
		return fmt.Errorf("Updating %s: %s", r.id(d), err)
	}

	return nil
}

func (r Resource) Delete(d *schema.ResourceData, meta interface{}) error {
	logger, diffLogger := r.newLogger(d, meta, "delete")

	r.clearDiff(d)

	ctx := meta.(schemamisc.Context)

	_, _, err := (&Kapp{d, ctx.Kubeconfig, diffLogger, logger}).Delete()
	if err != nil {
		return fmt.Errorf("Deleting %s: %s", r.id(d), err)
	}

	d.SetId("")

	return nil
}

func (r Resource) CustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	logger, diffLogger := r.newLogger(diff, meta, "customizeDiff")

	ctx := meta.(schemamisc.Context)

	_, _, err := (&Kapp{SettableDiff{diff, logger}, ctx.Kubeconfig, diffLogger, logger}).Diff()
	if err != nil {
		logger.Error("Ignoring diffing error: %s", err)
	}

	return nil
}

func (r Resource) clearDiff(d SettableResourceData) {
	err := d.Set(schemaClusterDriftDetectedKey, false)
	if err != nil {
		panic(fmt.Sprintf("Updating %s key: %s", schemaClusterDriftDetectedKey, err))
	}
}

func (r Resource) newLogger(d ResourceData, meta interface{}, desc string) (logger.Logger, logger.Logger) {
	var logger logger.Logger = logger.NewNoop()

	if d.Get(schemaDebugLogsKey).(bool) {
		logger = r.logger.WithLabel(r.id(d)).WithLabel(desc)
		logger.Debug("started")
	}

	diffLogger := meta.(schemamisc.Context).DiffPreviewLogger

	return logger, diffLogger.WithLabel(r.id(d)).WithLabel(desc)
}

func (r Resource) id(d ResourceData) string {
	ns := d.Get(schemaNamespaceKey).(string)
	name := d.Get(schemaAppKey).(string)
	return fmt.Sprintf("%s/%s", ns, name)
}
