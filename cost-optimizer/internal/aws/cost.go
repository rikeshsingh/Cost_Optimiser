package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// FetchCost returns total cost + service-wise breakdown
func FetchCost() (map[string]string, error) {

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := costexplorer.NewFromConfig(cfg)

	// Date range (last 30 days)
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")
	serviceKey := "SERVICE"

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startStr,
			End:   &endStr,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},

		// Group by AWS Service (EC2, S3, etc.)
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  &serviceKey,
			},
		},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cost data: %w", err)
	}

	// Initialize map with target services
	response := make(map[string]string)
	ec2Cost := 0.0
	s3Cost := 0.0
	lambdaCost := 0.0

	// Parse result and extract specific service costs
	for _, r := range result.ResultsByTime {
		for _, group := range r.Groups {
			service := group.Keys[0]
			costStr := *group.Metrics["UnblendedCost"].Amount

			var cost float64
			fmt.Sscanf(costStr, "%f", &cost)

			// Extract costs for specific services using flexible matching
			if strings.Contains(strings.ToLower(service), "ec2") || 
			   strings.Contains(service, "Elastic Compute Cloud") {
				ec2Cost += cost
			} else if strings.Contains(strings.ToLower(service), "s3") || 
			          strings.Contains(service, "Simple Storage Service") {
				s3Cost += cost
			} else if strings.Contains(strings.ToLower(service), "lambda") {
				lambdaCost += cost
			}
		}
	}

	// Format and return specific service costs
	response["EC2"] = fmt.Sprintf("$%.2f", ec2Cost)
	response["S3"] = fmt.Sprintf("$%.2f", s3Cost)
	response["Lambda"] = fmt.Sprintf("$%.2f", lambdaCost)
	response["TOTAL"] = fmt.Sprintf("$%.2f", ec2Cost+s3Cost+lambdaCost)

	return response, nil
}

// FetchEC2Instances returns count of EC2 instances
func FetchEC2Instances() (int, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Describe all EC2 instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch EC2 instances: %w", err)
	}

	instanceCount := 0
	// Count instances across all reservations
	for _, reservation := range result.Reservations {
		instanceCount += len(reservation.Instances)
	}

	return instanceCount, nil
}

// FetchAllServices returns all services with costs (for debugging)
func FetchAllServices() (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := costexplorer.NewFromConfig(cfg)

	// Date range (last 30 days)
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")
	serviceKey := "SERVICE"

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: &startStr,
			End:   &endStr,
		},
		Granularity: types.GranularityMonthly,
		Metrics:     []string{"UnblendedCost"},

		// Group by AWS Service (EC2, S3, etc.)
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  &serviceKey,
			},
		},
	}

	result, err := client.GetCostAndUsage(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cost data: %w", err)
	}

	response := make(map[string]string)

	// Return all services found
	for _, r := range result.ResultsByTime {
		for _, group := range r.Groups {
			service := group.Keys[0]
			costStr := *group.Metrics["UnblendedCost"].Amount
			response[service] = costStr
		}
	}

	return response, nil
}

// EC2InstanceDetail contains instance information and metrics
type EC2InstanceDetail struct {
	InstanceID       string
	InstanceType     string
	State            string
	CPUUtilization   string
	LaunchTime       string
}

// FetchEC2InstancesWithCPU returns EC2 instances count and CPU usage details
func FetchEC2InstancesWithCPU() ([]EC2InstanceDetail, int, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, 0, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Describe all EC2 instances
	result, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch EC2 instances: %w", err)
	}

	// Create CloudWatch client for CPU metrics
	cwClient := cloudwatch.NewFromConfig(cfg)

	var instances []EC2InstanceDetail
	instanceCount := 0

	// Collect instance details
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			instanceCount++

			state := "Unknown"
			if instance.State != nil {
				state = string(instance.State.Name)
			}

			launchTime := ""
			if instance.LaunchTime != nil {
				launchTime = instance.LaunchTime.Format("2006-01-02 15:04:05")
			}

			instanceType := ""
			if instance.InstanceType != "" {
				instanceType = string(instance.InstanceType)
			}

			// Fetch CPU utilization from CloudWatch
			cpuUtil := "N/A"
			if instance.InstanceId != nil {
				cpuUtil = getEC2CPUUtilization(cwClient, *instance.InstanceId)
			}

			detail := EC2InstanceDetail{
				InstanceID:     *instance.InstanceId,
				InstanceType:   instanceType,
				State:          state,
				CPUUtilization: cpuUtil,
				LaunchTime:     launchTime,
			}

			instances = append(instances, detail)
		}
	}

	return instances, instanceCount, nil
}

