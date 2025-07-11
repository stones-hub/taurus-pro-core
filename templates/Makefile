# 默认环境变量文件
env_file=.env.local

# 检查环境变量文件是否存在
ifeq ($(wildcard $(env_file)),)
$(error $(RED)Environment file '$(env_file)' not found. Please create it or specify a different file.$(RESET))
endif


# ---------------------------- 加载环境变量 --------------------------------
include $(env_file)
export $(shell sed 's/=.*//' $(env_file))

# 定义颜色
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
CYAN := \033[36m
RESET := \033[0m
# 定义分割符
SEPARATOR := $(CYAN)--------------------------------$(RESET)

# ---------------------------- 从环境变量中获取配置 --------------------------------

# 检查环境变量是否为空
ifeq ($(strip $(VERSION)),)
$(error $(RED)VERSION is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(APP_NAME)),)
$(error $(RED)APP_NAME is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(SERVER_ADDRESS)),)
$(error $(RED)SERVER_ADDRESS is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(SERVER_PORT)),)
$(error $(RED)SERVER_PORT is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(APP_CONFIG)),)
$(error $(RED)APP_CONFIG is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(HOST_PORT)),)
$(error $(RED)HOST_PORT is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(CONTAINER_PORT)),)
$(error $(RED)CONTAINER_PORT is not set in the environment variables$(RESET))
endif

ifeq ($(strip $(WORKDIR)),)
$(error $(RED)WORKDIR is not set in the environment variables$(RESET))
endif

BUILD_DIR := build
DOCKER_IMAGE := $(APP_NAME):$(VERSION)
DOCKER_CONTAINER := $(APP_NAME)
DOCKER_NETWORK := $(APP_NAME)-network
DOCKER_LOG_VOLUME := $(APP_NAME)_log
DOCKER_DOWNLOAD_VOLUME := $(APP_NAME)_download

# 定义发布目录
RELEASE_DIR := release
RELEASE_FILE_NAME := $(APP_NAME)-$(VERSION)
PACKAGE_DIR := $(RELEASE_DIR)/$(RELEASE_FILE_NAME)

# ---------------------------- 构建目标 --------------------------------
.PHONY: all build clean docker-run docker-stop local-run local-stop docker-compose-up docker-compose-down docker-compose-start docker-compose-stop docker-image-push docker-swarm-up docker-swarm-down docker-update-app docker-swarm-deploy-app local-release local-release-start local-release-stop local-release-logs local-release-status local-release-restart wire run 
# Default target
all: build

wire:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Generating wire code...$(RESET)"
	@./scripts/wire/gen_wire -project-root $(PWD)
	@wire ./internal/taurus/wire.go
	@echo -e "$(GREEN)Wire code generated.$(RESET)"
	@echo -e "$(SEPARATOR)"

run: wire
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Running the application...$(RESET)"
	@go run ./bin/taurus.go
	@echo -e "$(SEPARATOR)"

# Build the Go application
build: wire
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Building the application...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./bin/taurus.go
	@echo -e "$(GREEN)Build complete. Binary is located at $(BUILD_DIR)/$(APP_NAME)$(RESET)"
	@echo -e "$(SEPARATOR)"

# Clean up build artifacts
clean:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(YELLOW)Cleaning up...$(RESET)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(RELEASE_DIR)
	@echo -e "$(GREEN)Clean complete.$(RESET)"
	@echo -e "$(SEPARATOR)"

# Run the application locally in background with log file
local-run: clean build
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Running the application locally in background...$(RESET)"
	@mkdir -p logs
	@nohup $(BUILD_DIR)/$(APP_NAME) -config=$(APP_CONFIG) -env=$(env_file) > logs/app.log 2>&1 & echo $$! > logs/app.pid
	@echo -e "$(GREEN)Application started in background. PID: $$(cat logs/app.pid)$(RESET)"
	@echo -e "$(GREEN)Log file: logs/app.log$(RESET)"
	@echo -e "$(BLUE)To view logs: tail -f logs/app.log$(RESET)"
	@echo -e "$(BLUE)To stop: make local-stop$(RESET)"
	@echo -e "$(SEPARATOR)"

