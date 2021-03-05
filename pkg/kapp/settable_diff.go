package kapp

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type SettableDiff struct {
	diff *schema.ResourceDiff
}

func (d SettableDiff) Get(key string) interface{} {
	return d.diff.Get(key)
}

func (d SettableDiff) GetOk(key string) (interface{}, bool) {
	return d.diff.GetOk(key)
}

func (d SettableDiff) Set(key string, val interface{}) error {
	if keySchema, found := resourceSchema[key]; found {
		if keySchema.Computed {
			return d.diff.SetNew(key, val)
		}
		return nil
	}
	panic(fmt.Sprintf("Expected to find schema for key '%s'", key))
}
