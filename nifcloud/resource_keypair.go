package nifcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/alice02/nifcloud-sdk-go/nifcloud"
	"github.com/alice02/nifcloud-sdk-go/nifcloud/awserr"
	"github.com/alice02/nifcloud-sdk-go/service/computing"
)

func resourceKeyPair() *schema.Resource {
	return &schema.Resource{
		Create: resourceKeyPairCreate,
		Read:   resourceKeyPairRead,
		Update: resourceKeyPairUpdate,
		Delete: resourceKeyPairDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(6, 32),
			},
			"public_key_material": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 40),
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKeyPairCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.ImportKeyPairInput{
		KeyName:           nifcloud.String(d.Get("name").(string)),
		PublicKeyMaterial: nifcloud.String(d.Get("public_key_material").(string)),
		Description:       nifcloud.String(d.Get("description").(string)),
	}

	out, err := conn.ImportKeyPair(&input)
	if err != nil {
		return fmt.Errorf("Error ImportKeyPair: %s", err)
	}
	d.SetId(*out.KeyName)
	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.DeleteKeyPairInput{
		KeyName: nifcloud.String(d.Id()),
	}

	_, err := conn.DeleteKeyPair(&input)
	if err != nil {
		return fmt.Errorf("Error DeleteKeyPair: %s", err)
	}

	return err
}

func resourceKeyPairUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.NiftyModifyKeyPairAttributeInput{
		KeyName:   nifcloud.String(d.Id()),
		Attribute: nifcloud.String("description"),
		Value:     nifcloud.String(d.Get("description").(string)),
	}

	_, err := conn.NiftyModifyKeyPairAttribute(&input)
	if err != nil {
		return fmt.Errorf("Error NiftyModifyKeyPairAttribute: %s", err)
	}

	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := &computing.DescribeKeyPairsInput{
		KeyName: []*string{nifcloud.String(d.Id())},
	}

	out, err := conn.DescribeKeyPairs(input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "InvalidKeyPair.NotFound" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving KeyPair: %s", err)
	}

	for _, key := range out.KeySet {
		if *key.KeyName == d.Id() {
			d.Set("key_name", key.KeyName)
			d.Set("fingerprint", key.KeyFingerprint)
			d.Set("description", key.Description)
			return nil
		}
	}

	return fmt.Errorf("Unable to find key pair within: %#v", out.KeySet)
}