# Stop the local application (if running in the background)
local-stop:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(YELLOW)Stopping the local application...$(RESET)"
	@if [ -f logs/app.pid ]; then \
		PID=$$(cat logs/app.pid); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID; \
			echo -e "$(GREEN)Application (PID: $$PID) stopped.$(RESET)"; \
		else \
			echo -e "$(YELLOW)Application (PID: $$PID) is not running.$(RESET)"; \
		fi; \
		rm -f logs/app.pid; \
	else \
		pkill -f "$(BUILD_DIR)/$(APP_NAME)" || echo -e "$(RED)No local application is running.$(RESET)"; \
	fi
	@echo -e "$(GREEN)Local application stopped.$(RESET)"
	@echo -e "$(SEPARATOR)"

# 打包项目并创建发布包
local-release: clean build
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Packaging the application...$(RESET)"
	@mkdir -p $(PACKAGE_DIR)
	
	# 复制可执行文件
	@cp $(BUILD_DIR)/$(APP_NAME) $(PACKAGE_DIR)/
	
	# 复制环境配置文件
	@cp $(env_file) $(PACKAGE_DIR)/.env.local
	
	# 创建必要的目录
	@mkdir -p $(PACKAGE_DIR)/logs
	@mkdir -p $(PACKAGE_DIR)/downloads
	
	# 复制配置和模板目录
	@cp -r config $(PACKAGE_DIR)/
	@cp -r templates $(PACKAGE_DIR)/
	@cp -r static $(PACKAGE_DIR)/
	
	# 复制文档和脚本
	@cp -r docs $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)docs directory not found, skipping...$(RESET)"
	@cp -r scripts $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)scripts directory not found, skipping...$(RESET)"
	
	# 复制测试和基准测试目录（用于生产环境验证）
	@cp -r test $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)test directory not found, skipping...$(RESET)"
	@cp -r benchmark $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)benchmark directory not found, skipping...$(RESET)"
	
	# 复制示例配置
	@cp -r example $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)example directory not found, skipping...$(RESET)"
	
	# 复制重要的根目录文件
	@cp Makefile $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)Makefile not found, skipping...$(RESET)"
	@cp README.md $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)README.md not found, skipping...$(RESET)"
	@cp LICENSE $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)LICENSE not found, skipping...$(RESET)"
	@cp Dockerfile $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)Dockerfile not found, skipping...$(RESET)"
	@cp docker-compose.yml $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)docker-compose.yml not found, skipping...$(RESET)"
	@cp docker-compose-swarm.yml $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW)docker-compose-swarm.yml not found, skipping...$(RESET)"
	@cp .dockerignore $(PACKAGE_DIR)/ 2>/dev/null || echo -e "$(YELLOW).dockerignore not found, skipping...$(RESET)"
	@echo -e "$(GREEN)Package created at $(PACKAGE_DIR)$(RESET)"
	@echo -e "$(BLUE)Creating release archive...$(RESET)"
	@cd $(RELEASE_DIR) && tar -czf $(RELEASE_FILE_NAME).tar.gz $(RELEASE_FILE_NAME)
	@echo -e "$(GREEN)Release archive created at $(RELEASE_FILE_NAME).tar.gz$(RESET)"
	@echo -e "$(SEPARATOR)"


# 发布包后的项目启动
local-release-start:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Starting the application from release package...$(RESET)"
	@mkdir -p logs
	@nohup ./$(APP_NAME) -config=$(APP_CONFIG) -env=$(env_file)  > logs/app.log 2>&1 & echo $$! > logs/app.pid
	@echo -e "$(GREEN)Application started from release package. PID: $$(cat logs/app.pid)$(RESET)"
	@echo -e "$(GREEN)Log file: logs/app.log$(RESET)"
	@echo -e "$(BLUE)To view logs: tail -f logs/app.log$(RESET)"
	@echo -e "$(BLUE)To stop: make local-release-stop$(RESET)"
	@echo -e "$(SEPARATOR)"

# 发布包后的项目停止
local-release-stop:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(YELLOW)Stopping the application from release package...$(RESET)"
	@if [ -f "logs/app.pid" ]; then \
		PID=$$(cat logs/app.pid); \
		if kill -0 $$PID 2>/dev/null; then \
			kill $$PID; \
			echo -e "$(GREEN)Application (PID: $$PID) stopped.$(RESET)"; \
		else \
			echo -e "$(YELLOW)Application (PID: $$PID) is not running.$(RESET)"; \
		fi; \
		rm -f logs/app.pid; \
	else \
		pkill -f "$(APP_NAME)" || echo -e "$(RED)No application from release package is running.$(RESET)"; \
	fi
	@echo -e "$(GREEN)Application from release package stopped.$(RESET)"
	@echo -e "$(SEPARATOR)"

