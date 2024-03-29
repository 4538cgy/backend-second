GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
INSTALL_DIR=/vcom/backend/api
LOG_DIR=/vcom/backend/api/log
CONF_FILE=config.toml
FIREBASE_SERVICE_ACCOUNT_KEY=firebase_adminsdk.json

BINARY_NAME=vcom_api

all: test build
build:
		$(GOBUILD) -o $(BINARY_NAME) -v
install:
		sudo mkdir -p /vcom
		sudo mkdir -p /vcom/backend
		sudo mkdir -p /vcom/backend/api
		sudo mkdir -p /vcom/backend/api/log
		sudo mkdir -p /vcom/backend/api/download
		sudo mkdir -p /vcom/backend/asset
		sudo mkdir -p /vcom/backend/asset/profile
		sudo cp -f $(BINARY_NAME) /vcom/backend/api/$(BINARY_NAME)
test:
		$(GOTEST) -v ./...
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
configs:
		sudo cp -f $(CONF_FILE) $(INSTALL_DIR)/$(CONF_FILE)
		sudo cp -f $(FIREBASE_SERVICE_ACCOUNT_KEY) $(INSTALL_DIR)/$(FIREBASE_SERVICE_ACCOUNT_KEY)