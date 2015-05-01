package aws

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsOpsWorksCustomLayer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsWorksCustomLayerCreate,
		Read:   resourceAwsOpsWorksCustomLayerRead,
		Update: resourceAwsOpsWorksCustomLayerUpdate,
		Delete: resourceAwsOpsWorksCustomLayerDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"short_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"stack_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"auto_assign_elastic_ips": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"auto_assign_public_ips": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"custom_instance_profile_arn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_security_group_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set:      schema.HashString,
			},
			"enable_auto_healing": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"install_updates_on_boot": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"packages": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Set:      schema.HashString,
			},
			"shutdown_event_configuration": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delay_until_elb_connections_drained": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"execution_timeout": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  120,
						},
					},
				},
				Set: resourceAwsOpsWorksShutdownEventConfigHash,
			},
			"volume_configuration": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mount_point": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"num_disks": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"size": &schema.Schema{
							Type:     schema.TypeInt, // in GB
							Required: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true, // API specifies this as optional but it does not make sense
						},
						"iops": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
						},
						"raid_level": &schema.Schema{
							Type:     schema.TypeString, // 0 is valid so we need to distingush empty value and zero...
							Optional: true,
						},
					},
				},
				Set: resourceAwsOpsWorksVolumeConfigHash,
			},
		},
	}
}

func resourceAwsOpsWorksCustomLayerCreate(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	createOpts := &opsworks.CreateLayerInput{
		Name:      aws.String(d.Get("name").(string)),
		Shortname: aws.String(d.Get("short_name").(string)),
		StackID:   aws.String(d.Get("stack_id").(string)),
		Type:      aws.String("custom"),
	}

	// When use_opsworks_security_groups = false in the stack,
	// we need to set custom_security_group_ids here.
	createOpts.CustomSecurityGroupIDs = []*string{}
	if v, found := d.GetOk("custom_security_group_ids"); found {
		for _, sg := range v.(*schema.Set).List() {
			createOpts.CustomSecurityGroupIDs = append(
				createOpts.CustomSecurityGroupIDs, aws.String(sg.(string)))
		}
	}

	resp, err := opsworksconn.CreateLayer(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating layer: %s", err)
	}

	d.SetId(*resp.LayerID)

	return resourceAwsOpsWorksCustomLayerUpdate(d, meta)
}

func resourceAwsOpsWorksCustomLayerRead(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	describeOpts := &opsworks.DescribeLayersInput{
		LayerIDs: []*string{aws.String(d.Id())},
	}
	resp, err := opsworksconn.DescribeLayers(describeOpts)
	if err != nil {
		if opserr, ok := err.(aws.APIError); ok && opserr.Code == "ResourceNotFoundException" {
			// layer is gone now
			d.SetId("")
			return nil
		}
		return err
	}
	if len(resp.Layers) < 1 {
		// layer is gone now
		d.SetId("")
		return nil
	}

	layer := resp.Layers[0]
	log.Printf("[DEBUG] Layer: %#v", layer)

	d.Set("auto_assign_elastic_ips", *layer.AutoAssignElasticIPs)
	d.Set("auto_assign_public_ips", *layer.AutoAssignPublicIPs)
	d.Set("enable_auto_healing", *layer.EnableAutoHealing)

	d.Set("name", *layer.Name)
	if layer.Name != nil {
		d.Set("short_name", *layer.Shortname)
	}

	if layer.CustomInstanceProfileARN != nil {
		d.Set("custom_instance_profile_arn", *layer.CustomInstanceProfileARN)
	}

	if layer.InstallUpdatesOnBoot != nil {
		d.Set("install_updates_on_boot", *layer.InstallUpdatesOnBoot)
	}

	customSGs := make([]string, 0, len(layer.CustomSecurityGroupIDs))
	for _, sg := range layer.CustomSecurityGroupIDs {
		customSGs = append(customSGs, *sg)
	}
	d.Set("custom_security_group_ids", customSGs)

	packages := make([]string, 0, len(layer.Packages))
	for _, pkg := range layer.Packages {
		packages = append(packages, *pkg)
	}
	d.Set("packages", packages)

	// we don't want set value if we don't have shutdown_event_configuration
	if _, found := d.GetOk("shutdown_event_configuration"); found {
		shutdownConfig := layer.LifecycleEventConfiguration.Shutdown
		d.Set("shutdown_event_configuration", []map[string]interface{}{
			{
				"delay_until_elb_connections_drained": *shutdownConfig.DelayUntilELBConnectionsDrained,
				"execution_timeout":                   int(*shutdownConfig.ExecutionTimeout),
			},
		})
	}

	volumes := make([]map[string]interface{}, 0, len(layer.VolumeConfigurations))
	for _, v := range layer.VolumeConfigurations {
		volume := map[string]interface{}{
			"mount_point": *v.MountPoint,
			"num_disks":   int(*v.NumberOfDisks),
			"size":        int(*v.Size),
			"type":        *v.VolumeType,
		}
		if v.IOPS != nil {
			volume["iops"] = int(*v.IOPS)
		}
		if v.RAIDLevel != nil {
			volume["raid_level"] = int(*v.RAIDLevel)
		}
		volumes = append(volumes, volume)
	}
	d.Set("volume_configuration", volumes)

	return nil
}