# Build the Docker image
_docker-build:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Removing old Docker image if it exists...$(RESET)"
	@docker rmi -f $(DOCKER_IMAGE) || echo -e "$(YELLOW)No existing image to remove.$(RESET)"
	@echo -e "$(BLUE)Building Docker image...$(RESET)"
	@docker build --build-arg WORKDIR=$(WORKDIR) \
		--build-arg APP_CONFIG=$(APP_CONFIG) \
		-t $(DOCKER_IMAGE) . || echo -e "$(RED)Failed to build Docker image.$(RESET)"
	@echo -e "$(GREEN)Docker image built: $(DOCKER_IMAGE)$(RESET)"
	@echo -e "$(SEPARATOR)"

# Run the application in Docker
docker-run: _docker-build
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Creating Docker network if it does not exist...$(RESET)"
	@docker network inspect $(DOCKER_NETWORK) >/dev/null 2>&1 || docker network create $(DOCKER_NETWORK)
	@echo -e "$(BLUE)Checking if Docker log volume exists...$(RESET)"
	@if ! docker volume inspect $(DOCKER_LOG_VOLUME) >/dev/null 2>&1; then \
		echo -e "$(YELLOW)Creating Docker volume for logs...$(RESET)"; \
		docker volume create $(DOCKER_LOG_VOLUME); \
	fi
	@echo -e "$(BLUE)Running the application in Docker...$(RESET)"
	@docker run -d --name $(DOCKER_CONTAINER) --network $(DOCKER_NETWORK) -p ${HOST_PORT}:${CONTAINER_PORT} \
		--env-file $(env_file) \
		--mount type=volume,source=$(DOCKER_LOG_VOLUME),target=$(WORKDIR)/logs \
		--mount type=volume,source=$(DOCKER_DOWNLOAD_VOLUME),target=$(WORKDIR)/downloads \
		$(DOCKER_IMAGE) || echo -e "$(RED)Failed to run the application in Docker.$(RESET)"
	@echo -e "$(SEPARATOR)"

# Stop the Docker container and remove the network and image
docker-stop:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(YELLOW)Stopping the Docker container...$(RESET)"
	@docker stop $(DOCKER_CONTAINER) || echo -e "$(RED)No running container to stop.$(RESET)"
	@docker rm $(DOCKER_CONTAINER) || echo -e "$(RED)No container to remove.$(RESET)"
	@echo -e "$(YELLOW)Removing Docker network...$(RESET)"
	@docker network rm $(DOCKER_NETWORK) || echo -e "$(RED)No network to remove.$(RESET)"
	@echo -e "$(YELLOW)Removing Docker image...$(RESET)"
	@docker rmi $(DOCKER_IMAGE) || echo -e "$(RED)No image to remove.$(RESET)"
	@echo -e "$(GREEN)Docker container, network and image cleaned up.$(RESET)"
	@echo -e "$(SEPARATOR)"


# ---------------------------- 以下所有的命令，都是基于.env.docker-compose 文件 ----------------------------
# docker-compose.yml
docker-compose-up:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Starting Docker Compose...$(RESET)"
	@docker-compose -f docker-compose.yml up -d || echo -e "$(RED)Failed to start Docker Compose.$(RESET)"
	@echo -e "$(GREEN)Docker Compose started.$(RESET)"
	@echo -e "$(SEPARATOR)"

# docker-compose.yml
docker-compose-down:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Stopping Docker Compose...$(RESET)"
	@docker-compose -f docker-compose.yml down || echo -e "$(RED)Failed to stop Docker Compose.$(RESET)"
	@echo -e "$(GREEN)Docker Compose stopped.$(RESET)"
	@echo -e "$(SEPARATOR)"

docker-compose-start:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Starting Docker Compose...$(RESET)"
	@docker-compose -f docker-compose.yml start || echo -e "$(RED)Failed to start Docker Compose.$(RESET)"
	@echo -e "$(GREEN)Docker Compose started.$(RESET)"
	@echo -e "$(SEPARATOR)"

docker-compose-stop:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Stopping Docker Compose...$(RESET)"
	@docker-compose -f docker-compose.yml stop || echo -e "$(RED)Failed to stop Docker Compose.$(RESET)"
	@echo -e "$(GREEN)Docker Compose stopped.$(RESET)"
	@echo -e "$(SEPARATOR)"

