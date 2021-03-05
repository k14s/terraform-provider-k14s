package kapp

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/k14s/terraform-provider-k14s/pkg/logger"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
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
	logger := r.newLogger(d, "create")

	d.SetId(r.id(d))

	r.clearDiff(d)
	defer r.clearDiff(d)

	_, _, err := (&Kapp{d, meta.(schemamisc.Context).Kubeconfig, logger}).Deploy()
	if err != nil {
		return fmt.Errorf("Creating %s: %s", r.id(d), err)
	}

	return nil
}

func (r Resource) Read(d *schema.ResourceData, meta interface{}) error {
	logger := r.newLogger(d, "read")

	d.SetId(r.id(d))

	r.clearDiff(d)

	// Updates revision to indicate change
	_, _, err := (&Kapp{d, meta.(schemamisc.Context).Kubeconfig, logger}).Diff()
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
	logger := r.newLogger(d, "update")

	// TODO do we need to set this?
	d.SetId(r.id(d))

	r.clearDiff(d)
	defer r.clearDiff(d)

	_, _, err := (&Kapp{d, meta.(schemamisc.Context).Kubeconfig, logger}).Deploy()
	if err != nil {
		return fmt.Errorf("Updating %s: %s", r.id(d), err)
	}

	return nil
}

func (r Resource) Delete(d *schema.ResourceData, meta interface{}) error {
	logger := r.newLogger(d, "delete")

	r.clearDiff(d)

	_, _, err := (&Kapp{d, meta.(schemamisc.Context).Kubeconfig, logger}).Delete()
	if err != nil {
		return fmt.Errorf("Deleting %s: %s", r.id(d), err)
	}

	d.SetId("")

	return nil
}

func (r Resource) CustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	logger := r.newLogger(diff, "customizeDiff")

	_, _, err := (&Kapp{SettableDiff{diff}, meta.(schemamisc.Context).Kubeconfig, logger}).Diff()
	if err != nil {
		r.logger.Error("Ignoring diffing error: %s", err)
	}

	return nil
}

func (r Resource) clearDiff(d SettableResourceData) {
	err := d.Set(schemaClusterDriftDetectedKey, false)
	if err != nil {
		panic(fmt.Sprintf("Updating %s key: %s", schemaClusterDriftDetectedKey, err))
	}

	err = d.Set(schemaChangeDiffKey, "")
	if err != nil {
		panic(fmt.Sprintf("Updating %s key: %s", schemaChangeDiffKey, err))
	}
}

func (r Resource) newLogger(d ResourceData, desc string) logger.Logger {
	if d.Get(schemaDebugLogsKey).(bool) {
		logger := r.logger.WithLabel(r.id(d)).WithLabel(desc)
		logger.Debug("started")
		return logger
	}
	return logger.NewNoop()
}

func (r Resource) id(d ResourceData) string {
	ns := d.Get(schemaNamespaceKey).(string)
	name := d.Get(schemaAppKey).(string)
	return fmt.Sprintf("%s/%s", ns, name)
}
