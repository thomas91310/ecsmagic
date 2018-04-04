package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/urfave/cli"
)

func main() {
	app := newApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "servicename",
			Value: "something",
		},
		cli.StringFlag{
			Name:  "cluster",
			Value: "production",
		},
		cli.StringFlag{
			Name:  "status",
			Value: "RUNNING",
		},
		cli.StringFlag{
			Name:  "username",
			Value: "whoami",
		},
		cli.StringFlag{
			Name:  "ssh_password_key",
			Value: "something",
		},
		cli.StringFlag{
			Name:  "private_key_path",
			Value: "something_private",
		},
	}

	app.Action = func(c *cli.Context) error {
		cluster := aws.String(c.String("cluster"))
		desiredStatus := aws.String(c.String("status"))
		sshConf := SSHConf{
			Username:       c.String("username"),
			PasswordKey:    c.String("ssh_password_key"),
			PrivateKeyPath: c.String("private_key_path"),
		}

		sess := session.Must(session.NewSession())

		ecsSvc := ecs.New(sess, &aws.Config{
			Region: aws.String(endpoints.UsEast1RegionID),
		})
		ec2Svc := ec2.New(sess, &aws.Config{
			Region: aws.String(endpoints.UsEast1RegionID),
		})

		serviceName := ""
		err := ecsSvc.ListServicesPages(&ecs.ListServicesInput{
			Cluster: cluster,
		}, func(page *ecs.ListServicesOutput, lastPage bool) bool {
			for _, service := range page.ServiceArns {
				if strings.Contains(*service, c.String("servicename")) {
					serviceName = *service
				}
			}
			return true
		})
		if err != nil {
			return fmt.Errorf("error listing pages for cluster %v, got %v", cluster, err)
		}

		if serviceName == "" {
			return fmt.Errorf("no service for service name %v for cluster %v", c.String("servicename"), c.String("cluster"))
		}

		tasks := []*string{}
		err = ecsSvc.ListTasksPages(&ecs.ListTasksInput{
			DesiredStatus: desiredStatus,
			Cluster:       cluster,
			ServiceName:   aws.String(serviceName),
		}, func(page *ecs.ListTasksOutput, lastPage bool) bool {
			tasks = append(tasks, page.TaskArns...)
			return true
		})
		if err != nil {
			return fmt.Errorf("error listing tasks for service name %v, cluster %v and desired status %v, got %v", serviceName, *cluster, *desiredStatus, err)
		}

		if len(tasks) == 0 {
			return fmt.Errorf("No tasks for cluster %v and service %v and desired status %v", *cluster, serviceName, *desiredStatus)
		}

		containerInstancesArns := []*string{}
		respDescribe, errDescribe := ecsSvc.DescribeTasks(&ecs.DescribeTasksInput{
			Tasks:   tasks,
			Cluster: cluster,
		})
		if errDescribe != nil {
			return fmt.Errorf("error describing tasks, got %v", err)
		}

		for _, describedTask := range respDescribe.Tasks {
			containerInstancesArns = append(containerInstancesArns, describedTask.ContainerInstanceArn)
		}

		respContainers, errContainers := ecsSvc.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			ContainerInstances: containerInstancesArns,
			Cluster:            cluster,
		})
		if errContainers != nil {
			return fmt.Errorf("error describing container instances, got %v", err)
		}

		instanceIDs := []*string{}
		for _, container := range respContainers.ContainerInstances {
			instanceIDs = append(instanceIDs, container.Ec2InstanceId)
		}

		respEc2, errEc2 := ec2Svc.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: instanceIDs,
		})
		if errEc2 != nil {
			return fmt.Errorf("error listing instances, got %v", err)
		}

		containers := []*ECSContainer{}
		for _, resa := range respEc2.Reservations {
			for _, instance := range resa.Instances {
				wantedContainers, err := CallECSAgent(*desiredStatus, tasks, instance.PrivateIpAddress)
				if err != nil {
					return err
				}
				containers = append(containers, wantedContainers...)
			}
		}

		newMenu(sshConf, containers)

		return nil
	}

	return app
}
