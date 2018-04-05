package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// ecsStuff gets the ECSContainer(s) for desiredStatus tasks on cluster
func ecsStuff(ecsService string, cluster string, desiredStatus string) ([]*ECSContainer, error) {
	sess := session.Must(session.NewSession())

	ecsSvc := ecs.New(sess, &aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})
	ec2Svc := ec2.New(sess, &aws.Config{
		Region: aws.String(endpoints.UsEast1RegionID),
	})

	serviceName := ""
	err := ecsSvc.ListServicesPages(&ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}, func(page *ecs.ListServicesOutput, lastPage bool) bool {
		for _, service := range page.ServiceArns {
			if strings.Contains(*service, ecsService) {
				serviceName = *service
			}
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing pages for cluster %v, got %v", cluster, err)
	}

	if serviceName == "" {
		return nil, fmt.Errorf("no service for service name %v for cluster %v", ecsService, cluster)
	}

	tasks := []*string{}
	err = ecsSvc.ListTasksPages(&ecs.ListTasksInput{
		DesiredStatus: aws.String(desiredStatus),
		Cluster:       aws.String(cluster),
		ServiceName:   aws.String(serviceName),
	}, func(page *ecs.ListTasksOutput, lastPage bool) bool {
		tasks = append(tasks, page.TaskArns...)
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("error listing tasks for service name %v, cluster %v and desired status %v, got %v", serviceName, cluster, desiredStatus, err)
	}

	if len(tasks) == 0 {
		return nil, fmt.Errorf("No tasks for cluster %v and service %v and desired status %v", cluster, serviceName, desiredStatus)
	}

	containerInstancesArns := []*string{}
	respDescribe, errDescribe := ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
		Tasks:   tasks,
		Cluster: aws.String(cluster),
	})
	if errDescribe != nil {
		return nil, fmt.Errorf("error describing tasks, got %v", err)
	}

	for _, describedTask := range respDescribe.Tasks {
		containerInstancesArns = append(containerInstancesArns, describedTask.ContainerInstanceArn)
	}

	respContainers, errContainers := ecsSvc.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		ContainerInstances: containerInstancesArns,
		Cluster:            aws.String(cluster),
	})
	if errContainers != nil {
		return nil, fmt.Errorf("error describing container instances, got %v", err)
	}

	instanceIDs := []*string{}
	for _, container := range respContainers.ContainerInstances {
		instanceIDs = append(instanceIDs, container.Ec2InstanceId)
	}

	respEc2, errEc2 := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	})
	if errEc2 != nil {
		return nil, fmt.Errorf("error listing instances, got %v", err)
	}

	containers := []*ECSContainer{}
	for _, resa := range respEc2.Reservations {
		for _, instance := range resa.Instances {
			wantedContainers, err := callECSAgent(desiredStatus, tasks, instance.PrivateIpAddress)
			if err != nil {
				return nil, err
			}
			containers = append(containers, wantedContainers...)
		}
	}

	return containers, nil
}
