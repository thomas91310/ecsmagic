package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	protocol = "http"
	port     = "51678"
)

// ECSEnveloppe defines the JSON ECS enveloppe in the body
type ECSEnveloppe struct {
	Tasks     []ECSTask `json:"tasks"`
	PrivateIP string
}

// ECSTask defines an ECSTask
type ECSTask struct {
	ARN           string          `json:"Arn"`
	Containers    []*ECSContainer `json:"Containers"`
	DesiredStatus string          `json:"DesiredStatus"`
	Family        string          `json:"Family"`
	KnownStatus   string          `json:"KnownStatus"`
	Version       string          `json:"Version"`
}

// ECSContainer defines an ECSContainer
type ECSContainer struct {
	DockerCID   string `json:"DockerId"`
	DockerCName string `json:"DockerName"`
	Name        string `json:"Name"`
	PrivateIP   string
	TaskName    string
}

// Print prints an ecs container
func (c ECSContainer) Print() {
	fmt.Println("docker container id: ", c.DockerCID)
	fmt.Println("docker container name: ", c.DockerCName)
	fmt.Println("name: ", c.Name)
	fmt.Println("privateIP of running container: ", c.PrivateIP)
}

// CallECSAgent calls the metadata service to get the DockerCID
func CallECSAgent(desiredStatus string, tasks []*string, privateIP *string) ([]*ECSContainer, error) {
	requestURL := fmt.Sprintf("%v://%v:%v/v1/tasks", protocol, *privateIP, port)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, fmt.Errorf("Error getting running tasks from ecs agent")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body for request %v", requestURL)
	}

	ecsEnveloppe := new(ECSEnveloppe)
	err = json.Unmarshal(body, ecsEnveloppe)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response from request %v", requestURL)
	}

	containers := []*ECSContainer{}
	for _, task := range ecsEnveloppe.Tasks {
		if task.DesiredStatus != desiredStatus {
			continue
		}
		for _, wantedTask := range tasks {
			if *wantedTask == task.ARN {
				for _, c := range task.Containers {
					c.PrivateIP = *privateIP
					c.TaskName = task.ARN
				}
				containers = append(containers, task.Containers...)
			}
		}
	}

	return containers, nil
}
