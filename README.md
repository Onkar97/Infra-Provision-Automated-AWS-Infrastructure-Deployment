# InfraProvision – Automated AWS Infrastructure Deployment
## Overview
InfraProvision is a project demonstrating the principles of Infrastructure as Code (IaC) using Terraform to automate the deployment of a scalable and reproducible web application infrastructure on Amazon Web Services (AWS). This project defines and manages a complete cloud environment, from networking foundations to compute and database resources, all through version-controlled code.

The primary goal is to eliminate manual configuration, reduce deployment errors, and create a consistent environment that can be provisioned and destroyed on demand.

## Core Features
Automated Infrastructure Provisioning: Uses Terraform modules to define and create a complete AWS environment, including:

A custom Virtual Private Cloud (VPC) with public and private subnets.

Security Groups to control inbound and outbound traffic.

EC2 instances for application hosting.

A managed RDS PostgreSQL instance for the database layer.

CI/CD for Infrastructure: Integrates with GitHub Actions to create a seamless CI/CD pipeline.

terraform plan is automatically run on pull requests to validate changes.

terraform apply is triggered on merge to the main branch to deploy changes.

Secure Credential Management: Manages sensitive data like database passwords and API keys securely using Terraform variables and secrets management best practices, avoiding hardcoded values in the codebase.

## Technology Stack
Infrastructure as Code: Terraform

Cloud Provider: AWS

Compute: EC2 (Elastic Compute Cloud)

Networking: VPC (Virtual Private Cloud)

Database: RDS (Relational Database Service) for PostgreSQL

CI/CD: GitHub Actions

## Getting Started
## Prerequisites
Before you begin, ensure you have the following installed:

Terraform (v1.0.0 or later)

AWS CLI

An AWS account with appropriate IAM permissions.

Installation & Configuration
Clone the repository:

git clone [YOUR_GITHUB_REPOSITORY_LINK]
cd InfraProvision

Configure AWS Credentials:
Make sure your AWS CLI is configured with credentials. Terraform will use these to provision resources.

aws configure

Create a Variables File:
Create a file named terraform.tfvars to provide values for the variables defined in variables.tf. This file should be added to your .gitignore to prevent committing sensitive information.

## example terraform.tfvars
aws_region      = "us-east-1"
db_password     = "your-secure-password"
instance_type   = "t2.micro"

## Usage
Initialize Terraform:
Initialize the working directory to download the necessary providers and modules.

terraform init

Plan the Deployment:
Run a plan to see what infrastructure will be created, changed, or destroyed.

terraform plan

Apply the Configuration:
Apply the plan to provision the resources in your AWS account.

terraform apply

Type yes when prompted to confirm the changes.

Destroy the Infrastructure:
When you are finished, you can destroy all the created resources to avoid incurring further costs.

terraform destroy

CI/CD Pipeline
This project uses GitHub Actions to automate the Terraform workflow. The pipeline is defined in .github/workflows/terraform.yml.

On Pull Request: A workflow is triggered that runs terraform init and terraform plan. The output of the plan is added as a comment to the pull request for review.

On Merge to main: Upon merging a pull request, the workflow runs terraform init and terraform apply -auto-approve to automatically deploy the changes to the infrastructure.
