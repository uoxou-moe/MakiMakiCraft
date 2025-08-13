#!/bin/bash
echo "Shutdown script executed at $(date)" >> /tmp/shutdown_log.txt
# 実際にはここにサービスの停止やデータのバックアップ処理などを書く
# systemctl stop my-app.service
# aws s3 cp ...
