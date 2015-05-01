package aws

import (
	"fmt"
	"log"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsOpsWorksCustomRecipes() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsOpsWorksCustomRecipesCreate,
		Update: resourceAwsOpsWorksCustomRecipesCreate,
		Read:   resourceAwsOpsWorksCustomRecipesRead,
		Delete: resourceAwsOpsWorksCustomRecipesDelete,

		Schema: map[string]*schema.Schema{
			"layer_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configure": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"deploy": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"setup": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"shutdown": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"undeploy": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAwsOpsWorksCustomRecipesRead(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	describeOpts := &opsworks.DescribeLayersInput{
		LayerIDs: []*string{aws.String(d.Get("layer_id").(string))},
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

	recipes := resp.Layers[0].CustomRecipes
	log.Printf("[DEBUG] CustomRecipes: %#v", recipes)

	configure := make([]string, 0, len(recipes.Configure))
	deploy := make([]string, 0, len(recipes.Deploy))
	setup := make([]string, 0, len(recipes.Setup))
	shutdown := make([]string, 0, len(recipes.Shutdown))
	undeploy := make([]string, 0, len(recipes.Undeploy))
	for _, r := range recipes.Configure {
		configure = append(configure, *r)
	}
	for _, r := range recipes.Deploy {
		deploy = append(deploy, *r)
	}
	for _, r := range recipes.Setup {
		setup = append(setup, *r)
	}
	for _, r := range recipes.Shutdown {
		shutdown = append(shutdown, *r)
	}
	for _, r := range recipes.Undeploy {
		undeploy = append(undeploy, *r)
	}

	d.Set("configure", configure)
	d.Set("deploy", deploy)
	d.Set("setup", setup)
	d.Set("shutdown", shutdown)
	d.Set("undeploy", undeploy)
	return nil
}

func resourceAwsOpsWorksCustomRecipesCreate(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	recipes := &opsworks.Recipes{
		Configure: []*string{},
		Deploy:    []*string{},
		Setup:     []*string{},
		Shutdown:  []*string{},
		Undeploy:  []*string{},
	}
	if v, found := d.GetOk("configure"); found {
		for _, r := range v.([]interface{}) {
			recipes.Configure = append(recipes.Configure, aws.String(r.(string)))
		}
	}
	if v, found := d.GetOk("deploy"); found {
		for _, r := range v.([]interface{}) {
			recipes.Deploy = append(recipes.Deploy, aws.String(r.(string)))
		}
	}
	if v, found := d.GetOk("setup"); found {
		for _, r := range v.([]interface{}) {
			recipes.Setup = append(recipes.Setup, aws.String(r.(string)))
		}
	}
	if v, found := d.GetOk("shutdown"); found {
		for _, r := range v.([]interface{}) {
			recipes.Shutdown = append(recipes.Shutdown, aws.String(r.(string)))
		}
	}
	if v, found := d.GetOk("undeploy"); found {
		for _, r := range v.([]interface{}) {
			recipes.Undeploy = append(recipes.Undeploy, aws.String(r.(string)))
		}
	}

	updateOpts := &opsworks.UpdateLayerInput{
		LayerID:       aws.String(d.Get("layer_id").(string)),
		CustomRecipes: recipes,
	}
	if _, err := opsworksconn.UpdateLayer(updateOpts); err != nil {
		return fmt.Errorf("Error updating stack: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:custom_recipes", *updateOpts.LayerID))
	return resourceAwsOpsWorksCustomRecipesRead(d, meta)
}

func resourceAwsOpsWorksCustomRecipesDelete(d *schema.ResourceData, meta interface{}) error {
	opsworksconn := meta.(*AWSClient).opsworksconn

	log.Printf("[INFO] Removing custom recipes from %s", d.Get("layer_id").(string))

	updateOpts := &opsworks.UpdateLayerInput{
		LayerID: aws.String(d.Get("layer_id").(string)),
		CustomRecipes: &opsworks.Recipes{
			Configure: []*string{},
			Deploy:    []*string{},
			Setup:     []*string{},
			Shutdown:  []*string{},
			Undeploy:  []*string{},
		},
	}
	if _, err := opsworksconn.UpdateLayer(updateOpts); err != nil {
		return fmt.Errorf("Error removing custom recipes: %s", err)
	}

	return nil
}
