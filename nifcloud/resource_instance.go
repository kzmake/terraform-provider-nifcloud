package nifcloud

import (
	"fmt"
	"github.com/alice02/nifcloud-sdk-go/nifcloud"
	"github.com/alice02/nifcloud-sdk-go/nifcloud/awserr"
	"github.com/alice02/nifcloud-sdk-go/service/computing"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"time"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceCreate,
		Read:   resourceInstanceRead,
		Update: resourceInstanceUpdate,
		Delete: resourceInstanceDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 15),
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"accounting_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2",
			},
			"admin": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ip_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"agreement": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network_interfaces": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipaddress": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"license": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"license_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"license_num": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	var securityGroups []*string
	if v := d.Get("security_groups"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			securityGroups = append(securityGroups, nifcloud.String(v.(string)))
		}
	}

	var networkInterfaces []*computing.RequestNetworkInterfaceStruct
	if interfaces, ok := d.GetOk("network_interfaces"); ok {
		for _, ni := range interfaces.(*schema.Set).List() {
			networkInterface := &computing.RequestNetworkInterfaceStruct{}
			if v, ok := ni.(*schema.ResourceData).GetOk("network_id"); ok {
				networkInterface.SetNetworkId(v.(string))
			}
			if v, ok := ni.(*schema.ResourceData).GetOk("network_name"); ok {
				networkInterface.SetNetworkName(v.(string))
			}
			if v, ok := ni.(*schema.ResourceData).GetOk("ipaddress"); ok {
				networkInterface.SetIpAddress(v.(string))
			}

			networkInterfaces = append(networkInterfaces, networkInterface)
		}
	}

	var licenses []*computing.RequestLicenseStruct
	if licensesSet, ok := d.GetOk("license"); ok {
		for _, l := range licensesSet.(*schema.Set).List() {
			license := &computing.RequestLicenseStruct{}
			if v, ok := l.(*schema.ResourceData).GetOk("license_name"); ok {
				license.SetLicenseName(v.(string))
			}
			if v, ok := l.(*schema.ResourceData).GetOk("license_num"); ok {
				license.SetLicenseNum(v.(string))
			}

			licenses = append(licenses, license)
		}
	}

	input := computing.RunInstancesInput{
		InstanceId:            nifcloud.String(d.Get("instance_id").(string)),
		ImageId:               nifcloud.String(d.Get("image_id").(string)),
		KeyName:               nifcloud.String(d.Get("key_name").(string)),
		SecurityGroup:         securityGroups,
		UserData:              nifcloud.String(d.Get("user_data").(string)),
		InstanceType:          nifcloud.String(d.Get("instance_type").(string)),
		Placement:             &computing.RequestPlacementStruct{AvailabilityZone: nifcloud.String(d.Get("availability_zone").(string))},
		DisableApiTermination: nifcloud.Bool(d.Get("disable_api_termination").(bool)),
		AccountingType:        nifcloud.String(d.Get("accounting_type").(string)),
		Admin:                 nifcloud.String(d.Get("admin").(string)),
		Password:              nifcloud.String(d.Get("password").(string)),
		IpType:                nifcloud.String(d.Get("ip_type").(string)),
		PublicIp:              nifcloud.String(d.Get("public_ip").(string)),
		Agreement:             nifcloud.Bool(d.Get("agreement").(bool)),
		Description:           nifcloud.String(d.Get("description").(string)),
		NetworkInterface:      networkInterfaces,
		License:               licenses,
	}

	out, err := conn.RunInstances(&input)
	if err != nil {
		return fmt.Errorf("Error RunInstancesInput: %s", err)
	}
	d.SetId(*out.InstancesSet[0].InstanceId)

	return resourceInstanceRead(d, meta)
}

func resourceInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	stopInstancesInput := computing.StopInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Id())},
	}
	if _, err := conn.StopInstances(&stopInstancesInput); err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Server.ProcessingFailure.Instance.Stop" {
			// 何もしないで継続
		} else {
			return fmt.Errorf("Error StopInstances: %s", err)
		}
	}

	terminateInstancesInput := computing.TerminateInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Id())},
	}
	if _, err := conn.TerminateInstances(&terminateInstancesInput); err != nil {
		return fmt.Errorf("Error TerminateInstances: %s", err)
	}

	return nil
}

func resourceInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.ModifyInstanceAttributeInput{
		InstanceId: nifcloud.String(d.Id()),
		Attribute:  nifcloud.String("description"),
		Value:      nifcloud.String(d.Get("description").(string)),
	}

	_, err := conn.ModifyInstanceAttribute(&input)
	if err != nil {
		return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
	}

	return resourceInstanceRead(d, meta)
}

func resourceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.DescribeInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Id())},
	}

	out, err := conn.DescribeInstances(&input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Instance: %s", err)
	}

	inputDisableApiTermination := computing.DescribeInstanceAttributeInput{
		InstanceId: nifcloud.String(d.Id()),
		Attribute:  nifcloud.String("disableApiTermination"),
	}

	outDisableApiTermination, err := conn.DescribeInstanceAttribute(&inputDisableApiTermination)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Instance: %s", err)
	}

	inputUserData := computing.DescribeInstanceAttributeInput{
		InstanceId: nifcloud.String(d.Id()),
		Attribute:  nifcloud.String("userData"),
	}

	outUserData, err := conn.DescribeInstanceAttribute(&inputUserData)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Instance: %s", err)
	}

	for _, instance := range out.ReservationSet[0].InstancesSet {
		if *instance.InstanceId == d.Id() {
			d.Set("instance_id", instance.InstanceId)
			d.Set("image_id", instance.ImageId)
			d.Set("instance_type", instance.InstanceType)
			d.Set("accounting_type", instance.AccountingType)
			d.Set("description", instance.Description)
			d.Set("availability_zone", instance.Placement.AvailabilityZone)
			d.Set("disable_api_termination", outDisableApiTermination.DisableApiTermination)
			d.Set("user_data", outUserData.UserData)

			// only windows
			d.Set("admin", instance.Admin)

			// only linux
			d.Set("key_name", instance.KeyName)

			for _, v := range out.ReservationSet[0].GroupSet {
				d.Set("security_groups", v.GroupId)
			}

			return nil
		}
	}

	return fmt.Errorf("Unable to find instance within: %#v", out.ReservationSet[0].InstancesSet)
}
