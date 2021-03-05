package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

func tableAwsSSMMaintenanceWindow(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_ssm_maintenance_window",
		Description: "AWS SSM Maintenance Window",
		Get: &plugin.GetConfig{
			KeyColumns:        plugin.SingleColumn("window_id"),
			ShouldIgnoreError: isNotFoundError([]string{"DoesNotExistException"}),
			Hydrate:           getAwsSSMMaintenanceWindow,
		},
		List: &plugin.ListConfig{
			Hydrate: listAwsSSMMaintenanceWindow,
		},
		GetMatrixItem: BuildRegionList,
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "The name of the  Maintenance Window.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "window_id",
				Description: "The ID of the Maintenance Window.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "enabled",
				Description: "Indicates whether the Maintenance Window is enabled.",
				Type:        proto.ColumnType_BOOL,
			},
			{
				Name:        "allow_unassociated_targets",
				Description: "Whether targets must be registered with the Maintenance Window before tasks can be defined for those targets.",
				Type:        proto.ColumnType_BOOL,
				Hydrate:     getAwsSSMMaintenanceWindow,
			},
			{
				Name:        "description",
				Description: "A description of the Maintenance Window.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "tags_src",
				Description: "A list of tags assigned to the Maintenance Window",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsSSMMaintenanceWindowTags,
				Transform:   transform.FromField("TagList"),
			},
			{
				Name:        "duration",
				Description: "The duration of the Maintenance Window in hours.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "cutoff",
				Description: "The number of hours before the end of the Maintenance Window that Systems Manager stops scheduling new tasks for execution.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "schedule",
				Description: "The schedule of the Maintenance Window in the form of a cron or rate expression.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "schedule_offset",
				Description: "The number of days to wait to run a Maintenance Window after the scheduled CRON expression date and time.",
				Type:        proto.ColumnType_INT,
			},
			{
				Name:        "targets",
				Description: "The targets of Maintenance Window.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getMaintenanceWindowTargets,
				Transform:   transform.FromField("Targets"),
			},
			{
				Name:        "tasks",
				Description: "The Tasks of Maintenance Window.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getMaintenanceWindowTasks,
				Transform:   transform.FromField("Tasks"),
			},
			{
				Name:        "modified_date",
				Description: "The date the Maintenance Window was last modified.",
				Type:        proto.ColumnType_TIMESTAMP,
				Hydrate:     getAwsSSMMaintenanceWindow,
			},
			{
				Name:        "next_execution_time",
				Description: "The next time the maintenance window will actually run, taking into account any specified times for the Maintenance Window to become active or inactive.",
				Type:        proto.ColumnType_DATETIME,
			},

			/// Standard columns for all tables
			{
				Name:        "title",
				Description: resourceInterfaceDescription("title"),
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: resourceInterfaceDescription("tags"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsSSMMaintenanceWindowTags,
				Transform:   transform.FromField("TagList").Transform(ssmMaintenanceWindowTagListToTurbotTags),
			},
			{
				Name:        "akas",
				Description: resourceInterfaceDescription("akas"),
				Type:        proto.ColumnType_JSON,
				Hydrate:     getAwsSSMMaintenanceWindowAkas,
				Transform:   transform.FromValue(),
			},
		}),
	}
}

//// HYDRATE FUNCTIONS

func getAwsSSMMaintenanceWindow(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getAwsSSMMaintenanceWindow")

	// TODO put me in helper function
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	var id string
	if h.Item != nil {
		i := h.Item.(*ssm.GetMaintenanceWindowOutput)
		id = *i.WindowId
	} else {
		id = d.KeyColumnQuals["window_id"].GetStringValue()
	}

	// Create Session
	svc, err := SsmService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// Build the params
	params := &ssm.GetMaintenanceWindowInput{
		WindowId: &id,
	}

	// Get call
	data, err := svc.GetMaintenanceWindow(params)
	if err != nil {
		logger.Debug("getAwsSSMMaintenanceWindow", "ERROR", err)
		return nil, err
	}

	return data, nil
}

//// LIST FUNCTION

