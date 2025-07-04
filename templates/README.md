# Api网关框架使用指南

---

## 一、Wire 使用

**步骤：**

1. 进入 `app` 目录。
2. 运行 `wire` 命令。

```shell
cd app
wire
```

---

## 二、Makefile 使用

### 2.1、本地部署

- **启动项目**：

  ```shell
  make local-run
  ```

- **停止项目**：

  ```shell
  make local-stop
  ```

### 2.2、本地部署（Docker 环境）

- **启动项目**：

  ```shell
  make docker-run
  ```

- **停止项目**：

  ```shell
  make docker-stop
  ```

---

## 三、Docker-Compose 部署

- **初次部署**：

  ```shell
  make  docker-compose-up env_file=.env.docker-compose 
  ```

- **清空容器，镜像，重新打包**：

  ```shell
  make  docker-compose-down env_file=.env.docker-compose  
  ```

- **启动项目**：

  ```shell
  make docker-compose-start env_file=.env.docker-compose  
  ```

- **停止项目**：

  ```shell
   make docker-compose-stop env_file=.env.docker-compose  
  ```

---

## 四、Docker Swarm 集群部署

- **推送镜像**

  ```shell
  make docker-image-push env_file=.env.docker-compose
  ```

- **启动集群**
  ```shell
  make docker-swarm-up env_file=.env.docker-compose
  ```

- **停止集群**
  ```shell
  make docker-swarm-down env_file=.env.docker-compose
  ```

- **更新集群中的app服务**
  ```shell
  make docker-update-app env_file=.env.docker-compose
  ```
  > 注意：
  > - 更新前准备：
  >   1. 确保已经执行 `docker-image-push` 推送最新镜像。
  >   2. 适用条件（缺一不可）：
  >      - 副本是VIP（虚拟IP）模式（通过`docker-compose-swarm.yml`配置）。
  >      - `docker-compose-swarm.yml` 文件未被修改。
  >      - Nginx负载均衡模式不使用`ip_hash`。
  > 
  > - IP变化影响：
  >   - 更新服务会导致应用服务的IP发生变化。
  >   - 如果应用服务的上下游依赖IP连接，更新后可能导致服务不可用，需谨慎使用。
  > 
  > - 环境变量限制：
  >   - `docker service update` 不支持通过`env-file`传递环境变量。
  >   - 需手动读取环境变量文件并构建`--env-add`参数，否则使用原有环境变量。

- **重新部署集群中的app服务**
  ```shell
  make docker-swarm-deploy-app env_file=.env.docker-compose
  ```
  > 注意：
  > 1. 更新之前需要先执行 `docker-image-push`。
  > 2. 适用于app或nginx修改了任意配置的情况。
  > 3. 弊端：整个服务会被删掉重建，可能导致集群暂时不可用，恢复需要时间，需谨慎操作。
  > 4. 不支持 docker service scale service_name=5 扩缩容

- **注意事项**
  > 确保在每个命令中指定正确的`env_file`以加载相应的环境变量。
  > 在更新镜像时，确保新版本的镜像已经推送到注册表中。

---

## 五、配置文件指南

- **config 目录**：用于存储应用内的各种组件的配置。添加新配置后，请在 `config/config.go` 中做好映射。

- **.env.local**：用于本地部署的默认环境变量，解决 Docker 和非 Docker 环境下参数隔离的问题。

- **.env.docker-compose**：Docker-Compose 部署所需的环境变量。

- **docker-compose.yml**: Docker-Compose 单机部署所需的配置文件。

- **docker-compose-swarm.yml**: swan集群部署所需要的配置文件， 注意配置文件中的app镜像地址需要提前push到注册仓库，并且要找对镜像版本哟

---

## 六、初始化和更新项目

- **更新脚本**：项目更新使用的脚本是 `scripts/init.sh`。项目是否更新取决于 `.releaserc` 文件中的项目版本。

- **执行权限**：在执行 `init.sh` 之前，请确保该脚本具有执行权限。可以使用以下命令赋予权限：

  ```shell
  chmod +x scripts/init.sh
  ```

- **运行更新**：执行更新脚本以初始化或更新项目：

  ```shell
  ./scripts/init.sh
  ```

---

## 七、MCP(SSE模式)注意事项

- **Nginx配置**：
  1. 需要使用`ip_hash`模式，以确保同一个客户端的请求被分配到同一个Nginx实例，避免SSE连接失败。

- **App副本路由模式**：
  1. 需要使用`dnsrr`模式，以确保同一个客户端的请求被分配到同一个App实例，避免SSE连接失败。

- **协议支持**：
  1. Nginx需要支持`/see`特殊协议。

- **更新注意**：
  1. 一旦需要更新App，需要重新部署Nginx和App，因为`dnsrr`模式下，如果只更新App，Nginx不会自动更新到新的副本IP上。
  2. 不支持 docker service scale 服务名=5 扩缩容

---

## 八、注意事项

- **优先级**：环境变量中的配置会覆盖 `config` 文件内的配置。

- **自定义配置路径**：可在环境变量文件中修改配置文件目录，例如：

  ```shell
  APP_CONFIG=/your_path/your_config_path
  ```

- **环境变量文件**：建议通过环境变量文件（`env_file`）传入配置，而非命令行参数。例如：

  ```shell
  make local-run env_file=/your_path/your_env_file
  ```

- **集中管理**：建议将配置集中在 `config` 目录和环境变量中进行管理。