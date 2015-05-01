package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

// TODO(nabeken):
//   - SSH Key in opsworks.Source (waiting for vault integration)
//     But AWS will mask the value like `"SshKey": "*****FILTERED*****"`
func resourceAwsOpsWorksStack() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsWorksStackCreate,
		Read:   resourceAwsOpsWorksStackRead,
		Update: resourceAwsOpsWorksStackUpdate,
		Delete: resourceAwsOpsWorksStackDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"region": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"default_availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_instance_profile_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"default_os": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_root_device_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"default_ssh_key_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"berkshelf_version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"chef_version": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				// FIXME(nabeken): Document For Berkshelf, we need at least 11.10
			},
			"custom_json": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname_theme": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"manage_berkshelf": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"service_role_arn": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"use_custom_cookbooks": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"use_opsworks_security_groups": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cookbook_source": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"revision": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsOpsWorksSourceHash,
			},
		},
	}
}

func resourceAwsOpsWorksStackCreate(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	stackRegion := meta.(*AWSClient).region
	if region, found := d.GetOk("region"); found {
		stackRegion = region.(string)
	}

	createOpts := &opsworks.CreateStackInput{
		DefaultInstanceProfileARN: aws.String(d.Get("default_instance_profile_arn").(string)),
		Name:           aws.String(d.Get("name").(string)),
		Region:         aws.String(stackRegion),
		ServiceRoleARN: aws.String(d.Get("service_role_arn").(string)),
	}

	inVPC := false
	if v, found := d.GetOk("vpc_id"); found {
		createOpts.VPCID = aws.String(v.(string))
		createOpts.DefaultSubnetID = aws.String(d.Get("default_subnet_id").(string))
		inVPC = true
	}

	// Retry if we get ValidationException looks like the below:
	// Service Role Arn: arn:aws:iam::1234567890:role/tf-acc-aws-opsworks-vpc-service-role is not yet propagated, please try again in a couple of minutes
	var resp *opsworks.CreateStackOutput
	err := resource.Retry(20*time.Minute, func() error {
		var cerr error
		resp, cerr = opsworksconn.CreateStack(createOpts)
		if cerr != nil {
			if opserr, ok := cerr.(aws.APIError); ok {
				// I know this is a ugly error checking but AWS does not provide a code for this case.
				if opserr.Code == "ValidationException" && strings.Contains(opserr.Message, "is not yet propagated") {
					log.Printf("[INFO] Waiting for Service Role to be propagated: %s", cerr)
					return cerr
				}
			}
			return resource.RetryError{Err: cerr}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error creating stack: %s", err)
	}

	d.SetId(*resp.StackID)

	// Wait for several seconds until all built-in security groups are created
	// when the stack is in VPC. We can't wait for each SG because OpsWorks does not
	// specify what built-in security groups are.
	if inVPC {
		log.Print("[INFO] Waiting for built-in security groups created")
		time.Sleep(30 * time.Second)
	}
	return resourceAwsOpsWorksStackUpdate(d, meta)
}

func resourceAwsOpsWorksStackRead(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	describeOpts := &opsworks.DescribeStacksInput{
		StackIDs: []*string{aws.String(d.Id())},
	}
	resp, err := opsworksconn.DescribeStacks(describeOpts)
	if err != nil {
		if opserr, ok := err.(aws.APIError); ok && opserr.Code == "ResourceNotFoundException" {
			// stack is gone now
			d.SetId("")
			return nil
		}
		return err
	}
	if len(resp.Stacks) < 1 {
		// stack is gone now
		d.SetId("")
		return nil
	}

	stack := resp.Stacks[0]
	log.Printf("[DEBUG] Stack: %#v", stack)

	d.Set("chef_version", *stack.ConfigurationManager.Version)
	d.Set("default_availability_zone", *stack.DefaultAvailabilityZone)
	d.Set("default_instance_profile_arn", *stack.DefaultInstanceProfileARN)
	d.Set("default_os", *stack.DefaultOs)
	d.Set("default_root_device_type", *stack.DefaultRootDeviceType)
	d.Set("hostname_theme", *stack.HostnameTheme)
	d.Set("region", *stack.Region)
	d.Set("service_role_arn", *stack.ServiceRoleARN)
	d.Set("use_custom_cookbooks", *stack.UseCustomCookbooks)
	d.Set("use_opsworks_security_groups", *stack.UseOpsWorksSecurityGroups)

	if stack.CustomJSON != nil {
		d.Set("custom_json", *stack.CustomJSON)
	}
	if stack.DefaultSubnetID != nil {
		d.Set("default_subnet_id", *stack.DefaultSubnetID)
	}
	if stack.DefaultSSHKeyName != nil {
		d.Set("default_ssh_key_name", *stack.DefaultSSHKeyName)
	}
	if c := stack.ChefConfiguration; c != nil {
		if c.BerkshelfVersion != nil {
			d.Set("berkshelf_version", *c.BerkshelfVersion)
		}
		if c.ManageBerkshelf != nil {
			d.Set("manage_berkshelf", *c.ManageBerkshelf)
		}
	}

	log.Printf("[DEBUG] Stack CustomCookbooksSource: %#v", *stack.CustomCookbooksSource)

	if c := stack.CustomCookbooksSource; c != nil {
		source := map[string]interface{}{}
		if c.Type != nil {
			source["type"] = *c.Type
		}
		if c.URL != nil {
			source["url"] = *c.URL
		}
		if c.Revision != nil {
			source["revision"] = *c.Revision
		}
		if len(source) > 0 {
			d.Set("cookbook_source", []map[string]interface{}{source})
		}
	}

	return nil
}

func resourceAwsOpsWorksStackUpdate(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	updateOpts := &opsworks.UpdateStackInput{
		CustomJSON:                aws.String(d.Get("custom_json").(string)),
		DefaultAvailabilityZone:   aws.String(d.Get("default_availability_zone").(string)),
		DefaultInstanceProfileARN: aws.String(d.Get("default_instance_profile_arn").(string)),
		DefaultRootDeviceType:     aws.String(d.Get("default_root_device_type").(string)),
		DefaultSSHKeyName:         aws.String(d.Get("default_ssh_key_name").(string)),
		Name:                      aws.String(d.Get("name").(string)),
		ServiceRoleARN:            aws.String(d.Get("service_role_arn").(string)),
		StackID:                   aws.String(d.Id()),
		UseCustomCookbooks:        aws.Boolean(d.Get("use_custom_cookbooks").(bool)),
		UseOpsWorksSecurityGroups: aws.Boolean(d.Get("use_opsworks_security_groups").(bool)),
	}
	if v, found := d.GetOk("default_os"); found {
		updateOpts.DefaultOs = aws.String(v.(string))
	}
	if v, found := d.GetOk("default_subnet_id"); found {
		updateOpts.DefaultSubnetID = aws.String(v.(string))
	}
	if v, found := d.GetOk("hostname_theme"); found {
		updateOpts.HostnameTheme = aws.String(v.(string))
	}
	updateOpts.ChefConfiguration = &opsworks.ChefConfiguration{
		BerkshelfVersion: aws.String(d.Get("berkshelf_version").(string)),
		ManageBerkshelf:  aws.Boolean(d.Get("manage_berkshelf").(bool)),
	}
	updateOpts.ConfigurationManager = &opsworks.StackConfigurationManager{
		Name:    aws.String("Chef"),
		Version: aws.String(d.Get("chef_version").(string)),
	}

	if d.HasChange("cookbook_source") {
		vs := d.Get("cookbook_source").(*schema.Set).List()

		// Set an empty value to reset
		customSource := &opsworks.Source{
			Revision: aws.String(""),
			Type:     aws.String(""),
			URL:      aws.String(""),
		}
		if len(vs) > 0 {
			source := vs[0].(map[string]interface{})
			customSource.Revision = aws.String(source["revision"].(string))
			customSource.Type = aws.String(source["type"].(string))
			customSource.URL = aws.String(source["url"].(string))
		}
		updateOpts.CustomCookbooksSource = customSource
	}

	if _, err := opsworksconn.UpdateStack(updateOpts); err != nil {
		return fmt.Errorf("Error updating stack: %s", err)
	}
	return resourceAwsOpsWorksStackRead(d, meta)
}

func resourceAwsOpsWorksStackDelete(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	log.Printf("[INFO] Deleting stack: %s", d.Id())

	deleteOpts := &opsworks.DeleteStackInput{
		StackID: aws.String(d.Id()),
	}
	if _, err := opsworksconn.DeleteStack(deleteOpts); err != nil {
		return fmt.Errorf("Error deleting stack: %s", err)
	}

	// If the stack were in VPC, OpsWorks automatically created several built-in
	// security groups in the VPC. When the stack is being removed,
	// these SGs is also going to be removed asynchronously.
	// If you are going to delete the VPC before deletion is completed,
	// it might cause a dependency problem.
	// Here, just wait for several seconds if the stack were in VPC.
	if _, found := d.GetOk("vpc_id"); found {
		log.Print("[INFO] Waiting for built-in security groups removed")
		time.Sleep(30 * time.Second)
	}
	return nil
}

func resourceAwsOpsWorksSourceHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["url"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["revision"].(string)))
	return hashcode.String(buf.String())
}
