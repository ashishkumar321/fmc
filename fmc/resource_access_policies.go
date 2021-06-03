package fmc

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var access_policy_type string = "AccessPolicy"
var access_policy_default_action_type string = "AccessPolicyDefaultAction"
var access_policy_default_syslog_alert_type string = "SyslogAlert"

func resourceAccessPolicies() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for Access Control Policies in FMC\n" +
			"\n" +
			"## Example\n" +
			"An example is shown below: \n" +
			"```hcl\n" +
			"resource \"fmc_access_policies\" \"access_policy\" {\n" +
			"    name = \"Terraform Access Policy\"\n" +
			"    # default_action = \"block\" # Cannot have block with base IPS policy\n" +
			"    default_action = \"permit\"\n" +
			"    default_action_base_intrusion_policy_id = data.fmc_ips_policies.ips_policy.id\n" +
			"    default_action_send_events_to_fmc = \"true\"\n" +
			"    default_action_log_end = \"true\"\n" +
			"    default_action_syslog_config_id = data.fmc_syslog_alerts.syslog_alert.id\n" +
			"}\n" +
			"```",
		CreateContext: resourceAccessPoliciesCreate,
		ReadContext:   resourceAccessPoliciesRead,
		DeleteContext: resourceAccessPoliciesDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of this resource",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The description of this resource",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of this resource",
			},
			"default_action": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := strings.ToUpper(val.(string))
					allowedValues := []string{"BLOCK", "TRUST", "PERMIT", "NETWORK_DISCOVERY", "INHERIT_FROM_PARENT"}
					for _, allowed := range allowedValues {
						if v == allowed {
							return
						}
					}
					errs = append(errs, fmt.Errorf("%q must be in %v, got: %q", key, allowedValues, v))
					return
				},
				Description: `Default action for this resource, "BLOCK", "TRUST", "PERMIT", "NETWORK_DISCOVERY" or "INHERIT_FROM_PARENT".`,
			},
			"default_action_base_intrusion_policy_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Default action base policy ID to inherit from for this resource",
			},
			"default_action_send_events_to_fmc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `Enable sending events to FMC for this resource, "true" or "false"`,
			},
			"default_action_log_begin": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `Enable logging at the beginning of the connection for this resource, "true" or "false`,
			},
			"default_action_log_end": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: `Enable logging at the end of the connection for this resource, "true" or "false"`,
			},
			"default_action_syslog_config_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Syslog configuration ID for this resource",
			},
			"default_action_type": {
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: "The type of default action of this resource",
			},
		},
	}
}

func resourceAccessPoliciesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)
	// Warning or errors can be collected in a slice type
	// var diags diag.Diagnostics
	var diags diag.Diagnostics

	res, err := c.CreateAccessPolicy(ctx, &AccessPolicy{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Defaultaction: AccessPolicyDefaultAction{
			Type: access_policy_default_action_type,
			Intrusionpolicy: AccessPolicyDefaultActionIntrusionPolicy{
				ID:   d.Get("default_action_base_intrusion_policy_id").(string),
				Type: access_policy_default_action_type,
			},
			Syslogconfig: AccessPolicyDefaultActionSyslogConfig{
				ID:   d.Get("default_action_syslog_config_id").(string),
				Type: access_policy_default_syslog_alert_type,
			},
			Logbegin:        d.Get("default_action_log_begin").(string),
			Logend:          d.Get("default_action_log_end").(string),
			Sendeventstofmc: d.Get("default_action_send_events_to_fmc").(string),
			Action:          strings.ToUpper(d.Get("default_action").(string)),
		},
		Type: access_policy_type,
	})
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to create access policy",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(res.ID)
	return resourceAccessPoliciesRead(ctx, d, m)
}

func resourceAccessPoliciesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()
	item, err := c.GetAccessPolicy(ctx, id)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to read access policy",
			Detail:   err.Error(),
		})
		return diags
	}
	if err := d.Set("name", item.Name); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to read access policy",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("description", item.Description); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to read access policy",
			Detail:   err.Error(),
		})
		return diags
	}

	if err := d.Set("type", item.Type); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to read access policy",
			Detail:   err.Error(),
		})
		return diags
	}

	return diags
}

func resourceAccessPoliciesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	id := d.Id()

	err := c.DeleteAccessPolicy(ctx, id)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "unable to delete access policy",
			Detail:   err.Error(),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
