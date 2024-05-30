package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		// Security-group-conf
		sgArgs := &ec2.SecurityGroupArgs{
			//ingress
			Ingress: ec2.SecurityGroupIngressArray{
				ec2.SecurityGroupIngressArgs{
					Protocol:   pulumi.String("tcp"),
					FromPort:   pulumi.Int(80),
					ToPort:     pulumi.Int(80),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
			//egress
			Egress: ec2.SecurityGroupEgressArray{
				ec2.SecurityGroupEgressArgs{
					Protocol:   pulumi.String("-1"),
					FromPort:   pulumi.Int(0),
					ToPort:     pulumi.Int(0),
					CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
				},
			},
		}

		// create sg
		sg, err := ec2.NewSecurityGroup(ctx, "nginx-sg", sgArgs)
		if err != nil {
			return err
		}

		// key-pair
		kp, err := ec2.NewKeyPair(ctx, "nginx-kp", &ec2.KeyPairArgs{
			PublicKey: pulumi.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCTRXP2AfM1mNR0XSM7KhDTDpAwAbG4Hxlhfx4ahV86h4JI4Kn7nVOvscNGgfqaZr2lIio6y+A30IKASrIQttI9pm/4KIPzh2uGSuedO4HDWaiwWMkXLXeymQm+qYlWbGBvz0CZyoHeHpvbXl0TZAEaW6YSIiAzaicI9vUn+lOVjOst+/gyconivI93XggRHSQXkdr+Lx6LU4hgRTS0/FbD7bsZIjaSOSF63tLX6MD9nck0Amk2sPydOfBYVuOhaW9Lqn2Nb0VkL/q2eyYyycoWyexSvVW7idJ/7sRMFzU0eVrDO2ZkoE3wUa/EsLjen1Qyjm+5lB3knolUcSfllFdUOftuGsf/JU6YtzeZ6ptSiGPngSg9ix2gma3L0kgsPll0owoMgZy+nNgiaN7vrEm1Pf3Uc+3tVE0XZtVAhchYrsLdHgdb2c07dvm736sl9143Dhjt3DUzUluN1bao+UXnBLfvlJ173R0NmbV5VgVMV7u8qJ1WbY4IFOu9OFbJdkE= nightshade@cloud1"),
		})
		if err != nil {
			return err
		}

		// instance conf
		instanceArgs := &ec2.InstanceArgs{
			Ami:                 pulumi.String("ami-0f58b397bc5c1f2e8"),
			InstanceType:        pulumi.String("t2.micro"),
			VpcSecurityGroupIds: pulumi.StringArray{sg.ID()},
			KeyName:             kp.KeyName,
			UserData:            pulumi.String("#!/bin/bash\nsudo apt-get update\nsudo apt-get install -y nginx"),
			Tags: pulumi.StringMap{
				"Name": pulumi.String("nginx-server"),
			},
		}

		// create the instance, and start the server
		nginxServer, err := ec2.NewInstance(ctx, "nginx-server", instanceArgs)
		if err != nil {
			return err
		}

		// export details
		ctx.Export("publicIP", nginxServer.PublicIp)
		ctx.Export("publicDNS", nginxServer.PublicDns)

		return nil
	})
}
