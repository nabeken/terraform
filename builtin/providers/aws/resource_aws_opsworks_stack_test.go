package aws

import (
	"fmt"
	"testing"

	"github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSOpsWorksStack_NoVPC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSOpsWorksStack_NoVPCConfig,
				Check:  testAccAWSOpsWorksStackCheckResourceAttrs,
			},
			resource.TestStep{
				Config: testAccAWSOpsWorksStack_VPCConfigUpdate,
				Check:  testAccAWSOpsWorksStackCheckResourceAttrsUpdate,
			},
		},
	})
}

func TestAccAWSOpsWorksStack_VPC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSOpsWorksStackDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAWSOpsWorksStack_VPCConfig,
				Check:  testAccAWSOpsWorksStackCheckResourceAttrs,
			},

			resource.TestStep{
				Config: testAccAWSOpsWorksStack_VPCConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccAWSOpsWorksStackCheckResourceAttrsUpdate,
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

var testAccAWSOpsWorksStackCheckResourceAttrs = resource.ComposeTestCheckFunc(
	resource.TestCheckResourceAttr(
		"aws_opsworks_stack.tf-acc", "name", "tf-opsworks-acc"),
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

var testAccAWSOpsWorksStackCheckResourceAttrsUpdate = resource.ComposeTestCheckFunc(
	resource.TestCheckResourceAttr(
		"aws_opsworks_stack.tf-acc", "name", "tf-opsworks-acc"),
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

var testAccAWSOpsWorksStack_NoVPCConfig = `
resource "aws_iam_role" "tf-acc-opsworks-ec2-role" {
  name = "tf-acc-opsworks-ec2-role"
  assume_role_policy = "{\"Version\":\"2008-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_instance_profile" "tf-acc-opsworks-ec2-profile" {
  name = "tf-acc-opsworks-ec2-profile"
  roles = ["${aws_iam_role.tf-acc-opsworks-ec2-role.name}"]
}

resource "aws_iam_role" "tf-acc-opsworks-service-role" {
  name = "tf-acc-opsworks-service-role"
  assume_role_policy = "{\"Version\":\"2008-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"opsworks.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_role_policy" "tf-acc-opsworks-service-policy" {
  name = "tf-acc-opsworks-service-policy"
  role = "${aws_iam_role.tf-acc-opsworks-service-role.name}"
  policy = "{\"Statement\": [{\"Action\": [\"ec2:*\", \"iam:PassRole\",\"cloudwatch:GetMetricStatistics\",\"elasticloadbalancing:*\",\"rds:*\"],\"Effect\": \"Allow\",\"Resource\": [\"*\"] }]}"
}

resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  service_role_arn = "${aws_iam_role.tf-acc-opsworks-service-role.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.tf-acc-opsworks-ec2-profile.arn}"
  default_availability_zone = "us-west-2a"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
}
`

var testAccAWSOpsWorksStack_NoVPCConfigUpdate = `
resource "aws_iam_role" "tf-acc-opsworks-ec2-role" {
  name = "tf-acc-opsworks-ec2-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_instance_profile" "tf-acc-opsworks-ec2-profile" {
  name = "tf-acc-opsworks-ec2-profile"
  roles = ["${aws_iam_role.tf-acc-opsworks-ec2-role.name}"]
}

resource "aws_iam_role" "tf-acc-opsworks-service-role" {
  name = "tf-acc-opsworks-service-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"opsworks.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  service_role_arn = "${aws_iam_role.tf-acc-opsworks-service-role.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.tf-acc-opsworks-ec2-profile.arn}"
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
`

var testAccAWSOpsWorksStack_VPCConfig = `
resource "aws_vpc" "tf-acc" {
  cidr_block = "10.3.5.0/24"
}

resource "aws_subnet" "tf-acc" {
  vpc_id = "${aws_vpc.tf-acc.id}"
  cidr_block = "${aws_vpc.tf-acc.cidr_block}"
  availability_zone = "us-west-2a"
}

resource "aws_iam_role" "tf-acc-opsworks-vpc-ec2-role" {
  name = "tf-acc-opsworks-vpc-ec2-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_instance_profile" "tf-acc-opsworks-vpc-ec2-profile" {
  name = "tf-acc-opsworks-vpc-ec2-profile"
  roles = ["${aws_iam_role.tf-acc-opsworks-vpc-ec2-role.name}"]
}

resource "aws_iam_role" "tf-acc-opsworks-vpc-service-role" {
  name = "tf-acc-opsworks-vpc-service-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"opsworks.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_opsworks_stack" "tf-acc" {
  name = "tf-opsworks-acc"
  vpc_id = "${aws_vpc.tf-acc.id}"
  default_subnet_id = "${aws_subnet.tf-acc.id}"
  service_role_arn = "${aws_iam_role.tf-acc-opsworks-vpc-service-role.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.tf-acc-opsworks-vpc-ec2-profile.arn}"
  default_os = "Amazon Linux 2014.09"
  default_root_device_type = "ebs"
  custom_json = "{\"key\": \"value\"}"
  chef_version = "11.10"
  use_opsworks_security_groups = true
}
`

var testAccAWSOpsWorksStack_VPCConfigUpdate = `
resource "aws_vpc" "tf-acc" {
  cidr_block = "10.3.5.0/24"
}

resource "aws_iam_role" "tf-acc-opsworks-vpc-ec2-role" {
  name = "tf-acc-opsworks-vpc-ec2-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"ec2.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
}

resource "aws_iam_instance_profile" "tf-acc-opsworks-vpc-ec2-profile" {
  name = "tf-acc-opsworks-vpc-ec2-profile"
  roles = ["${aws_iam_role.tf-acc-opsworks-vpc-ec2-role.name}"]
}

resource "aws_iam_role" "tf-acc-opsworks-vpc-service-role" {
  name = "tf-acc-opsworks-vpc-service-role"
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Principal\":{\"Service\":\"opsworks.amazonaws.com\"},\"Effect\":\"Allow\",\"Sid\":\"\"}]}"
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
  service_role_arn = "${aws_iam_role.tf-acc-opsworks-vpc-service-role.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.tf-acc-opsworks-vpc-ec2-profile.arn}"
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
`
