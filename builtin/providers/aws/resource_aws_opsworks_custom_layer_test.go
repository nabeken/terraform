package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// To run tests, we need predefined IAM roles such as `aws-opsworks-ec2-role` and `aws-opsworks-service-role`.

func TestAccAWSOpsWorksCustomLayer(t *testing.T) {
	opsiam := testAccAWSOpsWorksIAM{}

	testAccAWSOpsWorksPopulateIAM(t, &opsiam)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksCustomLayerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccAWSOpsWorksCustomLayerConfig, opsiam.ServiceRoleARN, opsiam.InstanceProfileARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "name", "tf-ops-acc-custom-layer"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "enable_auto_healing", "true"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "shutdown_event_configuration.644301331.delay_until_elb_connections_drained", "true"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "shutdown_event_configuration.644301331.execution_timeout", "300"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.1368285564", "git"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.2937857443", "golang"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.type", "gp2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.num_disks", "1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.mount_point", "/home"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.size", "100"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccAWSOpsWorksCustomLayerConfigUpdate, opsiam.ServiceRoleARN, opsiam.InstanceProfileARN),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "name", "tf-ops-acc-custom-layer"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "shutdown_event_configuration.3123070724.delay_until_elb_connections_drained", "false"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "shutdown_event_configuration.3123070724.execution_timeout", "120"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "custom_security_group_ids.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.1368285564", "git"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.2937857443", "golang"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "packages.4101929740", "subversion"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.type", "gp2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.num_disks", "1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.mount_point", "/home"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.3723647151.size", "100"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.type", "io1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.num_disks", "2"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.mount_point", "/var"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.size", "100"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.raid_level", "1"),
					resource.TestCheckResourceAttr(
						"aws_opsworks_custom_layer.tf-acc", "volume_configuration.1230927504.iops", "3000"),
				),
			},
		},
	})
}

func testAccCheckAWSOpsWorksCustomLayerDestroy(s *terraform.State) error {
	if len(s.RootModule().Resources) > 0 {
		return fmt.Errorf("Expected all resources to be gone, but found: %#v", s.RootModule().Resources)
	}

	return nil
}

var testAccAWSOpsWorksCustomLayerSG = `
resource "aws_security_group" "tf-ops-acc-layer1" {
  name = "tf-ops-acc-layer1"
  ingress {
    from_port = 8
    to_port = -1
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-layer2" {
  name = "tf-ops-acc-layer2"
  ingress {
    from_port = 8
    to_port = -1
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`

var testAccAWSOpsWorksCustomLayerConfig = testAccAWSOpsWorksStack_NoVPCConfig + testAccAWSOpsWorksCustomLayerSG + `
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id = "${aws_opsworks_stack.tf-acc.id}"
  name = "tf-ops-acc-custom-layer"
  short_name = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
  ]

  shutdown_event_configuration {
    delay_until_elb_connections_drained = true
    execution_timeout = 300
  }

  packages = [
    "git",
    "golang",
  ]

  volume_configuration {
    type = "gp2"
    num_disks = 1
    mount_point = "/home"
    size = 100
  }
}
`

var testAccAWSOpsWorksCustomLayerConfigUpdate = testAccAWSOpsWorksStack_NoVPCConfig + testAccAWSOpsWorksCustomLayerSG + `
resource "aws_security_group" "tf-ops-acc-layer3" {
  name = "tf-ops-acc-layer3"
  ingress {
    from_port = 8
    to_port = -1
    protocol = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id = "${aws_opsworks_stack.tf-acc.id}"
  name = "tf-ops-acc-custom-layer"
  short_name = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    "${aws_security_group.tf-ops-acc-layer1.id}",
    "${aws_security_group.tf-ops-acc-layer2.id}",
    "${aws_security_group.tf-ops-acc-layer3.id}",
  ]

  shutdown_event_configuration {
    delay_until_elb_connections_drained = false
    execution_timeout = 120
  }

  packages = [
    "git",
    "golang",
    "subversion",
  ]

  volume_configuration {
    type = "gp2"
    num_disks = 1
    mount_point = "/home"
    size = 100
  }

  volume_configuration {
    type = "io1"
    num_disks = 2
    mount_point = "/var"
    size = 100
    raid_level = "1"
    iops = 3000
  }
}
`
