package aws

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/iam"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSOpsWorksStack_NoVPC(t *testing.T) {
	var opsworksIAM testAccAWSOpsWorksIAM
	novpcConfig := &bytes.Buffer{}
	novpcConfigUpdate := &bytes.Buffer{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			// Ensure that necessary IAM roles are available
			resource.TestStep{
				Config: testAccOpsWorksConfig_pre, // noop
				Check: testAccCheckAWSOpsWorksEnsureIAM(t, &opsworksIAM, func() error {
					err := testAccAWSOpsWorksStack_NoVPCConfig.Execute(
						novpcConfig,
						&opsworksIAM,
					)
					err = testAccAWSOpsWorksStack_NoVPCConfigUpdate.Execute(
						novpcConfigUpdate,
						&opsworksIAM,
					)
					return err
				}),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: novpcConfig.String(),
				Check:  testAccAWSOpsWorksStackCheckResourceAttrs(&opsworksIAM),
			},
			resource.TestStep{
				Config: novpcConfigUpdate.String(),
				Check:  testAccAWSOpsWorksStackCheckResourceAttrsUpdate(&opsworksIAM),
			},
		},
	})
}

func TestAccAWSOpsWorksStack_VPC(t *testing.T) {
	var opsworksIAM testAccAWSOpsWorksIAM
	vpcConfig := &bytes.Buffer{}
	vpcConfigUpdate := &bytes.Buffer{}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			// Ensure that necessary IAM roles are available
			resource.TestStep{
				Config: testAccOpsWorksConfig_pre, // noop
				Check: testAccCheckAWSOpsWorksEnsureIAM(t, &opsworksIAM, func() error {
					err := testAccAWSOpsWorksStack_VPCConfig.Execute(
						vpcConfig,
						&opsworksIAM,
					)
					err = testAccAWSOpsWorksStack_VPCConfigUpdate.Execute(
						vpcConfigUpdate,
						&opsworksIAM,
					)
					return err
				}),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: vpcConfig.String(),
				Check:  testAccAWSOpsWorksStackCheckResourceAttrs(&opsworksIAM),
			},

			resource.TestStep{
				Config: vpcConfigUpdate.String(),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSOpsWorksStackCheckResourceAttrsUpdate(&opsworksIAM),
					testAccCheckAWSOpsWorksVPC,
				),
			},
		},
	})
}

func testAccCheckAWSOpsWorksStackDestroy(s *terraform.State) error {
	if len(s.RootModule().Resources) > 0 {
		return fmt.Errorf("Expected all resources to be gone, but found: %#v", s.RootModule().Resources)
	}

	return nil
}

func testAccCheckAWSOpsWorksVPC(s *terraform.State) error {
	rs, ok := s.RootModule().Resources["aws_opsworks_stack.tf-acc"]
	if !ok {
		return fmt.Errorf("Not found: %s", "aws_opsworks_stack.tf-acc")
	}
	if rs.Primary.ID == "" {
		return fmt.Errorf("No ID is set")
	}

	p := rs.Primary

	opsworksconn := testAccProvider.Meta().(*AWSClient).opsworksconn
	describeOpts := &opsworks.DescribeStacksInput{
		StackIDs: []*string{aws.String(p.ID)},
	}
	resp, err := opsworksconn.DescribeStacks(describeOpts)
	if err != nil {
		return err
	}
	if len(resp.Stacks) == 0 {
		return fmt.Errorf("No stack %s not found", p.ID)
	}
	if p.Attributes["vpc_id"] != *resp.Stacks[0].VPCID {
		return fmt.Errorf("VPCID Got %s, expected %s", *resp.Stacks[0].VPCID, p.Attributes["vpc_id"])
	}
	if p.Attributes["default_subnet_id"] != *resp.Stacks[0].DefaultSubnetID {
		return fmt.Errorf("VPCID Got %s, expected %s", *resp.Stacks[0].DefaultSubnetID, p.Attributes["default_subnet_id"])
	}
	return nil
}

func testAccCheckAWSOpsWorksEnsureIAM(t *testing.T, oiam *testAccAWSOpsWorksIAM, f func() error) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		iamconn := testAccProvider.Meta().(*AWSClient).iamconn

		serviceRoleOpts := &iam.GetRoleInput{
			RoleName: aws.String("aws-opsworks-service-role"),
		}
		respServiceRole, err := iamconn.GetRole(serviceRoleOpts)
		if err != nil {
			return err
		}

		instanceProfileOpts := &iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String("aws-opsworks-ec2-role"),
		}
		respInstanceProfile, err := iamconn.GetInstanceProfile(instanceProfileOpts)
		if err != nil {
			return err
		}

		*oiam = testAccAWSOpsWorksIAM{
			ServiceRoleARN:     *respServiceRole.Role.ARN,
			InstanceProfileARN: *respInstanceProfile.InstanceProfile.ARN,
		}

		t.Logf("[DEBUG] ServiceRoleARN for OpsWorks: %s", oiam.ServiceRoleARN)
		t.Logf("[DEBUG] Instance Profile ARN for OpsWorks: %s", oiam.InstanceProfileARN)

		return f()
	}
}

