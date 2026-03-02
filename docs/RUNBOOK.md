# 运维手册

本文档提供 Peanut 应用的部署、监控和故障排除指南。

## 部署

### 本地部署

```bash
# 构建
make build

# 配置环境变量
export ARK_API_KEY="your-api-key"

# 运行
./bin/peanut
```

### Docker 部署

```bash
# 构建镜像
make docker-build

# 运行容器
make docker-run

# 或使用 Docker Compose
make docker-up

# 查看日志
make docker-logs

# 停止服务
make docker-down
```

### 生产环境部署

```bash
# 1. 构建生产镜像
docker build -t peanut:prod .

# 2. 运行容器
docker run -d \
  --name peanut \
  -p 8080:8080 \
  -e ARK_API_KEY="your-api-key" \
  -e GIN_MODE=release \
  -v peanut-data:/app/data \
  peanut:prod

# 3. 验证部署
curl http://localhost:8080/health
```

## 监控

### 健康检查

| 端点 | 说明 | 预期响应 |
|------|------|----------|
| `GET /health` | 服务健康状态 | `{"status": "healthy"}` |
| `GET /ready` | 服务就绪状态 | `{"status": "ready"}` |

### 日志

日志输出到标准输出，支持 JSON 格式（生产环境）和 console 格式（开发环境）。

```bash
# 查看日志
docker logs -f peanut

# 或
make docker-logs
```

### 关键指标

- **请求延迟**：HTTP 请求处理时间
- **错误率**：5xx 响应比例
- **GEO 分析完成率**：completed / total
- **LLM API 调用成功率**：豆包 API 调用统计

## 常见问题

### 1. 服务无法启动

**症状**：服务启动失败

**排查步骤**：

```bash
# 检查端口占用
lsof -i :8080

# 检查环境变量
echo $ARK_API_KEY

# 查看详细日志
./bin/peanut 2>&1 | tee app.log
```

**解决方案**：
- 确保端口 8080 未被占用
- 检查 ARK_API_KEY 是否正确配置

### 2. GEO 分析失败

**症状**：分析任务状态为 `failed`

**排查步骤**：

```bash
# 查看任务详情
curl http://localhost:8080/api/v1/geo/analysis/{id}

# 检查 error_message 字段
```

**常见原因**：
- ARK_API_KEY 无效或过期
- 网络连接问题（无法访问豆包 API）
- 目标 URL 无法访问

**解决方案**：
```bash
# 验证 API Key
curl -X POST https://ark.cn-beijing.volces.com/api/v3/chat/completions \
  -H "Authorization: Bearer $ARK_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"doubao-pro-256k-240628","messages":[{"role":"user","content":"test"}]}'
```

### 3. 数据库问题

**症状**：数据库相关错误

**排查步骤**：

```bash
# 检查数据库文件
ls -la peanut.db

# 检查文件权限
chmod 644 peanut.db
```

**解决方案**：
- 删除数据库重新创建：`rm peanut.db && make run`
- 检查磁盘空间

### 4. 内存不足

**症状**：服务崩溃或响应缓慢

**排查步骤**：

```bash
# 检查内存使用
docker stats peanut

# 或本地
ps aux | grep peanut
```

**解决方案**：
- 增加 Docker 内存限制
- 减少并发请求数

## 回滚

### Docker 回滚

```bash
# 1. 停止当前服务
docker stop peanut

# 2. 回滚到上一版本
docker run -d \
  --name peanut-rollback \
  -p 8080:8080 \
  -e ARK_API_KEY="your-api-key" \
  peanut:previous-version

# 3. 验证
curl http://localhost:8080/health

# 4. 清理旧容器
docker rm peanut
docker rename peanut-rollback peanut
```

### 本地回滚

```bash
# 1. 停止服务
pkill -f peanut

# 2. 回滚代码
git checkout HEAD~1

# 3. 重新构建
make build

# 4. 重启服务
./bin/peanut
```

## 备份与恢复

### 数据库备份

```bash
# 备份 SQLite 数据库
cp peanut.db peanut.db.backup.$(date +%Y%m%d_%H%M%S)

# 或使用 Docker volume 备份
docker run --rm -v peanut-data:/data -v $(pwd):/backup alpine \
  cp /data/peanut.db /backup/peanut.db.backup
```

### 数据库恢复

```bash
# 恢复数据库
cp peanut.db.backup.YYYYMMDD_HHMMSS peanut.db

# 或从 Docker volume 恢复
docker run --rm -v peanut-data:/data -v $(pwd):/backup alpine \
  cp /backup/peanut.db.backup /data/peanut.db
```

## 紧急联系人

- **项目负责人**：solariswu
- **GitHub Issues**：https://github.com/solariswu/peanut/issues
