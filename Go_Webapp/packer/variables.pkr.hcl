variable "aws_region" {
  type        = string
  description = "The AWS region to build the AMI in."
  default     = "us-east-1"
}

variable "source_ami" {
  type        = string
  description = "The search pattern for the source Ubuntu AMI."
}

variable "ssh_username" {
  type        = string
  description = "The SSH username for the source AMI."
  default     = "ubuntu"
}

variable "ubuntu_owner" {
  type        = string
  description = "The AWS Account ID of Canonical, the publisher of Ubuntu."
}

variable "instance_type" {
  type        = string
  description = "The EC2 instance type to use for the build."
  default     = "t2.micro"
}

variable "aws_dev_id" {
  type        = string
  description = "The 12-digit AWS Account ID of the Dev account."
}

variable "aws_demo_id" {
  type        = string
  description = "The 12-digit AWS Account ID of the Demo account to share the AMI with."
}

variable "ami_name" {
  type        = string
  description = "The name of the custom AMI that will be created."
  default     = "webapp-csye6225-onkar"
}


//Variables to remove:

variable "db_password" {
  type    = string
  default = null
}

variable "db_name" {
  type    = string
  default = null
}

variable "app_port" {
  type    = string
  default = null
}

variable "db_user" {
  type    = string
  default = null
}

variable "db_port" {
  type    = string
  default = null
}

variable "db_host" {
  type    = string
  default = null
}

