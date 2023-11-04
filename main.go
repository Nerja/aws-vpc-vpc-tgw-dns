package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2transitgateway"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcA, err := ec2.NewVpc(ctx, "vpc-a", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.0.0.0/16"),
			EnableDnsHostnames: pulumi.BoolPtr(true),
			EnableDnsSupport:   pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		vpcASubnet, err := ec2.NewSubnet(ctx, "vpc-a-subnet1", &ec2.SubnetArgs{
			CidrBlock:        pulumi.String("10.0.0.0/24"),
			VpcId:            vpcA.ID(),
			AvailabilityZone: pulumi.String("eu-north-1a"),
		})

		vpcASubnet2, err := ec2.NewSubnet(ctx, "vpc-a-subnet2", &ec2.SubnetArgs{
			CidrBlock:        pulumi.String("10.0.1.0/24"),
			VpcId:            vpcA.ID(),
			AvailabilityZone: pulumi.String("eu-north-1b"),
		})

		vpcB, err := ec2.NewVpc(ctx, "vpc-b", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.1.0.0/16"),
			EnableDnsHostnames: pulumi.BoolPtr(true),
			EnableDnsSupport:   pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		vpcBSubnet, err := ec2.NewSubnet(ctx, "vpc-b-subnet1", &ec2.SubnetArgs{
			CidrBlock:        pulumi.String("10.1.0.0/24"),
			VpcId:            vpcB.ID(),
			AvailabilityZone: pulumi.String("eu-north-1a"),
		})

		vpcBSubnet2, err := ec2.NewSubnet(ctx, "vpc-b-subnet2", &ec2.SubnetArgs{
			CidrBlock:        pulumi.String("10.1.1.0/24"),
			VpcId:            vpcB.ID(),
			AvailabilityZone: pulumi.String("eu-north-1b"),
		})

		tgw, err := ec2transitgateway.NewTransitGateway(ctx, "tgw", &ec2transitgateway.TransitGatewayArgs{
			DnsSupport:                   pulumi.String("enable"),
			DefaultRouteTableAssociation: pulumi.String("enable"),
			DefaultRouteTablePropagation: pulumi.String("enable"),
			TransitGatewayCidrBlocks:     pulumi.StringArray{pulumi.String("10.2.0.0/24")},
		})
		if err != nil {
			return err
		}

		_, err = ec2transitgateway.NewVpcAttachment(ctx, "attach-vpca", &ec2transitgateway.VpcAttachmentArgs{
			TransitGatewayId: tgw.ID(),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
				vpcASubnet2.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2transitgateway.NewVpcAttachment(ctx, "attach-vpcb", &ec2transitgateway.VpcAttachmentArgs{
			TransitGatewayId: tgw.ID(),
			SubnetIds: pulumi.StringArray{
				vpcBSubnet.ID(),
				vpcBSubnet2.ID(),
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		// Create route tables and attach
		rtA, err := ec2.NewRouteTable(ctx, "vpc-a-rt", &ec2.RouteTableArgs{
			VpcId: vpcA.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: tgw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-a-rt-attach", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtA.ID(),
			SubnetId:     vpcASubnet.ID(),
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-a-rt-attach2", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtA.ID(),
			SubnetId:     vpcASubnet2.ID(),
		})
		if err != nil {
			return err
		}
		rtb, err := ec2.NewRouteTable(ctx, "vpc-b-rt", &ec2.RouteTableArgs{
			VpcId: vpcB.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: tgw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-b-rt-attach", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtb.ID(),
			SubnetId:     vpcBSubnet.ID(),
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-b-rt-attach2", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtb.ID(),
			SubnetId:     vpcBSubnet2.ID(),
		})
		if err != nil {
			return err
		}

		vpcASG, err := ec2.NewSecurityGroup(ctx, "vpcA-sg", &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("traffic from VPCs"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					Protocol:    pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpoint, err := ec2.NewVpcEndpoint(ctx, "vpca-ssm", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssm"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpointEC2Messages, err := ec2.NewVpcEndpoint(ctx, "vpca-ec2messages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ec2messages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpointSSMMessages, err := ec2.NewVpcEndpoint(ctx, "vpca-ssmmessages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssmmessages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		zone, err := route53.NewZone(ctx, "vpcb-zone", &route53.ZoneArgs{
			Name: pulumi.String("vpcb.internal"),
			Vpcs: route53.ZoneVpcArray{
				route53.ZoneVpcArgs{
					VpcId:     vpcB.ID(),
					VpcRegion: pulumi.String("eu-north-1"),
				},
			},
		})
		if err != nil {
			return err
		}

		_, err = route53.NewRecord(ctx, "vpcb-zone-api-record", &route53.RecordArgs{
			Name:   pulumi.String("api.vpcb.internal"),
			ZoneId: zone.ZoneId,
			Type:   pulumi.String("A"),
			Ttl:    pulumi.Int(5),
			Records: pulumi.StringArray{
				pulumi.String("1.1.1.1"),
			},
		})
		if err != nil {
			return err
		}

		vpcDNSInboundSG, err := ec2.NewSecurityGroup(ctx, "vpcb-resolver-inbound-sg", &ec2.SecurityGroupArgs{
			VpcId: vpcB.ID(),
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("DNS inbound from VPC A"),
					FromPort:    pulumi.Int(53),
					ToPort:      pulumi.Int(53),
					Protocol:    pulumi.String("udp"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
					},
				},
			},
		})
		if err != nil {
			return err
		}

		inboundEndpoint, err := route53.NewResolverEndpoint(ctx, "vpcb-resolver-inbound", &route53.ResolverEndpointArgs{
			Direction: pulumi.String("INBOUND"),
			IpAddresses: route53.ResolverEndpointIpAddressArray{
				route53.ResolverEndpointIpAddressArgs{
					SubnetId: vpcBSubnet.ID(),
				},
				route53.ResolverEndpointIpAddressArgs{
					SubnetId: vpcBSubnet2.ID(),
				},
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcDNSInboundSG.ID(),
			},
		})
		if err != nil {
			return err
		}

		vpcDNSOutboundSG, err := ec2.NewSecurityGroup(ctx, "vpca-resolver-outbound-sg", &ec2.SecurityGroupArgs{
			VpcId: vpcA.ID(),
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					Description: pulumi.String("DNS outbound to VPC B"),
					FromPort:    pulumi.Int(53),
					ToPort:      pulumi.Int(53),
					Protocol:    pulumi.String("udp"),
					CidrBlocks: pulumi.StringArray{
						vpcB.CidrBlock,
					},
				},
			},
		})
		if err != nil {
			return err
		}

		outboundEndpoint, err := route53.NewResolverEndpoint(ctx, "vpca-resolver-outbound", &route53.ResolverEndpointArgs{
			Direction: pulumi.String("OUTBOUND"),
			IpAddresses: route53.ResolverEndpointIpAddressArray{
				route53.ResolverEndpointIpAddressArgs{
					SubnetId: vpcASubnet.ID(),
				},
				route53.ResolverEndpointIpAddressArgs{
					SubnetId: vpcASubnet2.ID(),
				},
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcDNSOutboundSG.ID(),
			},
		})
		if err != nil {
			return err
		}

		ips := inboundEndpoint.IpAddresses.ApplyT(func(ips []route53.ResolverEndpointIpAddress) (route53.ResolverRuleTargetIpArrayInput, error) {
			return route53.ResolverRuleTargetIpArray{
				route53.ResolverRuleTargetIpArgs{
					Ip: pulumi.String(*ips[0].Ip),
				},
				route53.ResolverRuleTargetIpArgs{
					Ip: pulumi.String(*ips[1].Ip),
				},
			}, nil
		}).(route53.ResolverRuleTargetIpArrayInput)

		vpcAResolverRule, err := route53.NewResolverRule(ctx, "vpca-resolver-rule", &route53.ResolverRuleArgs{
			DomainName:         pulumi.String("api.vpcb.internal"),
			RuleType:           pulumi.String("FORWARD"),
			ResolverEndpointId: outboundEndpoint.ID(),
			TargetIps:          ips,
		})
		if err != nil {
			return err
		}

		_, err = route53.NewResolverRuleAssociation(ctx, "vpca-resolver-rule-association", &route53.ResolverRuleAssociationArgs{
			ResolverRuleId: vpcAResolverRule.ID(),
			VpcId:          vpcA.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewInstance(ctx, "vpcA-EC2", &ec2.InstanceArgs{
			Ami:      pulumi.String("ami-0b5483e9d9802be1f"),
			SubnetId: vpcASubnet.ID(),
			VpcSecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			InstanceType:       pulumi.String("t4g.nano"),
			IamInstanceProfile: pulumi.String("ec2-ssm-mgmt"),
		}, pulumi.DependsOn([]pulumi.Resource{vpcAEndpoint, vpcAEndpointEC2Messages, vpcAEndpointSSMMessages}))

		ctx.Export("vpcA", vpcA.ID())
		ctx.Export("vpcB", vpcB.ID())
		return nil
	})

}