func resourceAwsOpsWorksCustomLayerUpdate(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	updateOpts := &opsworks.UpdateLayerInput{
		AutoAssignElasticIPs:     aws.Boolean(d.Get("auto_assign_elastic_ips").(bool)),
		AutoAssignPublicIPs:      aws.Boolean(d.Get("auto_assign_public_ips").(bool)),
		CustomInstanceProfileARN: aws.String(d.Get("custom_instance_profile_arn").(string)),
		EnableAutoHealing:        aws.Boolean(d.Get("enable_auto_healing").(bool)),
		InstallUpdatesOnBoot:     aws.Boolean(d.Get("install_updates_on_boot").(bool)),
		LayerID:                  aws.String(d.Id()),
		Name:                     aws.String(d.Get("name").(string)),
	}
	if v, found := d.GetOk("short_name"); found {
		updateOpts.Shortname = aws.String(v.(string))
	}

	updateOpts.CustomSecurityGroupIDs = []*string{}
	if v, found := d.GetOk("custom_security_group_ids"); found {
		for _, sg := range v.(*schema.Set).List() {
			updateOpts.CustomSecurityGroupIDs = append(
				updateOpts.CustomSecurityGroupIDs, aws.String(sg.(string)))
		}
	}

	updateOpts.Packages = []*string{}
	if v, found := d.GetOk("packages"); found {
		for _, pkg := range v.(*schema.Set).List() {
			updateOpts.Packages = append(updateOpts.Packages, aws.String(pkg.(string)))
		}
	}

	if d.HasChange("shutdown_event_configuration") {
		shutdownConfig := &opsworks.ShutdownEventConfiguration{}
		vs := d.Get("shutdown_event_configuration").(*schema.Set).List()
		if len(vs) > 0 {
			config := vs[0].(map[string]interface{})
			shutdownConfig.DelayUntilELBConnectionsDrained = aws.Boolean(config["delay_until_elb_connections_drained"].(bool))
			shutdownConfig.ExecutionTimeout = aws.Long(int64(config["execution_timeout"].(int)))
		} else {
			// resetting to default value
			shutdownConfig.DelayUntilELBConnectionsDrained = aws.Boolean(false)
			shutdownConfig.ExecutionTimeout = aws.Long(int64(120))
		}
		updateOpts.LifecycleEventConfiguration = &opsworks.LifecycleEventConfiguration{
			Shutdown: shutdownConfig,
		}
	}

	if d.HasChange("volume_configuration") {
		vcs := []*opsworks.VolumeConfiguration{}
		vs := d.Get("volume_configuration").(*schema.Set).List()
		for _, v := range vs {
			volume := v.(map[string]interface{})
			vc := &opsworks.VolumeConfiguration{
				MountPoint:    aws.String(volume["mount_point"].(string)),
				NumberOfDisks: aws.Long(int64(volume["num_disks"].(int))),
				Size:          aws.Long(int64(volume["size"].(int))),
				VolumeType:    aws.String(volume["type"].(string)),
			}
			if v, ok := volume["iops"].(int); ok && v > 0 {
				vc.IOPS = aws.Long(int64(v))
			}
			if v, ok := volume["raid_level"].(string); ok && v != "" {
				level, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				vc.RAIDLevel = aws.Long(level)
			}
			log.Printf("[DEBUG] VolumeConfiguration: %#v", vc)
			vcs = append(vcs, vc)
		}
		updateOpts.VolumeConfigurations = vcs
	}

	log.Printf("[DEBUG] Layer: %#v", updateOpts)

	if _, err := opsworksconn.UpdateLayer(updateOpts); err != nil {
		return fmt.Errorf("Error updating stack: %s", err)
	}
	return resourceAwsOpsWorksCustomLayerRead(d, meta)
}

func resourceAwsOpsWorksCustomLayerDelete(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	log.Printf("[INFO] Deleting layer: %s", d.Id())

	deleteOpts := &opsworks.DeleteLayerInput{
		LayerID: aws.String(d.Id()),
	}
	if _, err := opsworksconn.DeleteLayer(deleteOpts); err != nil {
		return fmt.Errorf("Error deleting stack: %s", err)
	}

	return nil
}

func resourceAwsOpsWorksShutdownEventConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%t-", m["delay_until_elb_connections_drained"].(bool)))
	buf.WriteString(fmt.Sprintf("%d-", m["execution_timeout"].(int)))
	return hashcode.String(buf.String())
}

func resourceAwsOpsWorksVolumeConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["mount_point"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["num_disks"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["size"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["type"].(string)))
	if v, found := m["iops"]; found {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, found := m["raid_level"]; found {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}
