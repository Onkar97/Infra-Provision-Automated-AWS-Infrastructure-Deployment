#!/bin/bash

print_message() { echo -e "\e[32m[INFO]\e[0m $1"; }
print_error() { echo -e "\e[31m[ERROR]\e[0m $1"; }

# Static application settings
APP_USER="csye6225"
APP_GROUP="csye6225"
APP_DIR="/opt/csye6225"
APP_ARCHIVE="/tmp/webapp.zip"
GO_VERSION="1.21.6" # Define Go version here

print_message "Starting application setup script..."

# 1. Update Packages & Install Prerequisites
print_message "Step 1: Updating packages and installing dependencies..."
sudo apt-get update -y
sudo apt-get upgrade -y
# Explicitly enable all standard repositories
sudo add-apt-repository main
sudo add-apt-repository universe
sudo add-apt-repository restricted
sudo add-apt-repository multiverse
sudo apt-get update -y
# Removed 'nodejs' and 'npm'. Added 'wget' for downloading Go.
sudo apt-get install -y ca-certificates curl gnupg unzip build-essential wget
print_message "Prerequisites installed."

# 2. Install Go (Replaces Node.js)
print_message "Step 2: Installing Go ${GO_VERSION}..."
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O go.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go.tar.gz
rm go.tar.gz

# Add Go to PATH temporarily for this script execution
export PATH=$PATH:/usr/local/go/bin
print_message "Go $(go version) installed successfully."

# 3. Install the CloudWatch Agent
print_message "Step 3: Installing the CloudWatch Agent..."
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
sudo dpkg -i -E ./amazon-cloudwatch-agent.deb
sudo systemctl enable amazon-cloudwatch-agent
print_message "CloudWatch Agent installed and enabled."

# 4. Create Application User and Deploy Files
print_message "Step 4: Creating application user and deploying files..."
sudo groupadd -f ${APP_GROUP}
# Use '|| true' to prevent script failure if the user already exists
sudo useradd -r -g ${APP_GROUP} -d ${APP_DIR} -s /usr/sbin/nologin ${APP_USER} || true
sudo mkdir -p ${APP_DIR}
sudo unzip -o ${APP_ARCHIVE} -d ${APP_DIR}
print_message "Application files extracted to ${APP_DIR}."

# 5. Prepare Log Directories and Files
print_message "Step 5: Preparing log directories and files..."
# Create the directory for the CloudWatch Agent's own log file
sudo mkdir -p /var/logs
# Create the application's log file so the agent can find it on startup
sudo touch ${APP_DIR}/webapp.log
# Set ownership so the 'csye6225' user can write to the log file
sudo chown ${APP_USER}:${APP_GROUP} ${APP_DIR}/webapp.log
print_message "Log files and directories created and permissions set."

# 6. Handle nested webapp folder
if [ -d "${APP_DIR}/webapp" ]; then
    print_message "Reorganizing directory structure..."
    sudo mv ${APP_DIR}/webapp/* ${APP_DIR}/
    sudo rm -rf ${APP_DIR}/webapp
fi

# 7. Place CloudWatch Agent Config in Default Location
print_message "Step 6: Moving CloudWatch Agent config file to default directory..."
# The config file was uploaded by Packer's 'file' provisioner.
# We move it to the official path the agent looks for by default.
sudo mv /tmp/config-cloudwatch-agent.json /opt/csye6225/config-cloudwatch-agent.json
print_message "CloudWatch config moved to default agent directory."

# 8. Build Application (Replaces npm install)
print_message "Step 7: Downloading modules and building binary..."
cd ${APP_DIR}

# Ensure Go is in the path for the build command
export PATH=$PATH:/usr/local/go/bin

# Download Go dependencies
sudo /usr/local/go/bin/go mod download

# Build the binary named 'webapp'
print_message "Compiling Go binary..."
sudo /usr/local/go/bin/go build -o webapp .

print_message "Go application built successfully."

# 9. Set Final Permissions
print_message "Step 8: Setting final file permissions..."
# Ensure the app user owns the new binary
sudo chown -R ${APP_USER}:${APP_GROUP} ${APP_DIR}
sudo chmod -R 755 ${APP_DIR}
# Specifically ensure the binary is executable
sudo chmod +x ${APP_DIR}/webapp
print_message "Permissions set."

# 10. Configure Systemd Service
print_message "Step 9: Creating and enabling systemd service..."

# We don't need NODE_EXEC_PATH checks anymore.
# We point directly to our compiled binary: /opt/csye6225/webapp

sudo tee /etc/systemd/system/webapp.service > /dev/null <<EOF
[Unit]
Description=CSYE6225 Web Application Service
# This is a critical fix: waits for user_data to complete
After=cloud-init.service

[Service]
# This is a critical fix: points to the correct environment file
EnvironmentFile=${APP_DIR}/.env
User=${APP_USER}
Group=${APP_GROUP}
Type=simple
WorkingDirectory=${APP_DIR}
# CHANGED: Point to the compiled Go binary
ExecStart=${APP_DIR}/webapp
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable webapp.service
print_message "Systemd service 'webapp.service' configured and enabled."

# 11. Final Cleanup
print_message "Step 10: Cleaning up..."
sudo apt-get clean
sudo rm -f ${APP_ARCHIVE}
sudo rm -f ./amazon-cloudwatch-agent.deb
# Optional: Remove Go compiler to save space if only the binary is needed for runtime
# sudo rm -rf /usr/local/go 
print_message "Setup complete. The image is ready."

exit 0