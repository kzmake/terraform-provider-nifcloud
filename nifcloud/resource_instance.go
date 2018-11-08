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
	"strconv"
	"time"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceCreate,
		Read:   resourceInstanceRead,
		Update: resourceInstanceUpdate,
		Delete: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				conn := meta.(*NifcloudClient).computingconn
				out, _ := conn.DescribeInstances(&computing.DescribeInstancesInput{})
				for _, i := range out.ReservationSet[0].InstancesSet {
					if *i.InstanceUniqueId == d.Id() {
						d.Set("name", i.InstanceId)
					}
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
				Default:  "static",
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
			"instance_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	var securityGroups []*string
	if sgs := d.Get("security_groups").([]interface{}); sgs != nil {
		for _, v := range sgs {
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
		InstanceId:            nifcloud.String(d.Get("name").(string)),
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

	instance := out.InstancesSet[0]

	log.Printf("[INFO] Instance Id: %s", *instance.InstanceId)

	d.SetId(*instance.InstanceUniqueId)
	d.Set("name", instance.InstanceId)

	log.Printf("[DEBUG] Waiting for instance (%s) to become running", *instance.InstanceId)

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"running"},
		Refresh:    InstanceStateRefreshFunc(meta, *instance.InstanceId, []string{"warning", "terminated"}),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to become ready: %s",
			*instance.InstanceId, err)
	}

	return resourceInstanceRead(d, meta)
}

func resourceInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	stopInstancesInput := computing.StopInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Get("name").(string))},
	}
	if _, err := conn.StopInstances(&stopInstancesInput); err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Server.ProcessingFailure.Instance.Stop" {
			// 何もしないで継続
		} else {
			return fmt.Errorf("Error StopInstances: %s", err)
		}
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become stopped", d.Id())

	stopStateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running"},
		Target:     []string{"stopped"},
		Refresh:    InstanceStateRefreshFunc(meta, d.Get("name").(string), []string{"warning"}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err := stopStateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to stopped: %s", d.Id(), err)
	}

	terminateInstancesInput := computing.TerminateInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Get("name").(string))},
	}
	if _, err := conn.TerminateInstances(&terminateInstancesInput); err != nil {
		return fmt.Errorf("Error TerminateInstances: %s", err)
	}

	log.Printf("[DEBUG] Waiting for instance (%s) to become terminate", d.Id())

	terminateStateConf := &resource.StateChangeConf{
		Pending:    []string{"pending", "running", "stopped"},
		Target:     []string{"terminated"},
		Refresh:    InstanceStateRefreshFunc(meta, d.Get("name").(string), []string{"warning"}),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err := terminateStateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to terminate: %s", d.Id(), err)
	}

	return nil
}

func resourceInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	updateStateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"running", "stopped"},
		Refresh:    InstanceStateRefreshFunc(meta, d.Get("name").(string), []string{"warning", "terminated"}),
		Timeout:    d.Timeout(schema.TimeoutUpdate),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if d.HasChange("description") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("description"),
			Value:      nifcloud.String(d.Get("description").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("instance_type") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("instanceType"),
			Value:      nifcloud.String(d.Get("instance_type").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("disable_api_termination") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("disableApiTermination"),
			Value:      nifcloud.String(strconv.FormatBool(d.Get("disable_api_termination").(bool))),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("name") {
		before, after := d.GetChange("name")
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(before.(string)),
			Attribute:  nifcloud.String("instanceName"),
			Value:      nifcloud.String(after.(string)),
		})

		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		updateStateConf := &resource.StateChangeConf{
			Pending:    []string{"pending", "terminated"},
			Target:     []string{"running", "stopped"},
			Refresh:    InstanceStateRefreshFunc(meta, d.Get("name").(string), []string{"warning"}),
			Timeout:    d.Timeout(schema.TimeoutUpdate),
			Delay:      10 * time.Second,
			MinTimeout: 5 * time.Second,
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("accounting_type") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("accountingType"),
			Value:      nifcloud.String(d.Get("accounting_type").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("security_groups") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("groupId"),
			Value:      nifcloud.String(d.Get("security_groups").([]interface{})[0].(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	if d.HasChange("ip_type") {
		_, err := conn.ModifyInstanceAttribute(&computing.ModifyInstanceAttributeInput{
			InstanceId: nifcloud.String(d.Get("name").(string)),
			Attribute:  nifcloud.String("ipType"),
			Value:      nifcloud.String(d.Get("ip_type").(string)),
		})
		if err != nil {
			return fmt.Errorf("Error ModifyInstanceAttribute: %s", err)
		}

		if _, err := updateStateConf.WaitForState(); err != nil {
			return fmt.Errorf(
				"Error waiting for instance (%s) to become ready: %s",
				d.Id(), err)
		}
	}

	return resourceInstanceRead(d, meta)
}

func resourceInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*NifcloudClient).computingconn

	input := computing.DescribeInstancesInput{
		InstanceId: []*string{nifcloud.String(d.Get("name").(string))},
	}

	out, err := conn.DescribeInstances(&input)
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Couldn't find Instance resource: %s", err)
	}

	return setInstanceResourceData(d, meta, out)
}

func InstanceStateRefreshFunc(meta interface{}, instanceId string, failStates []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*NifcloudClient).computingconn

		input := computing.DescribeInstancesInput{
			InstanceId: []*string{nifcloud.String(instanceId)},
		}

		out, err := conn.DescribeInstances(&input)

		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
				return "", "terminated", nil
			} else {
				log.Printf("Error on InstanceStateRefresh: %s", err)
				return nil, "", err
			}
		}

		instance := out.ReservationSet[0].InstancesSet[0]
		state := *instance.InstanceState.Name

		for _, failState := range failStates {
			if state == failState {
				return instance, state, fmt.Errorf("Failed to reach target state. Reason: %s", state)
			}
		}

		return instance, state, nil
	}
}

func setInstanceResourceData(d *schema.ResourceData, meta interface{}, out *computing.DescribeInstancesOutput) error {
	conn := meta.(*NifcloudClient).computingconn

	instance := out.ReservationSet[0].InstancesSet[0]

	outDisableApiTermination, err := conn.DescribeInstanceAttribute(&computing.DescribeInstanceAttributeInput{
		InstanceId: nifcloud.String(d.Get("name").(string)),
		Attribute:  nifcloud.String("disableApiTermination"),
	})

	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Instance: %s", err)
	}

	outUserData, err := conn.DescribeInstanceAttribute(&computing.DescribeInstanceAttributeInput{
		InstanceId: nifcloud.String(d.Get("name").(string)),
		Attribute:  nifcloud.String("userData"),
	})

	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "Client.InvalidParameterNotFound.Instance" {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Instance: %s", err)
	}



	d.Set("name", instance.InstanceId)
	d.Set("image_id", instance.ImageId)
	d.Set("instance_type", instance.InstanceType)
	d.Set("accounting_type", instance.AccountingType)
	d.Set("description", instance.Description)
	d.Set("availability_zone", instance.Placement.AvailabilityZone)
	d.Set("user_data", outUserData.UserData)
	d.Set("ip_type", instance.IpType)

	disableApiTermination, _ := strconv.ParseBool(*outDisableApiTermination.DisableApiTermination.Value)
	d.Set("disable_api_termination", disableApiTermination)

	// only windows
	d.Set("admin", instance.Admin)

	// only linux
	d.Set("key_name", instance.KeyName)

	d.Set("instance_state", instance.InstanceState.Name)

	sgs := make([]string, 0, len(out.ReservationSet[0].GroupSet))
	for _, sg := range out.ReservationSet[0].GroupSet {
		sgs = append(sgs, *sg.GroupId)
	}

	log.Printf("[DEBUG] Setting Security Group Ids: %#v", sgs)
	if err := d.Set("security_groups", sgs); err != nil {
		return err
	}

	return nil
}
