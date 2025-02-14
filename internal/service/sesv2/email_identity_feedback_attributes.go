package sesv2

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceEmailIdentityFeedbackAttributes() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailIdentityFeedbackAttributesCreate,
		ReadWithoutTimeout:   resourceEmailIdentityFeedbackAttributesRead,
		UpdateWithoutTimeout: resourceEmailIdentityFeedbackAttributesUpdate,
		DeleteWithoutTimeout: resourceEmailIdentityFeedbackAttributesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"email_forwarding_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"email_identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameEmailIdentityFeedbackAttributes = "Email Identity Feedback Attributes"
)

func resourceEmailIdentityFeedbackAttributesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Client

	in := &sesv2.PutEmailIdentityFeedbackAttributesInput{
		EmailIdentity:          aws.String(d.Get("email_identity").(string)),
		EmailForwardingEnabled: d.Get("email_forwarding_enabled").(bool),
	}

	out, err := conn.PutEmailIdentityFeedbackAttributes(ctx, in)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameEmailIdentityFeedbackAttributes, d.Get("email_identity").(string), err)
	}

	if out == nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResNameEmailIdentityFeedbackAttributes, d.Get("email_identity").(string), errors.New("empty output"))
	}

	d.SetId(d.Get("email_identity").(string))

	return resourceEmailIdentityFeedbackAttributesRead(ctx, d, meta)
}

func resourceEmailIdentityFeedbackAttributesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Client

	out, err := FindEmailIdentityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SESV2 EmailIdentityFeedbackAttributes (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResNameEmailIdentityFeedbackAttributes, d.Id(), err)
	}

	d.Set("email_identity", d.Id())
	d.Set("email_forwarding_enabled", out.FeedbackForwardingStatus)

	return nil
}

func resourceEmailIdentityFeedbackAttributesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Client

	update := false

	in := &sesv2.PutEmailIdentityFeedbackAttributesInput{
		EmailIdentity: aws.String(d.Id()),
	}

	if d.HasChanges("email_forwarding_enabled") {
		in.EmailForwardingEnabled = d.Get("email_forwarding_enabled").(bool)
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating SESV2 EmailIdentityFeedbackAttributes (%s): %#v", d.Id(), in)
	_, err := conn.PutEmailIdentityFeedbackAttributes(ctx, in)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionUpdating, ResNameEmailIdentityFeedbackAttributes, d.Id(), err)
	}

	return resourceEmailIdentityFeedbackAttributesRead(ctx, d, meta)
}

func resourceEmailIdentityFeedbackAttributesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Client

	log.Printf("[INFO] Deleting SESV2 EmailIdentityFeedbackAttributes %s", d.Id())

	_, err := conn.PutEmailIdentityFeedbackAttributes(ctx, &sesv2.PutEmailIdentityFeedbackAttributesInput{
		EmailIdentity: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.SESV2, create.ErrActionDeleting, ResNameEmailIdentityFeedbackAttributes, d.Id(), err)
	}

	return nil
}