func listAwsSSMMaintenanceWindow(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	// TODO put me in helper function
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	plugin.Logger(ctx).Trace("listAwsSSMMaintenanceWindow", "AWS_REGION", region)

	// Create session
	svc, err := SsmService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// List call
	err = svc.DescribeMaintenanceWindowsPages(
		&ssm.DescribeMaintenanceWindowsInput{},
		func(page *ssm.DescribeMaintenanceWindowsOutput, isLast bool) bool {
			for _, parameter := range page.WindowIdentities {
				d.StreamListItem(ctx, &ssm.GetMaintenanceWindowOutput{
					Name:              parameter.Name,
					Cutoff:            parameter.Cutoff,
					NextExecutionTime: parameter.NextExecutionTime,
					Schedule:          parameter.Schedule,
					ScheduleOffset:    parameter.ScheduleOffset,
					ScheduleTimezone:  parameter.ScheduleTimezone,
					WindowId:          parameter.WindowId,
					Description:       parameter.Description,
					Enabled:           parameter.Enabled,
					Duration:          parameter.Duration,
				})

			}
			return !isLast
		},
	)

	return nil, err
}

func getAwsSSMMaintenanceWindowAkas(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getAwsSSMMaintenanceWindowAkas")
	maintenanceWindowData := h.Item.(*ssm.GetMaintenanceWindowOutput)
	c, err := getCommonColumns(ctx, d, h)
	if err != nil {
		return nil, err
	}
	commonColumnData := c.(*awsCommonColumnData)
	aka := "arn:" + commonColumnData.Partition + ":ssm:" + commonColumnData.Region + ":" + commonColumnData.AccountId + ":maintenancewindow" + "/" + *maintenanceWindowData.WindowId

	return []string{aka}, nil
}

func getAwsSSMMaintenanceWindowTags(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getAwsSSMMaintenanceWindowTags")

	// TODO put me in helper function
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	maintenanceWindowData := h.Item.(*ssm.GetMaintenanceWindowOutput)

	// Create Session
	svc, err := SsmService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// Build the params
	params := &ssm.ListTagsForResourceInput{
		ResourceType: types.String("MaintenanceWindow"),
		ResourceId:   maintenanceWindowData.WindowId,
	}

	// Get call
	op, err := svc.ListTagsForResource(params)
	if err != nil {
		logger.Debug("getAwsSSMMaintenanceWindowTags", "ERROR", err)
		return nil, err
	}

	return op, nil
}

func getMaintenanceWindowTargets(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getMaintenanceWindowTargets")

	// TODO put me in helper function
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	maintenanceWindowData := h.Item.(*ssm.GetMaintenanceWindowOutput)

	// Create Session
	svc, err := SsmService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// Build the params
	params := &ssm.DescribeMaintenanceWindowTargetsInput{
		WindowId: maintenanceWindowData.WindowId,
	}

	// Get call
	op, err := svc.DescribeMaintenanceWindowTargets(params)
	if err != nil {
		logger.Debug("getMaintenanceWindowTargets", "ERROR", err)
		return nil, err
	}

	return op, nil
}

func getMaintenanceWindowTasks(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	logger := plugin.Logger(ctx)
	logger.Trace("getMaintenanceWindowTasks")

	// TODO put me in helper function
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	maintenanceWindowData := h.Item.(*ssm.GetMaintenanceWindowOutput)

	// Create Session
	svc, err := SsmService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// Build the params
	params := &ssm.DescribeMaintenanceWindowTasksInput{
		WindowId: maintenanceWindowData.WindowId,
	}

	// Get call
	op, err := svc.DescribeMaintenanceWindowTasks(params)
	if err != nil {
		logger.Debug("getMaintenanceWindowTasks", "ERROR", err)
		return nil, err
	}

	return op, nil
}

func ssmMaintenanceWindowTagListToTurbotTags(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	plugin.Logger(ctx).Trace("ssmMaintenanceWindowTagListToTurbotTags")
	tagList := d.Value.([]*ssm.Tag)

	// Mapping the resource tags inside turbotTags
	var turbotTagsMap map[string]string
	if tagList != nil {
		turbotTagsMap = map[string]string{}
		for _, i := range tagList {
			turbotTagsMap[*i.Key] = *i.Value
		}
	}

	return turbotTagsMap, nil
}
