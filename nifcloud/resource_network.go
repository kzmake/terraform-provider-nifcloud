package nifcloud

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/kzmake/nifcloud-sdk-go/nifcloud"
	"github.com/kzmake/nifcloud-sdk-go/nifcloud/awserr"
	"github.com/kzmake/nifcloud-sdk-go/service/computing"
	"log"
	"time"
)

func resourceNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkCreate,
		Read:   resourceNetworkRead,
		Update: resourceNetworkUpdate,
		Delete: resourceNetworkDelete,
		Importer: &schema.ResourceImporter{},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 15),
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"accounting_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2",
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.NiftyCreatePrivateLanInput{
		PrivateLanName:   nifcloud.String(d.Get("name").(string)),
		CidrBlock:        nifcloud.String(d.Get("cidr_block").(string)),
		AvailabilityZone: nifcloud.String(d.Get("availability_zone").(string)),
		AccountingType:   nifcloud.String(d.Get("accounting_type").(string)),
		Description:      nifcloud.String(d.Get("description").(string)),
	}

	out, err := conn.NiftyCreatePrivateLan(&input)
	if err != nil {
		return fmt.Errorf("Error NiftyCreatePrivateLanInput: %s", err)
	}

	network := out.PrivateLan

	log.Printf("[INFO] Network Id: %s", *network.NetworkId)

	d.SetId(*network.NetworkId)

	log.Printf("[DEBUG] Waiting for (%s) to become running", *network.NetworkId)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"running"},
		Refresh:    NetworkStateRefreshFunc(meta, *network.NetworkId, []string{"terminated"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for (%s) to become ready: %s",
			*network.NetworkId, err)
	}

	return resourceInstanceRead(d, meta)
}

func resourceNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.NiftyDeletePrivateLanInput{
		NetworkId: nifcloud.String(d.Id()),
	}

	if _, err := conn.NiftyDeletePrivateLan(&input); err != nil {
		return fmt.Errorf("Error NiftyDeletePrivateLanInput: %s", err)
	}

	log.Printf("[DEBUG] Waiting for (%s) to become terminate", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"terminated"},
		Refresh:    NetworkStateRefreshFunc(meta, d.Id(), []string{"available"}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for (%s) to terminate: %s", d.Id(), err)
	}

	return nil
}

func resourceNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	updateStateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"available"},
		Refresh:    NetworkStateRefreshFunc(meta, d.Id(), []string{"terminated"}),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if d.HasChange("description") {
		_, err := conn.NiftyModifyPrivateLanAttribute(&computing.NiftyModifyPrivateLanAttributeInput{
			NetworkId: nifcloud.String(d.Id()),
			Attribute: nifcloud.String("description"),
			Value:     nifcloud.String(d.Get("description").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error NiftyModifyPrivateLanAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("name") {
		_, err := conn.NiftyModifyPrivateLanAttribute(&computing.NiftyModifyPrivateLanAttributeInput{
			NetworkId: nifcloud.String(d.Id()),
			Attribute: nifcloud.String("privateLanName"),
			Value:     nifcloud.String(d.Get("name").(string)),
		})

		if err != nil {
			return fmt.Errorf("Error NiftyModifyPrivateLanAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("accounting_type") {
		_, err := conn.NiftyModifyPrivateLanAttribute(&computing.NiftyModifyPrivateLanAttributeInput{
			NetworkId: nifcloud.String(d.Id()),
			Attribute: nifcloud.String("accountingType"),
			Value:     nifcloud.String(d.Get("accounting_type").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("cidr_block") {
		_, err := conn.NiftyModifyPrivateLanAttribute(&computing.NiftyModifyPrivateLanAttributeInput{
			NetworkId: nifcloud.String(d.Id()),
			Attribute: nifcloud.String("cidrBlock"),
			Value:     nifcloud.String(d.Get("cidr_block").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	return resourceInstanceRead(d, meta)
}

func resourceNetworkRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.NiftyDescribePrivateLansInput{
		NetworkId: []*string{nifcloud.String(d.Id())},
	}

	out, err := conn.NiftyDescribePrivateLans(&input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.NetworkId" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find Instance resource: %s", err)
	}

	return setNetworkResourceData(d, meta, out)
}

func NetworkStateRefreshFunc(meta interface{}, networkId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*NifcloudClient).computingconn

		input := computing.NiftyDescribePrivateLansInput{
			NetworkId: []*string{nifcloud.String(networkId)},
		}

		out, err := conn.NiftyDescribePrivateLans(&input)

		if err != nil {
			awsErr, ok := err.(awserr.Error);
			if ok && awsErr.Code() == "Client.InvalidParameterNotFound.NetworkId" {
				return "", "terminated", nil
			} else {
				log.Printf("Error on InstanceStateRefresh: %s", err)
				return nil, "", err
			}
		}

		network := out.PrivateLanSet[0]
		state := *network.State

		for _, failState := range failStates {
			if state == failState {
				return network, state, fmt.Errorf("Failed to reach target state. Reason: %s", state)
			}
		}

		return network, state, nil
	}
}

func setNetworkResourceData(d *schema.ResourceData, meta interface{}, out *computing.NiftyDescribePrivateLansOutput) error {
	network := out.PrivateLanSet[0]

	d.Set("network_id", network.NetworkId)
	d.Set("name", network.PrivateLanName)
	d.Set("cidr_block", network.CidrBlock)
	d.Set("availability_zone", network.AvailabilityZone)
	d.Set("accounting_type", network.AccountingType)
	d.Set("description", network.Description)
	d.Set("state", network.State)

	return nil
}
