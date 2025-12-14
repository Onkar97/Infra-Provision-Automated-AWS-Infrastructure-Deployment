# Nexus-Serve: AWS Cloud Native Full Stack Automation Platform
This project is a highly available microservices application engineered with Go (Golang) and deployed on AWS. It demonstrates a robust, cloud-native architecture designed for scalability, security, and automated deployment.

## Architecture & Infrastructure


* **High Availability:** Deployed across **3 Availability Zones (AZs)** using EC2 Auto Scaling Groups to ensure fault tolerance and seamless scaling.
* **Traffic Management:** Utilizes an **Application Load Balancer (ALB)** to efficiently distribute incoming traffic across microservices.
* **Compute:**
    * **EC2:** Hosting core microservices.
    * **AWS Lambda:** Implementing event-driven functions for asynchronous tasks.
* **Infrastructure as Code (IaC):** Complete infrastructure provisioning managed via **Terraform**.
* **State Management:** Terraform state is securely managed using **DynamoDB** for locking and consistency.

## Application Features
* **Microservices:** Built with Go, implementing **20+ RESTful APIs**.
* **Event-Driven Architecture:** Integrated Lambda functions to handle event-based workflows.
* **Observability:** Comprehensive monitoring and logging implemented via **Amazon CloudWatch** for real-time insights and troubleshooting.

## Security & DevOps
* **CI/CD Pipeline:** Fully automated deployment pipeline for consistent and reliable software delivery.
* **Access Control:** Implemented **15+ IAM Role-Based Access Control (RBAC)** policies to enforce least privilege principles.
* **Data Security:** Applied **AWS KMS (Key Management Service)** encryption for safeguarding sensitive data at rest and in transit.

## Tech Stack
* **Language:** Go (Golang)
* **Cloud Provider:** AWS (EC2, Lambda, ALB, DynamoDB, CloudWatch, IAM, KMS)
* **IaC Tool:** Terraform
