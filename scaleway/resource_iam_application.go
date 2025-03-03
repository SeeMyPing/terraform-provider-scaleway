package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayIamApplication() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayIamApplicationCreate,
		ReadContext:   resourceScalewayIamApplicationRead,
		UpdateContext: resourceScalewayIamApplicationUpdate,
		DeleteContext: resourceScalewayIamApplicationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The name of the iam application",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the iam application",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the application",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the application",
			},
			"editable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the application is editable.",
			},
			"organization_id": organizationIDOptionalSchema(),
		},
	}
}

func resourceScalewayIamApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)
	app, err := api.CreateApplication(&iam.CreateApplicationRequest{
		Name:           expandOrGenerateString(d.Get("name"), "application"),
		Description:    d.Get("description").(string),
		OrganizationID: d.Get("organization_id").(string),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(app.ID)

	return resourceScalewayIamApplicationRead(ctx, d, meta)
}

func resourceScalewayIamApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)
	app, err := api.GetApplication(&iam.GetApplicationRequest{
		ApplicationID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	_ = d.Set("name", app.Name)
	_ = d.Set("description", app.Description)
	_ = d.Set("created_at", flattenTime(app.CreatedAt))
	_ = d.Set("updated_at", flattenTime(app.UpdatedAt))
	_ = d.Set("organization_id", app.OrganizationID)
	_ = d.Set("editable", app.Editable)

	return nil
}

func resourceScalewayIamApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	req := &iam.UpdateApplicationRequest{
		ApplicationID: d.Id(),
	}

	hasChanged := false

	if d.HasChange("name") {
		req.Name = expandStringPtr(d.Get("name"))
		hasChanged = true
	}
	if d.HasChange("description") {
		req.Description = expandUpdatedStringPtr(d.Get("description"))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateApplication(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayIamApplicationRead(ctx, d, meta)
}

func resourceScalewayIamApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := iamAPI(meta)

	err := api.DeleteApplication(&iam.DeleteApplicationRequest{
		ApplicationID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}
