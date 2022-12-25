package sato

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	responses := search()

	var diags diag.Diagnostics

	if len(responses) == 0 {
		d.SetId("")

		return diags
	}

	var dummy []map[string]interface{}
	for _, response := range responses {
		dummy = append(dummy, map[string]interface{}{
			"hardware_address": response.HardwareAddress.String(),
			"ip_address":       response.IPAddress.String(),
			"subnet_mask":      response.SubnetMask.String(),
			"gateway_address":  response.GatewayAddress.String(),
			"name":             response.Name,
			"dhcp":             response.DHCP,
			"rarp":             response.RARP,
		})
	}

	if err := d.Set("networks", dummy); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func dataSourceNetworks() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkRead,
		Schema: map[string]*schema.Schema{
			"networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hardware_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_mask": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"gateway_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dhcp": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"rarp": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
