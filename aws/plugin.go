/*
Package aws implements a steampipe plugin for aws.

This plugin provides data that Steampipe uses to present foreign
tables that represent Amazon AWS resources.
*/
package aws

import (
	"context"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

const pluginName = "steampipe-plugin-aws"

// Plugin creates this (aws) plugin
func Plugin(ctx context.Context) *plugin.Plugin {
	p := &plugin.Plugin{
		Name:             pluginName,
		DefaultTransform: transform.FromCamel(),
		DefaultGetConfig: &plugin.GetConfig{
			ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFoundException", "NoSuchEntity"}),
		},
		TableMap: map[string]*plugin.Table{
			"aws_account":                            tableAwsAccount(ctx),
			"aws_acm_certificate":                    tableAwsAcmCertificate(ctx),
			"aws_api_gateway_api_key":                tableAwsAPIGatewayAPIKey(ctx),
			"aws_api_gateway_authorizer":             tableAwsAPIGatewayAuthorizer(ctx),
			"aws_api_gateway_rest_api":               tableAwsAPIGatewayRestAPI(ctx),
			"aws_api_gateway_stage":                  tableAwsAPIGatewayStage(ctx),
			"aws_api_gateway_usage_plan":             tableAwsAPIGatewayUsagePlan(ctx),
			"aws_api_gatewayv2_api":                  tableAwsAPIGatewayV2Api(ctx),
			"aws_api_gatewayv2_domain_name":          tableAwsAPIGatewayV2DomainName(ctx),
			"aws_api_gatewayv2_stage":                tableAwsAPIGatewayV2Stage(ctx),
			"aws_availability_zone":                  tableAwsAvailabilityZone(ctx),
			"aws_cloudformation_stack":               tableAwsCloudFormationStack(ctx),
			"aws_cloudtrail_trail_event":							tableAwsCloudtrailEvent(ctx),
			"aws_cloudwatch_log_group":               tableAwsCloudwatchLogGroup(ctx),
			"aws_cloudwatch_log_metric_filter":       tableAwsCloudwatchLogMetricFilter(ctx),
			"aws_dynamodb_backup":                    tableAwsDynamoDBBackup(ctx),
			"aws_dynamodb_global_table":              tableAwsDynamoDBGlobalTable(ctx),
			"aws_dynamodb_table":                     tableAwsDynamoDBTable(ctx),
			"aws_ebs_snapshot":                       tableAwsEBSSnapshot(ctx),
			"aws_ebs_volume":                         tableAwsEBSVolume(ctx),
			"aws_ec2_ami":                            tableAwsEc2Ami(ctx),
			"aws_ec2_application_load_balancer":      tableAwsEc2ApplicationLoadBalancer(ctx),
			"aws_ec2_autoscaling_group":              tableAwsEc2ASG(ctx),
			"aws_ec2_classic_load_balancer":          tableAwsEc2ClassicLoadBalancer(ctx),
			"aws_ec2_instance":                       tableAwsEc2Instance(ctx),
			"aws_ec2_instance_availability":          tableAwsInstanceAvailability(ctx),
			"aws_ec2_instance_type":                  tableAwsInstanceType(ctx),
			"aws_ec2_key_pair":                       tableAwsEc2KeyPair(ctx),
			"aws_ec2_launch_configuration":           tableAwsEc2LaunchConfiguration(ctx),
			"aws_ec2_load_balancer_listener":         tableAwsEc2ApplicationLoadBalancerListener(ctx),
			"aws_ec2_network_interface":              tableAwsEc2NetworkInterface(ctx),
			"aws_ec2_network_load_balancer":          tableAwsEc2NetworkLoadBalancer(ctx),
			"aws_ec2_target_group":                   tableAwsEc2TargetGroup(ctx),
			"aws_ec2_transit_gateway":                tableAwsEc2TransitGateway(ctx),
			"aws_ec2_transit_gateway_route_table":    tableAwsEc2TransitGatewayRouteTable(ctx),
			"aws_ec2_transit_gateway_vpc_attachment": tableAwsEc2TransitGatewayVpcAttachment(ctx),
			"aws_iam_access_key":                     tableAwsIamAccessKey(ctx),
			"aws_iam_group":                          tableAwsIamGroup(ctx),
			"aws_iam_policy":                         tableAwsIamPolicy(ctx),
			"aws_iam_role":                           tableAwsIamRole(ctx),
			"aws_iam_user":                           tableAwsIamUser(ctx),
			"aws_kms_key":                            tableAwsKmsKey(ctx),
			"aws_lambda_alias":                       tableAwsLambdaAlias(ctx),
			"aws_lambda_function":                    tableAwsLambdaFunction(ctx),
			"aws_lambda_version":                     tableAwsLambdaVersion(ctx),
			"aws_rds_db_cluster":                     tableAwsRDSDBCluster(ctx),
			"aws_rds_db_cluster_parameter_group":     tableAwsRDSDBClusterParameterGroup(ctx),
			"aws_rds_db_cluster_snapshot":            tableAwsRDSDBClusterSnapshot(ctx),
			"aws_rds_db_instance":                    tableAwsRDSDBInstance(ctx),
			"aws_rds_db_option_group":                tableAwsRDSDBOptionGroup(ctx),
			"aws_rds_db_parameter_group":             tableAwsRDSDBParameterGroup(ctx),
			"aws_rds_db_snapshot":                    tableAwsRDSDBSnapshot(ctx),
			"aws_rds_db_subnet_group":                tableAwsRDSDBSubnetGroup(ctx),
			"aws_region":                             tableAwsRegion(ctx),
			"aws_s3_account_settings":                tableAwsS3AccountSettings(ctx),
			"aws_s3_bucket":                          tableAwsS3Bucket(ctx),
			"aws_sns_topic":                          tableAwsSnsTopic(ctx),
			"aws_sns_topic_subscription":             tableAwsSnsTopicSubscription(ctx),
			"aws_sqs_queue":                          tableAwsSqsQueue(ctx),
			"aws_ssm_parameter":                      tableAwsSSMParameter(ctx),
			"aws_vpc":                                tableAwsVpc(ctx),
			"aws_vpc_customer_gateway":               tableAwsVpcCustomerGateway(ctx),
			"aws_vpc_dhcp_options":                   tableAwsVpcDhcpOptions(ctx),
			"aws_vpc_egress_only_internet_gateway":   tableAwsVpcEgressOnlyIGW(ctx),
			"aws_vpc_eip":                            tableAwsVpcEip(ctx),
			"aws_vpc_endpoint":                       tableAwsVpcEndpoint(ctx),
			"aws_vpc_endpoint_service":               tableAwsVpcEndpointService(ctx),
			"aws_vpc_internet_gateway":               tableAwsVpcInternetGateway(ctx),
			"aws_vpc_nat_gateway":                    tableAwsVpcNatGateway(ctx),
			"aws_vpc_network_acl":                    tableAwsVpcNetworkACL(ctx),
			"aws_vpc_route":                          tableAwsVpcRoute(ctx),
			"aws_vpc_route_table":                    tableAwsVpcRouteTable(ctx),
			"aws_vpc_security_group":                 tableAwsVpcSecurityGroup(ctx),
			"aws_vpc_security_group_rule":            tableAwsVpcSecurityGroupRule(ctx),
			"aws_vpc_subnet":                         tableAwsVpcSubnet(ctx),
			"aws_vpc_vpn_gateway":                    tableAwsVpcVpnGateway(ctx),
		},
	}

	return p
}
