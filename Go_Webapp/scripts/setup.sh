#!/bin/bash

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to prompt for input with default value
prompt_input() {
    local varname=$1
    local prompt_text=$2
    local is_password=$4
    
    if [ "$is_password" = "true" ]; then
        read -sp "$prompt_text: " input_value
        echo
    else
        read -p "$prompt_text: " input_value
    fi
    
    eval "$varname=\"$input_value\""
}

print_message "Please provide the following configuration values"

# Prompt for all configuration values
prompt_input "DB_PASSWORD" "Enter database password" "" "true"
prompt_input "DB_USER" "Enter database user"
prompt_input "DB_NAME" "Enter database name"
prompt_input "DB_HOST" "Enter database host"
prompt_input "DB_PORT" "Enter database port" 
prompt_input "APP_PORT" "Enter application port"

APP_USER="csye6225";
APP_GROUP="csye6225";
APP_DIR="/opt/csye6225";
GO_VERSION="1.21.6"; # Updated from NODE_VERSION
APP_ARCHIVE="webapp.zip";

print_message "Configuration received. Starting deployment..."

# 1. Update Package Lists
print_message "Step 1: Updating package lists..."
sudo apt update -y
print_message "Package lists updated successfully"

# 2. Upgrade System Packages
print_message "Step 2: Upgrading system packages..."
sudo apt upgrade -y
print_message "System packages upgraded successfully"

# 3. Install Database Management System
print_message "Step 3: Installing PostgreSQL..."
sudo apt install -y postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
print_message "PostgreSQL installed and started"

# 4. Create Application Database
print_message "Step 4: Creating PostgreSQL database and user..."

sudo -u postgres psql -c "CREATE DATABASE \"${DB_NAME}\";" || print_warning "Database already exists"

USER_EXISTS=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'")

if [ "$USER_EXISTS" = "1" ]; then
    print_warning "User already exists, updating password..."
    sudo -u postgres psql -c "ALTER USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';"
    print_message "Password updated successfully"
else
    print_message "Creating new user..."
    sudo -u postgres psql -c "CREATE USER ${DB_USER} WITH PASSWORD '${DB_PASSWORD}';"
    print_message "User created successfully"
fi

sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE \"${DB_NAME}\" TO ${DB_USER};"
print_message "Privileges granted successfully"

# 5. Create Application Linux Group
print_message "Step 5: Creating application group..."
sudo groupadd ${APP_GROUP}

# 6. Create Application User Account
print_message "Step 6: Creating application user account..."
sudo useradd -g ${APP_GROUP} ${APP_USER}
print_message "User created successfully"

# 7. Deploy Application Files
print_message "Step 7: Deploying application files..."

if [ ! -d "${APP_DIR}" ]; then
    mkdir -p ${APP_DIR}
    print_message "Created directory ${APP_DIR}"
fi

if [ -f "${APP_ARCHIVE}" ]; then
    print_message "Extracting application files..."
    
    sudo apt install -y unzip
    
    sudo unzip -o ${APP_ARCHIVE} -d ${APP_DIR} 
    print_message "Application files extracted"
    
    # Handle nested webapp folder
    if [ -d "${APP_DIR}/webapp" ]; then
        print_message "Reorganizing directory structure..."
        cp -r ${APP_DIR}/webapp/* ${APP_DIR}/ || true
        cp ${APP_DIR}/webapp/.* ${APP_DIR}/ || true
        rm -rf ${APP_DIR}/webapp
    fi
else
    print_error "Application archive '${APP_ARCHIVE}' not found"
    exit 1
fi

# 8. Create .env File
print_message "Step 8: Creating .env file..."

cat > ${APP_DIR}/.env << EOF
PORT=${APP_PORT}
DBUSER=${DB_USER}
DBHOST=${DB_HOST}
DBNAME=${DB_NAME}
DBPASSWORD=${DB_PASSWORD}
DBPORT=${DB_PORT}
EOF

sudo chmod 600 ${APP_DIR}/.env
sudo chown ${APP_USER}:${APP_GROUP} ${APP_DIR}/.env
print_message ".env file created securely"

# 9. Set File Permissions
print_message "Step 9: Setting file permissions..."

sudo chown -R ${APP_USER}:${APP_GROUP} ${APP_DIR}
# Fix: Ensure directories have execute permission (755) so they can be accessed
sudo find ${APP_DIR} -type d -exec chmod 755 {} \;
sudo find ${APP_DIR} -type f -exec chmod 644 {} \;
sudo chmod 600 ${APP_DIR}/.env

print_message "Permissions set successfully"

# 10. Install Go (Replaces Node.js)
print_message "Step 10: Installing Go ${GO_VERSION}..."

# Download Go tarball
wget "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -O go.tar.gz

# Remove previous installation and extract new one
sudo rm -rf /usr/local/go 
sudo tar -C /usr/local -xzf go.tar.gz
rm go.tar.gz

# Add Go to PATH for the current session so we can build immediately
export PATH=$PATH:/usr/local/go/bin

# Verify installation
GO_INSTALLED_VERSION=$(go version)
print_message "Go installed: ${GO_INSTALLED_VERSION}"

# 11. Build and Run Application
print_message "Step 11: Building and running application..."

cd ${APP_DIR}

# Download Go module dependencies
print_message "Downloading Go modules..."
go mod download

# Build the application binary (output named 'webapp')
print_message "Building binary..."
go build -o webapp .

# Fix permissions: The binary must be executable
sudo chown ${APP_USER}:${APP_GROUP} webapp
sudo chmod +x webapp

print_message "Running the application"
# Using nohup to run in background. 
# Ideally, this should be a systemd service, but sticking to your script pattern:
sudo nohup ./webapp > app.log 2>&1 &

print_message "Service is running on port ${APP_PORT}"
print_message "Deployment completed successfully"

exit 0