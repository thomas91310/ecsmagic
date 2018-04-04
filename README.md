## What I use this for

I use this cli to know where our ECS tasks are running inside our ECS cluster (which EC2 instance).  
Since our ECS cluster is composed of several EC2 instances, this is sometimes challenging to identify which task is running on which server inside the cluster.

This tool aims to return the private IPs of the EC2 instances that run tasks for the service you provide as a parameter using `--servicename`.
You can also leverage the `--cluster` parameter if you have multiple ECS clusters.

## Examples
```
./ecsmagic --cluster=production --servicename=super-service
0) Access container id 4ccf9423dec0 on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
1) Access container id 6c4c4d74d56c on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
2) Access container id 045728208def on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
What container do you want to `ssh` into?
```

```
./ecsmagic --cluster=staging --servicename=super-service
0) Access container id 746ec62d7d2c on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
What container do you want to `ssh` into?
```

## SSH access
The ssh access works but I do not think it is that useful. If you want it to work, you have to specify 3 additional parameters:
* `--username` which is your ssh username
* `--private_key_path` which is the path for the private ssh key that you will need to authenticate to the server
* `--ssh_password_key` which is the password for the private encrypted ssh key

```
./ecsmagic --cluster=production --servicename=super-service --username=**** --private_key_path=/Users/****/.ssh/**** --ssh_password_key=****
0) Access container id 746ec62d7d2c on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
1) Access container id 6c4c4d74d56c on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
2) Access container id 045728208def on host * running task arn:aws:ecs:us-east-1:account_id:task/task_id
What container do you want to `ssh` into?
0
Last login: Wed Apr  4 22:17:45 2018 from *
   __|  __|  __|
   _|  (   \__ \   Amazon ECS-Optimized Amazon Linux AMI 2017.09.c
 ____|\___|____/
 For documentation visit, http://aws.amazon.com/documentation/ecs
[myname@* ~]$
```
