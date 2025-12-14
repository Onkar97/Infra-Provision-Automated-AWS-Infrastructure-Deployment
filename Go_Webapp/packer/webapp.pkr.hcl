packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "webapp" {
  ami_name      = "webapp-csye6225-onkar-${formatdate("YYYY-MM-DD-hh-mm-ss", timestamp())}"
  instance_type = "${var.instance_type}"
  region        = "${var.aws_region}"
  source_ami_filter {
    filters = {
      name                = "${var.source_ami}"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["${var.ubuntu_owner}"]
  }
  ami_users    = ["${var.aws_dev_id}", "${var.aws_demo_id}"]
  ssh_username = "${var.ssh_username}"
}

build {
  name = "start-webapp"
  sources = [
    "source.amazon-ebs.webapp"
  ]


  provisioner "file" {
    source      = "webapp.zip"
    destination = "/tmp/webapp.zip"
  }

  provisioner "file" {
    source      = "packer/config-cloudwatch-agent.json"
    destination = "/tmp/config-cloudwatch-agent.json"
  }

  provisioner "shell" {
    script = "scripts/webapp_install.sh"
  }

  post-processor "manifest" {
    output     = "manifest.json"
    strip_path = true
  }
}