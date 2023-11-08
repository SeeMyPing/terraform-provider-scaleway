package scaleway

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayDomainExternal() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayDomainExternalCreate,
		ReadContext:   resourceScalewayDomainExternalRead,
		UpdateContext: resourceScalewayDomainExternalUpdate,
		DeleteContext: resourceScalewayDomainExternalDelete,
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultDomainZoneTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Description: "The external domain to register into scaleway nameservers.",
				Required:    true,
				ForceNew:    true,
			},
			"token": {
				Type:        schema.TypeString,
				Description: "The domain validation token.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "The domain zone status.",
				Computed:    true,
			},
			"message": {
				Type:        schema.TypeString,
				Description: "Message",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "The date and time of the last update of the DNS zone.",
				Computed:    true,
			},
			"project_id": projectIDSchema(),
		},
	}
}

func resourceScalewayDomainExternalCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	domainName := strings.ToLower(d.Get("domain").(string))

	domains, err := domainAPI.ListDomains(&domain.ListDomainsRequest{
		ProjectID: expandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	for i := range domains.Domains {
		if domain.Domains[i].Domain == domainName 
		{
			d.SetId(fmt.Sprintf("%s", domainName))

			return resourceScalewayDomainExternalRead(ctx, d, meta)
		}
	}

	var dnsZone *domain.DNSZone

	dnsZone, err = domainAPI.CreateExternalDomain(&domain.RegistrarAPIRegisterExternalDomainRequest{
		ProjectID: d.Get("project_id").(string),
		Domain:    domainName,
	}, scw.WithContext(ctx))

	if err != nil {
		if is409Error(err) {
			return resourceScalewayDomainExternalRead(ctx, d, meta)
		}
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%s", domains.Domain))

	return resourceScalewayDomainExternalRead(ctx, d, meta)
}

func resourceScalewayDomainExternalRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	domainAPI := newDomainAPI(meta)

	var domain *domain.Domain

	domain, err := domainAPI.ListDomains(&domain.ListDomainsRequest{
		ProjectID: expandStringPtr(d.Get("project_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if len(domain.Domains) == 0 {
		return diag.FromErr(fmt.Errorf("no domain found with the name %s", d.Id()))
	}

	if len(domain.Domains) > 1 {
		return diag.FromErr(fmt.Errorf("%d domain found with the same name %s", len(domain.Domains), d.Id()))
	}

	domain = domain.Domains[0]
	_ = d.Set("domain", domain.Domain)
	_ = d.Set("dnssec_status", domain.DnssecStatus.String())
	_ = d.Set("status", domain.Status.String())
	_ = d.Set("epp_code", domain.EppCode.String())
	_ = d.Set("updated_at", domain.UpdatedAt.String())
	_ = d.Set("expire_at", domain.ExpiredAt.String())
	_ = d.Set("is_external", domain.IsExternal.String())
	_ = d.Set("registrar", domain.Registrar.String())
	_ = d.Set("project_id", domain.ProjectID)

	return nil
}