// getEC2CPUUtilization fetches CPU utilization for a specific instance from CloudWatch
func getEC2CPUUtilization(client *cloudwatch.Client, instanceID string) string {
	end := time.Now()
	start := end.Add(-1 * time.Hour)

	ns := "AWS/EC2"
	mn := "CPUUtilization"
	period := int32(300)
	dn := "InstanceId"

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  &ns,
		MetricName: &mn,
		StartTime:  &start,
		EndTime:    &end,
		Period:     &period,
		Statistics: []cloudwatch_types.Statistic{cloudwatch_types.StatisticAverage},
		Dimensions: []cloudwatch_types.Dimension{
			{
				Name:  &dn,
				Value: &instanceID,
			},
		},
	}

	result, err := client.GetMetricStatistics(context.TODO(), input)
	if err != nil || result == nil || len(result.Datapoints) == 0 {
		return "N/A"
	}

	if len(result.Datapoints) > 0 && result.Datapoints[0].Average != nil {
		return fmt.Sprintf("%.2f%%", *result.Datapoints[0].Average)
	}

	return "N/A"
}

// FetchSecurityGroupsCount returns count of security groups
func FetchSecurityGroupsCount() (int, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get all regions
	regions, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch regions: %w", err)
	}

	totalSecurityGroups := 0

	// Check security groups in each region
	for _, region := range regions.Regions {
		if region.RegionName == nil {
			continue
		}

		// Create client for this region
		regionCfg, _ := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = *region.RegionName
			return nil
		})
		regionClient := ec2.NewFromConfig(regionCfg)

		// Describe security groups in this region
		result, err := regionClient.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			continue
		}

		totalSecurityGroups += len(result.SecurityGroups)
	}

	return totalSecurityGroups, nil
}

// FetchKeyPairsCount returns count of key pairs across all regions
func FetchKeyPairsCount() (int, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return 0, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get all regions
	regions, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch regions: %w", err)
	}

	totalKeyPairs := 0

	// Check key pairs in each region
	for _, region := range regions.Regions {
		if region.RegionName == nil {
			continue
		}

		// Create client for this region
		regionCfg, _ := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = *region.RegionName
			return nil
		})
		regionClient := ec2.NewFromConfig(regionCfg)

		// Describe key pairs in this region
		result, err := regionClient.DescribeKeyPairs(context.TODO(), &ec2.DescribeKeyPairsInput{})
		if err != nil {
			continue
		}

		totalKeyPairs += len(result.KeyPairs)
	}

	return totalKeyPairs, nil
}

// SecurityGroupDetail contains security group info
type SecurityGroupDetail struct {
	GroupID     string
	GroupName   string
	Description string
	Region      string
}

// KeyPairDetail contains key pair info
type KeyPairDetail struct {
	KeyName string
	Region  string
}

// FetchSecurityGroupsDetails returns detailed list of all security groups
func FetchSecurityGroupsDetails() ([]SecurityGroupDetail, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get all regions
	regions, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	var allSGs []SecurityGroupDetail

	// Check security groups in each region
	for _, region := range regions.Regions {
		if region.RegionName == nil {
			continue
		}

		// Create client for this region
		regionCfg, _ := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = *region.RegionName
			return nil
		})
		regionClient := ec2.NewFromConfig(regionCfg)

		// Describe security groups in this region
		result, err := regionClient.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{})
		if err != nil {
			continue
		}

		for _, sg := range result.SecurityGroups {
			detail := SecurityGroupDetail{
				GroupID:     *sg.GroupId,
				GroupName:   *sg.GroupName,
				Region:      *region.RegionName,
				Description: "",
			}
			if sg.Description != nil {
				detail.Description = *sg.Description
			}
			allSGs = append(allSGs, detail)
		}
	}

	return allSGs, nil
}

// FetchKeyPairsDetails returns detailed list of all key pairs
func FetchKeyPairsDetails() ([]KeyPairDetail, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	// Get all regions
	regions, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	var allKPs []KeyPairDetail

	// Check key pairs in each region
	for _, region := range regions.Regions {
		if region.RegionName == nil {
			continue
		}

		// Create client for this region
		regionCfg, _ := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
			o.Region = *region.RegionName
			return nil
		})
		regionClient := ec2.NewFromConfig(regionCfg)

		// Describe key pairs in this region
		result, err := regionClient.DescribeKeyPairs(context.TODO(), &ec2.DescribeKeyPairsInput{})
		if err != nil {
			continue
		}

		for _, kp := range result.KeyPairs {
			detail := KeyPairDetail{
				KeyName: *kp.KeyName,
				Region:  *region.RegionName,
			}
			allKPs = append(allKPs, detail)
		}
	}

	return allKPs, nil
}