package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// To run tests, we need predefined IAM roles such as `aws-opsworks-ec2-role` and `aws-opsworks-service-role`.

func TestAccAWSOpsWorksCustomRecipes(t *testing.T) {
	opsiam := testAccAWSOpsWorksIAM{}

	testAccAWSOpsWorksPopulateIAM(t, &opsiam)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksCustomLayerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccAWSOpsWorksCustomRecipesConfig, opsiam.ServiceRoleARN, opsiam.InstanceProfileARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.0", "tfacc::configure1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.1", "tfacc::configure2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.0", "tfacc::setup1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.1", "tfacc::setup2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.0", "tfacc::deploy1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.1", "tfacc::deploy2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.0", "tfacc::undeploy1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.1", "tfacc::undeploy2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.0", "tfacc::shutdown1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.1", "tfacc::shutdown2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccAWSOpsWorksCustomRecipesConfigUpdate, opsiam.ServiceRoleARN, opsiam.InstanceProfileARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.0", "tfacc::configure1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.1", "tfacc::configure2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "configure.2", "tfacc::configure3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.0", "tfacc::setup1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.1", "tfacc::setup2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "setup.2", "tfacc::setup3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.0", "tfacc::deploy1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.1", "tfacc::deploy2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "deploy.2", "tfacc::deploy3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.0", "tfacc::undeploy1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.1", "tfacc::undeploy2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "undeploy.2", "tfacc::undeploy3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.0", "tfacc::shutdown1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.1", "tfacc::shutdown2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_recipes.tf-acc", "shutdown.2", "tfacc::shutdown3"),
				),
			},
		},
	})
}

func testAccCheckAWSOpsWorksCustomRecipesDestroy(s *terraform.State) error {
	if len(s.RootModule().Resources) > 0 {
		return fmt.Errorf("Expected all resources to be gone, but found: %#v", s.RootModule().Resources)
	}

	return nil
}

var testAccAWSOpsWorksCustomRecipesConfig = testAccAWSOpsWorksCustomLayerConfig + `
resource "aws_opsworks_custom_recipes" "tf-acc" {
  layer_id = "${aws_opsworks_custom_layer.tf-acc.id}"

  configure = [
    "tfacc::configure1",
    "tfacc::configure2",
  ]
  setup = [
    "tfacc::setup1",
    "tfacc::setup2",
  ]
  deploy = [
    "tfacc::deploy1",
    "tfacc::deploy2",
  ]
  undeploy = [
    "tfacc::undeploy1",
    "tfacc::undeploy2",
  ]
  shutdown = [
    "tfacc::shutdown1",
    "tfacc::shutdown2",
  ]
}
`

var testAccAWSOpsWorksCustomRecipesConfigUpdate = testAccAWSOpsWorksCustomLayerConfig + `
resource "aws_opsworks_custom_recipes" "tf-acc" {
  layer_id = "${aws_opsworks_custom_layer.tf-acc.id}"

  configure = [
    "tfacc::configure1",
    "tfacc::configure2",
    "tfacc::configure3",
  ]
  setup = [
    "tfacc::setup1",
    "tfacc::setup2",
    "tfacc::setup3",
  ]
  deploy = [
    "tfacc::deploy1",
    "tfacc::deploy2",
    "tfacc::deploy3",
  ]
  undeploy = [
    "tfacc::undeploy1",
    "tfacc::undeploy2",
    "tfacc::undeploy3",
  ]
  shutdown = [
    "tfacc::shutdown1",
    "tfacc::shutdown2",
    "tfacc::shutdown3",
  ]
}
`