var testAccAWSOpsWorksStackCheckResourceAttrs = func(oiam *testAccAWSOpsWorksIAM) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "name", "tf-opsworks-acc"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "service_role_arn", oiam.ServiceRoleARN),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_instance_profile_arn", oiam.InstanceProfileARN),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_availability_zone", "us-west-2a"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_os", "Amazon Linux 2014.09"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_root_device_type", "ebs"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "custom_json", `{"key": "value"}`),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "chef_version", "11.10"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "use_opsworks_security_groups", "true"),
	)
}

var testAccAWSOpsWorksStackCheckResourceAttrsUpdate = func(oiam *testAccAWSOpsWorksIAM) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "name", "tf-opsworks-acc"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "service_role_arn", oiam.ServiceRoleARN),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_instance_profile_arn", oiam.InstanceProfileARN),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_availability_zone", "us-west-2a"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_os", "Amazon Linux 2014.09"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "default_root_device_type", "ebs"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "custom_json", `{"key": "value"}`),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "chef_version", "11.10"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "use_opsworks_security_groups", "true"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "use_custom_cookbooks", "true"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "manage_berkshelf", "true"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "cookbook_source.3517999628.type", "git"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "cookbook_source.3517999628.revision", "master"),
		resource.TestCheckResourceAttr(
			"aws_opsworks_stack.tf-acc", "cookbook_source.3517999628.url", "https://github.com/awslabs/opsworks-example-cookbooks.git"),
	)
}

// testAccAWSOpsWorksIAM is a IAM struct for acceptance test of OpsWorks
type testAccAWSOpsWorksIAM struct {
	ServiceRoleARN     string
	InstanceProfileARN string
}

const testAccOpsWorksConfig_pre = `
resource "aws_security_group" "tf_test_foo" {
	name = "tf_test_foo"
	description = "foo"

	ingress {
		protocol = "icmp"
		from_port = -1
		to_port = -1
		cidr_blocks = ["0.0.0.0/0"]
	}
}
`

var testAccAWSOpsWorksStack_NoVPCConfig = template.Must(
	template.New("aws_opsworks_stack_novpc_config").Parse(`
resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  service_role_arn = "{{.ServiceRoleARN}}"
  default_instance_profile_arn = "{{.InstanceProfileARN}}"
  default_availability_zone = "us-west-2a"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
}
`))

var testAccAWSOpsWorksStack_NoVPCConfigUpdate = template.Must(
	template.New("aws_opsworks_stack_novpc_config_update").Parse(`
resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  service_role_arn = "{{.ServiceRoleARN}}"
  default_instance_profile_arn = "{{.InstanceProfileARN}}"
  default_availability_zone = "us-west-2a"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
  use_custom_cookbooks = true
  manage_berkshelf = true
  cookbook_source {
    type = "git"
    revision = "master"
    url = "https://github.com/awslabs/opsworks-example-cookbooks.git"
  }
}
`))

var testAccAWSOpsWorksStack_VPCConfig = template.Must(
	template.New("aws_opsworks_stack_vpc_config").Parse(`
resource "aws_vpc" "tf-acc" {
  cidr_block = "10.3.5.0/24"
}

resource "aws_subnet" "tf-acc" {
  vpc_id = "${aws_vpc.tf-acc.id}"
  cidr_block = "${aws_vpc.tf-acc.cidr_block}"
  availability_zone = "us-west-2a"
}

resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  vpc_id = "${aws_vpc.tf-acc.id}"
  default_subnet_id = "${aws_subnet.tf-acc.id}"
  service_role_arn = "{{.ServiceRoleARN}}"
  default_instance_profile_arn = "{{.InstanceProfileARN}}"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
}
`))

var testAccAWSOpsWorksStack_VPCConfigUpdate = template.Must(
	template.New("aws_opsworks_stack_vpc_config_update").Parse(`
resource "aws_vpc" "tf-acc" {
  cidr_block = "10.3.5.0/24"
}

resource "aws_subnet" "tf-acc" {
  vpc_id = "${aws_vpc.tf-acc.id}"
  cidr_block = "${aws_vpc.tf-acc.cidr_block}"
  availability_zone = "us-west-2a"
}

resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  vpc_id = "${aws_vpc.tf-acc.id}"
  default_subnet_id = "${aws_subnet.tf-acc.id}"
  service_role_arn = "{{.ServiceRoleARN}}"
  default_instance_profile_arn = "{{.InstanceProfileARN}}"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
  use_custom_cookbooks = true
  manage_berkshelf = true
  cookbook_source {
    type = "git"
    revision = "master"
    url = "https://github.com/awslabs/opsworks-example-cookbooks.git"
  }
}
`))
