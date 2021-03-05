package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/k14s/terraform-provider-k14s/pkg/kapp"
	"github.com/k14s/terraform-provider-k14s/pkg/kbld"
	"github.com/k14s/terraform-provider-k14s/pkg/logger"
	"github.com/k14s/terraform-provider-k14s/pkg/schemamisc"
	"github.com/k14s/terraform-provider-k14s/pkg/ytt"
)

func Provider() terraform.ResourceProvider {
	logger := logger.MustNewFileRoot("/tmp/terraform-provider-k14s.log")
	yttLogger := logger.WithLabel("ytt")
	kbldLogger := logger.WithLabel("kbld")

	// TODO different naming
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"k14s_ytt":  ytt.NewResource(yttLogger),
			"k14s_kbld": kbld.NewResource(kbldLogger),
		},
		ResourcesMap: map[string]*schema.Resource{
			"k14s_ytt":  schema.DataSourceResourceShim("k14s_ytt", ytt.NewResource(yttLogger)),
			"k14s_kbld": schema.DataSourceResourceShim("k14s_kbld", kbld.NewResource(kbldLogger)),
			"k14s_kapp": kapp.NewResource(logger.WithLabel("kapp")),
		},

		Schema: resourceSchema,

		ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
			return schemamisc.Context{Kubeconfig: NewKubeconfig(d)}, nil
		},
	}
}
