package service

import "github.com/user/cost-optimizer/internal/aws"

func GetCostData() (map[string]string, error) {
	return aws.FetchCost()
}

func GetEC2InstancesWithCPU() ([]aws.EC2InstanceDetail, int, error) {
	return aws.FetchEC2InstancesWithCPU()
}

func GetAllServices() (map[string]string, error) {
	return aws.FetchAllServices()
}

func GetSecurityGroupsCount() (int, error) {
	return aws.FetchSecurityGroupsCount()
}

func GetKeyPairsCount() (int, error) {
	return aws.FetchKeyPairsCount()
}

func GetSecurityGroupsDetails() ([]aws.SecurityGroupDetail, error) {
	return aws.FetchSecurityGroupsDetails()
}

func GetKeyPairsDetails() ([]aws.KeyPairDetail, error) {
	return aws.FetchKeyPairsDetails()
}