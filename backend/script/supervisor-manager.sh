#!/bin/bash

# Supervisor 管理脚本
# 用法: ./supervisor-manager.sh [start|stop|restart|status|logs]

SERVICE_NAME="golang-gin-demo"

case "$1" in
    start)
        echo "启动 $SERVICE_NAME 服务..."
        sudo supervisorctl start $SERVICE_NAME
        ;;
    stop)
        echo "停止 $SERVICE_NAME 服务..."
        sudo supervisorctl stop $SERVICE_NAME
        ;;
    restart)
        echo "重启 $SERVICE_NAME 服务..."
        sudo supervisorctl restart $SERVICE_NAME
        ;;
    status)
        echo "查看 $SERVICE_NAME 服务状态..."
        sudo supervisorctl status $SERVICE_NAME
        ;;
    logs)
        echo "查看 $SERVICE_NAME 服务日志..."
        sudo tail -f /var/log/golang-gin-demo.log
        ;;
    update)
        echo "更新 Supervisor 配置..."
        sudo supervisorctl reread
        sudo supervisorctl update
        ;;
    *)
        echo "用法: $0 {start|stop|restart|status|logs|update}"
        echo ""
        echo "命令说明:"
        echo "  start   - 启动服务"
        echo "  stop    - 停止服务"
        echo "  restart - 重启服务"
        echo "  status  - 查看服务状态"
        echo "  logs    - 查看服务日志"
        echo "  update  - 更新配置"
        exit 1
        ;;
esac
