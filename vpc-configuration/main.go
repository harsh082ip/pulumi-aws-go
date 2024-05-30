package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

/*
	Create a VPC.
	Create an Internet gateway and attach this gateway to the VPC.
	Create two subnets named "private" and "public."
	Create two route tables. Name one "public," and give it access to the Internet by providing the
	same Internet gateway that we created. Assign this route table to the subnet named "public."
	Name the second route table "private," and do not give it access to the Internet. Assign this
	table to the subnet named "private."
	Finally, create an EC2 instance in the public subnet and install nginx on it.
*/

func main() {
	pulumi.Run(SetupVpcAndStartTheServer)
}

func SetupVpcAndStartTheServer(ctx *pulumi.Context) error {
	// ---------------------- VPC --------------------------------
	// create vpc-args
	vpcArgs := &ec2.VpcArgs{
		CidrBlock:          pulumi.String("12.0.0.0/16"),
		EnableDnsSupport:   pulumi.Bool(true),
		EnableDnsHostnames: pulumi.Bool(true),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("test-infra"),
		},
		InstanceTenancy: ec2.TenancyDefault,
	}

	// Create vpc
	vpc, err := ec2.NewVpc(ctx, "test-infra", vpcArgs)
	if err != nil {
		return err
	}

	// --------------------------------- INTERNET GATEWAY ----------------------------------------
	// Create an Internet-gateway
	igw, err := ec2.NewInternetGateway(ctx, "test-infra-igw", &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("test-infra-igw"),
		},
	})
	if err != nil {
		return err
	}

	// -------------------------------- PUBLIC SUBNET -----------------------------------------------
	// public subnet args
	publicSubnetArgs := &ec2.SubnetArgs{
		VpcId:               vpc.ID(),
		CidrBlock:           pulumi.String("12.0.1.0/24"),
		AvailabilityZone:    pulumi.String("ap-south-1a"),
		MapPublicIpOnLaunch: pulumi.Bool(true), // otherwise we wont get a public ip on ec2 launch
		Tags: pulumi.StringMap{
			"Name": pulumi.String("test-infra-publicSubnet"),
		},
	}

	//Create public subnet
	publicSubnet, err := ec2.NewSubnet(ctx, "test-infra-publicSubnet", publicSubnetArgs)
	if err != nil {
		return err
	}

	// --------------------------------- PRIVATE SUBNET ----------------------------------------------
	// private subnet args
	privateSubnetArgs := &ec2.SubnetArgs{
		VpcId:            vpc.ID(),
		CidrBlock:        pulumi.String("12.0.2.0/24"),
		AvailabilityZone: pulumi.String("ap-south-1a"),
		Tags: pulumi.StringMap{
			"Name": pulumi.String("test-infra-privateSubnet"),
		},
	}

	// create private subnet
	privateSubnet, err := ec2.NewSubnet(ctx, "test-infra-privateSubnet", privateSubnetArgs)
	if err != nil {
		return err
	}

	// ----------------------------------- PUBLIC ROUTE TABLE ---------------------------------------------
	// public route table args
	publicRouteTableArgs := &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		Routes: ec2.RouteTableRouteArray{
			&ec2.RouteTableRouteArgs{
				CidrBlock: pulumi.String("0.0.0.0/0"),
				GatewayId: igw.ID(),
			},
		},
		Tags: pulumi.StringMap{
			"Name": pulumi.String("infra-public-route-table"),
		},
	}

	// create public route table
	publicRouteTable, err := ec2.NewRouteTable(ctx, "infra-public-route-table", publicRouteTableArgs)
	if err != nil {
		return err
	}

	// Associate the public route table with the public subnet
	_, err = ec2.NewRouteTableAssociation(ctx, "publicRouteTableAssociation", &ec2.RouteTableAssociationArgs{
		SubnetId:     publicSubnet.ID(),
		RouteTableId: publicRouteTable.ID(),
	})
	if err != nil {
		return err
	}

	// ---------------------------------- PRIVATE ROUTE TABLE ------------------------------------------------
	// private route table args
	privateRouteTableArgs := &ec2.RouteTableArgs{
		VpcId: vpc.ID(),
		// Not providing the igw
		Tags: pulumi.StringMap{
			"Name": pulumi.String("infra-private-route-table"),
		},
	}

	// create private route table
	privateRouteTable, err := ec2.NewRouteTable(ctx, "infra-private-route-table", privateRouteTableArgs)
	if err != nil {
		return err
	}

	// Associate the private route table with the private subnet
	_, err = ec2.NewRouteTableAssociation(ctx, "privateSubnetAssociation", &ec2.RouteTableAssociationArgs{
		SubnetId:     privateSubnet.ID(),
		RouteTableId: privateRouteTable.ID(),
	})
	if err != nil {
		return err
	}

	// --------------------------------- EC2 CONFIGURATION -----------------------------
	p_sid := publicSubnet.ID()
	vpc_id := vpc.ID()
	nginxServer, err := CreateEc2ServerWithNginx(ctx, p_sid, vpc_id)
	if err != nil {
		return err
	}

	// --------------------------- FINAL EXPORTS ------------------------------------------------
	// Export the IDs of the VPC, subnets, and EC2 instance
	ctx.Export("vpcId", vpc.ID())
	ctx.Export("publicSubnetId", publicSubnet.ID())
	ctx.Export("privateSubnetId", privateSubnet.ID())
	ctx.Export("publicIP", nginxServer.PublicIp)
	ctx.Export("publicDNS", nginxServer.PublicDns)

	return nil
}
