package aws

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

//// TABLE DEFINITION

func tableAwsIamUser(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_iam_user",
		Description: "AWS IAM User",
		Get: &plugin.GetConfig{
			KeyColumns:  plugin.SingleColumn("name"),
			ItemFromKey: userFromKey,
			Hydrate:     getIamUser,
		},
		List: &plugin.ListConfig{
			Hydrate: listIamUsers,
		},
		HydrateDependencies: []plugin.HydrateDependencies{
			{
				Func:    getAwsIamUserInlinePolicies,
				Depends: []plugin.HydrateFunc{listAwsIamUserInlinePolicies},
			},
		},
		Columns: awsColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "The friendly name identifying the user",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
			},
			{
				Name:        "user_id",
				Description: "The stable and unique string identifying the user",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "path",
				Description: "The path to the user",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "arn",
				Description: "The Amazon Resource Name (ARN) that identifies the user",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "create_date",
				Description: "The date and time, when the user was created",
				Type:        proto.ColumnType_DATETIME,
			},
			{
				Name:        "password_last_used",
				Description: "The date and time, when the user's password was last used to sign in to an AWS website",
				Type:        proto.ColumnType_DATETIME,
			},
			{
				Name:        "permissions_boundary_arn",
				Description: "The ARN of the policy used to set the permissions boundary for the user",
				Type:        proto.ColumnType_STRING,
				Hydrate:     getAwsIamUserData,
			},
			{
				Name: "permissions_boundary_type",
				Description: "The permissions boundary usage type that indicates what type of IAM resource " +
					"is used as the permissions boundary for an entity. This data type can only have a value of Policy",
				Type:    proto.ColumnType_STRING,
				Hydrate: getAwsIamUserData,
			},
			{
				Name:        "groups",
				Description: "A list of groups attached to the user",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsIamUserGroups,
			},
			{
				Name:        "inline_policies",
				Description: "A list of policy documents that are embedded as inline policies for the user",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsIamUserInlinePolicies,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "attached_policy_arns",
				Description: "A list of managed policies attached to the user",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsIamUserAttachedPolicies,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "tags_src",
				Description: "A list of tags that are attached to the user",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsIamUserData,
			},

			// Standard columns for all tables
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsIamUserData,
			},
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("UserName"),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("Arn").Transform(arnToAkas),
			},
		}),
	}
}

//// BUILD HYDRATE INPUT

func userFromKey(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	quals := d.KeyColumnQuals
	name := quals["name"].GetStringValue()
	item := &iam.User{
		UserName: &name,
	}
	return item, nil
}

//// LIST FUNCTION

func listIamUsers(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	err = svc.ListUsersPages(
		&iam.ListUsersInput{},
		func(page *iam.ListUsersOutput, lastPage bool) bool {
			for _, user := range page.Users {
				d.StreamListItem(ctx, user)
			}
			return true
		},
	)

	return nil, err
}

//// HYDRATE FUNCTIONS

func getIamUser(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getIamUser")
	user := h.Item.(*iam.User)

	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	params := &iam.GetUserInput{
		UserName: user.UserName,
	}

	op, err := svc.GetUser(params)
	if err != nil {
		return nil, err
	}

	return op.User, nil
}

func getAwsIamUserData(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsIamUserData")
	user := h.Item.(*iam.User)

	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	params := &iam.GetUserInput{
		UserName: user.UserName,
	}

	userData, _ := svc.GetUser(params)
	if err != nil {
		return nil, err
	}

	var tags []*iam.Tag
	var turbotTags map[string]string
	PermissionsBoundaryArn := ""
	PermissionsBoundaryType := ""

	if userData.User.Tags != nil {
		tags = userData.User.Tags
		turbotTags = map[string]string{}
		for _, t := range userData.User.Tags {
			turbotTags[*t.Key] = *t.Value
		}
	}

	if userData.User.PermissionsBoundary != nil && userData.User.PermissionsBoundary.PermissionsBoundaryArn != nil {
		v := userData.User.PermissionsBoundary
		PermissionsBoundaryArn = *v.PermissionsBoundaryArn
		PermissionsBoundaryType = *v.PermissionsBoundaryType
	}

	return map[string]interface{}{
		"TagsRaw":                 tags,
		"Tags":                    turbotTags,
		"PermissionsBoundaryArn":  PermissionsBoundaryArn,
		"PermissionsBoundaryType": PermissionsBoundaryType,
	}, nil
}

func getAwsIamUserAttachedPolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsIamUserAttachedPolicies")
	user := h.Item.(*iam.User)

	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	params := &iam.ListAttachedUserPoliciesInput{
		UserName: user.UserName,
	}

	userData, _ := svc.ListAttachedUserPolicies(params)
	if err != nil {
		return nil, err
	}

	var attachedPolicyArns []string

	if userData.AttachedPolicies != nil {
		for _, policy := range userData.AttachedPolicies {
			attachedPolicyArns = append(attachedPolicyArns, *policy.PolicyArn)
		}
	}

	return attachedPolicyArns, nil
}

func getAwsIamUserGroups(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsIamUserGroups")
	user := h.Item.(*iam.User)

	// create service
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	params := &iam.ListGroupsForUserInput{
		UserName: user.UserName,
	}

	userData, _ := svc.ListGroupsForUser(params)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func listAwsIamUserInlinePolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("listAwsIamUserInlinePolicies")
	user := h.Item.(*iam.User)

	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	params := &iam.ListUserPoliciesInput{
		UserName: user.UserName,
	}

	userData, err := svc.ListUserPolicies(params)
	if err != nil {
		return nil, err
	}

	return userData, nil
}

func getAwsIamUserInlinePolicies(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsIamUserInlinePolicies")
	user := h.Item.(*iam.User)
	listUserPoliciesOutput := h.HydrateResults["listAwsIamUserInlinePolicies"].(*iam.ListUserPoliciesOutput)

	// Create Session
	svc, err := IAMService(ctx, d.ConnectionManager)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	policyCh := make(chan map[string]interface{}, len(listUserPoliciesOutput.PolicyNames))
	errorCh := make(chan error, len(listUserPoliciesOutput.PolicyNames))
	for _, policy := range listUserPoliciesOutput.PolicyNames {
		wg.Add(1)
		go getUserPolicyDataAsync(policy, user.UserName, svc, &wg, policyCh, errorCh)
	}

	// wait for all inline policies to be processed
	wg.Wait()
	// NOTE: close channel before ranging over results
	close(policyCh)
	close(errorCh)

	for err := range errorCh {
		// return the first error
		return nil, err
	}

	var userPolicies []map[string]interface{}

	for userPolicy := range policyCh {
		userPolicies = append(userPolicies, userPolicy)
	}

	return userPolicies, nil
}

func getUserPolicyDataAsync(policy *string, userName *string, svc *iam.IAM, wg *sync.WaitGroup, policyCh chan map[string]interface{}, errorCh chan error) {
	defer wg.Done()

	rowData, err := getUserInlinePolicy(policy, userName, svc)
	if err != nil {
		errorCh <- err
	} else if rowData != nil {
		policyCh <- rowData
	}
}

func getUserInlinePolicy(policyName *string, userName *string, svc *iam.IAM) (map[string]interface{}, error) {
	userPolicy := make(map[string]interface{})
	params := &iam.GetUserPolicyInput{
		PolicyName: policyName,
		UserName:   userName,
	}

	tmpPolicy, err := svc.GetUserPolicy(params)
	if err != nil {
		return nil, err
	}

	if tmpPolicy != nil && tmpPolicy.PolicyDocument != nil {
		decoded, decodeErr := url.QueryUnescape(*tmpPolicy.PolicyDocument)
		if decodeErr != nil {
			return nil, decodeErr
		}

		var rawPolicy interface{}
		err := json.Unmarshal([]byte(decoded), &rawPolicy)
		if err != nil {
			return nil, err
		}

		userPolicy = map[string]interface{}{
			"PolicyDocument": rawPolicy,
			"PolicyName":     *tmpPolicy.PolicyName,
		}
	}

	return userPolicy, nil
}