# 推送Docker镜像到注册中心
docker-image-push: _docker-build
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Tagging Docker image...$(RESET)"
	@docker tag $(DOCKER_IMAGE) $(REGISTRY_URL)/$(DOCKER_IMAGE)
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Pushing Docker image to registry...$(RESET)"
	@docker push $(REGISTRY_URL)/$(DOCKER_IMAGE) || echo -e "$(RED)Failed to push Docker image.$(RESET)"
	@echo -e "$(GREEN)Docker image pushed to registry.$(RESET)"
	@echo -e "$(SEPARATOR)"

# 初始化swarm集群，并部署. 先docker-image-push
docker-swarm-up: docker-image-push
	@echo -e "$(BLUE)Deploying to Docker Swarm...$(RESET)"
	@docker stack deploy -c docker-compose-swarm.yml $(APP_NAME) || echo -e "$(RED)Failed to deploy to Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)Docker Swarm deployment complete.$(RESET)"
	@echo -e "$(SEPARATOR)"

# 删除整个swarm集群
docker-swarm-down:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Removing stack from Docker Swarm...$(RESET)"
	@docker stack rm $(APP_NAME) || echo -e "$(RED)Failed to remove stack from Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)Stack removed from Docker Swarm.$(RESET)"
	@echo -e "$(SEPARATOR)"


# 删除swarm集群中的app服务
_docker-swarm-rm-app:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Removing app service from Docker Swarm...$(RESET)"
	@docker service rm $(APP_NAME)_app || echo -e "$(RED)Failed to remove app service from Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)App service removed from Docker Swarm.$(RESET)"
	@echo -e "$(SEPARATOR)"


# 删除swarm集群中的nginx服务
_docker-swarm-rm-nginx:
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Removing nginx service from Docker Swarm...$(RESET)"
	@docker service rm $(APP_NAME)_nginx || echo -e "$(RED)Failed to remove nginx service from Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)Nginx service removed from Docker Swarm.$(RESET)"
	@echo -e "$(SEPARATOR)"




# 注意：
# 1. 更新之前需要先docker-image-push
# 2. update app服务，只适用于 1. 副本是vip（虚拟ip）模式，2. docker-compose-swarm.yml 文件并没有修改过, 3. nginx 负载均衡模式不使用ip_hash 三者缺一不可
# 3. update app服务，会导致app服务的ip发生变化，因此如果app服务的上下游是跟IP相关的，用此命令更新app服务，会导致上下游服务不可用，因此慎用
# 4. docker service update 更新服务时，不支持给服务传env-file 所以只能读取环境变量文件，然后构建 --env-add 参数, 否则就是用的原来的环境变量
docker-update-app: docker-image-push
	@echo -e "$(SEPARATOR)"
	@echo -e "$(BLUE)Updating Docker Swarm...$(RESET)"
	@if [ -z "$(env_file)" ]; then \
		echo "env_file is not set! Please provide the path to the environment file."; \
		exit 1; \
	fi
	@if [ ! -f "$(env_file)" ]; then \
		echo "Environment file $(env_file) not found!"; \
		exit 1; \
	fi
	@ENV_VARS=$$(awk -F= '/^[^#]/ && NF==2 {print "--env-add", $$1"="$$2}' $(env_file)); \
	docker service update $$ENV_VARS --image $(REGISTRY_URL)/$(DOCKER_IMAGE) $(APP_NAME)_app || echo -e "$(RED)Failed to update Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)Docker Swarm updated.$(RESET)"
	@echo -e "$(SEPARATOR)"


# 注意：
# 1. 更新之前需要先docker-image-push
# 2. 适用于app或nginx修改了任意配置，都可以使用此命令更新
# 3. 弊端，整个服务会被删掉重建，所以会出现集群不可用，恢复需要时间切记
docker-swarm-deploy-app: docker-image-push _docker-swarm-rm-app _docker-swarm-rm-nginx
	@echo -e "$(BLUE)Deploying to Docker Swarm...$(RESET)"
	@docker stack deploy -c docker-compose-swarm.yml $(APP_NAME) || echo -e "$(RED)Failed to deploy to Docker Swarm.$(RESET)"
	@echo -e "$(GREEN)Docker Swarm deployment complete.$(RESET)"
	@echo -e "$(SEPARATOR)"

