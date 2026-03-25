#!/bin/bash

# SChill 服务启动脚本 (Linux/Mac)
# 用于依次启动所有 RPC 和 API 服务

echo "========================================"
echo "  SChill 服务启动脚本"
echo "========================================"
echo ""

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
echo "项目根目录: $PROJECT_ROOT"
echo ""

# 定义服务列表
SERVICES=(
    "user-rpc|service/user/rpc|go run user.go"
    "content-rpc|service/content/rpc|go run content.go"
    "relation-rpc|service/relation/rpc|go run relation.go"
    "user-api|service/user/api|go run user.go"
    "content-api|service/content/api|go run content.go"
    "relation-api|service/relation/api|go run relation.go"
    "comment-api|service/comment/api|go run comment.go"
)

# 存储所有 PID
PIDS=()

# 清理函数
cleanup() {
    echo ""
    echo "正在停止所有服务..."
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null
            echo "已停止进程 PID: $pid"
        fi
    done
    echo "所有服务已停止"
    exit 0
}

# 捕获中断信号
trap cleanup SIGINT SIGTERM

# 启动服务函数
start_service() {
    local service_info="$1"
    IFS='|' read -r name path cmd <<< "$service_info"
    
    local service_path="$PROJECT_ROOT/$path"
    
    if [ ! -d "$service_path" ]; then
        echo "[错误] 服务目录不存在: $service_path"
        return 1
    fi
    
    echo "[启动] $name..."
    
    cd "$service_path" || return 1
    
    # 在后台启动服务
    eval "$cmd" &
    local pid=$!
    PIDS+=("$pid")
    
    echo "[成功] $name 已启动 (PID: $pid)"
    sleep 2
    
    cd "$PROJECT_ROOT" || return 1
    return 0
}

echo "开始启动服务..."
echo ""

success_count=0
for service in "${SERVICES[@]}"; do
    if start_service "$service"; then
        ((success_count++))
    fi
    echo ""
done

echo "========================================"
echo "启动完成!"
echo "成功启动: $success_count / ${#SERVICES[@]} 个服务"
echo ""
echo "按 Ctrl+C 停止所有服务"
echo "========================================"
echo ""

# 等待所有子进程
wait
