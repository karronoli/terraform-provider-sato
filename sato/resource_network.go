package sato

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOrder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrderCreate,
		ReadContext:   resourceOrderRead,
		UpdateContext: resourceOrderUpdate,
		DeleteContext: resourceOrderDelete,
		Schema: map[string]*schema.Schema{
			"hardware_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnet_mask": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"gateway_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"dhcp": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"rarp": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceOrderCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	mac, err := net.ParseMAC(string(d.Get("hardware_address").(string)))

	if err != nil {
		panic(err)
	}

	dhcp := d.Get("dhcp").(bool)

	if dhcp {
		assign_dhcp(mac)
	} else {

		_ip := d.Get("ip_address").(string)
		_subnet := d.Get("subnet_mask").(string)
		_gateway := d.Get("gateway_address").(string)

		if len(_ip) == 0 || len(_subnet) == 0 || len(_gateway) == 0 {
			d.SetId("")
			tflog.Error(ctx, "some parameter not found. ip_address or subnet_mask or gateway_address")

			return diags
		}

		ip := net.ParseIP(_ip)
		subnet := net.ParseIP(_subnet)
		gateway := net.ParseIP(_gateway)

		assign_static_ip(mac, ip, subnet, gateway)
	}

	d.SetId(mac.String())

	resourceOrderRead(ctx, d, m)

	return diags
}

func resourceOrderRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	responses := search()

	if len(responses) == 0 {
		d.SetId("")

		return diags
	}

	hardware_address := d.Get("hardware_address").(string)
	var id string

	if len(hardware_address) > 0 {
		id = hardware_address
	} else {
		id = d.Id()
	}

	mac, err := net.ParseMAC(id)

	if err != nil {
		fmt.Println(d.Id())
		return diag.FromErr(err)
	}

	var response Response
	var found bool
	for _, v := range responses {
		if v.HardwareAddress.String() == mac.String() {
			response = v
			found = true
			break
		}
	}

	if !found {
		d.SetId("")

		return diags
	}

	if err := d.Set("ip_address", response.IPAddress.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("subnet_mask", response.SubnetMask.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("gateway_address", response.GatewayAddress.String()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("dhcp", response.DHCP); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("rarp", response.RARP); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceOrderUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	hw_addr_before, hw_addr_after := d.GetChange("hardware_address")
	mac_before, err := net.ParseMAC(hw_addr_before.(string))

	if err != nil {
		return diag.FromErr(err)
	}

	mac_after, err := net.ParseMAC(hw_addr_after.(string))

	if err != nil {
		return diag.FromErr(err)
	}

	tflog.Trace(ctx, fmt.Sprintf("before: %s, after: %s", mac_before, mac_after))

	if mac_before.String() != mac_after.String() ||
		d.HasChange("ip_address") ||
		d.HasChange("subnet_mask") ||
		d.HasChange("gateway_address") ||
		d.HasChange("dhcp") ||
		d.HasChange("rarp") {

		resourceOrderCreate(ctx, d, m)

		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceOrderRead(ctx, d, m)
}

func resourceOrderDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	mac, err := net.ParseMAC(string(d.Get("hardware_address").(string)))

	if err != nil {
		panic(err)
	}

	reset_network_settings(mac)
	d.SetId("")

	return diags
}
