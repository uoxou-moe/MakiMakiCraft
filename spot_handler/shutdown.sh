#!/bin/bash

set -euo pipefail

# --- 設定ファイルの読み込み ---
# スクリプトと同じディレクトリにある .conf ファイルを読み込む
# 別の場所にある場合はフルパスを指定してください (例: /etc/spot-handler/shutdown.conf)
CONFIG_FILE="$(dirname "$0")/shutdown.conf"

if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "FATAL: Configuration file not found at $CONFIG_FILE" >&2
    exit 1
fi


# --- ロギング関数 ---
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | sudo tee -a "$SHUTDOWN_LOG_FILE"
}


# --- メイン処理 ---
log "====== Minecraft Spot Shutdown Script Started ======"

# 1. 前提条件のチェック (変数は .conf ファイルから読み込まれている)
if [ -z "$RCON_PASSWORD_FILE" ] || [ ! -f "$RCON_PASSWORD_FILE" ]; then
    log "ERROR: RCON_PASSWORD_FILE is not set or file not found at '$RCON_PASSWORD_FILE'. Please check shutdown.conf. Exiting."
    exit 1
fi
if [ ! -x "$MCRCON_PATH" ]; then
    log "ERROR: mcrcon not found or is not executable at '$MCRCON_PATH'. Exiting."
    exit 1
fi

# 2. Minecraftサーバーが起動しているか確認
if ! systemctl is-active --quiet "$MINECRAFT_SERVICE"; then
    log "INFO: Minecraft service ('$MINECRAFT_SERVICE') is not running. Nothing to do. Exiting."
    exit 0
fi

# 3. RCONで 'stop' コマンドを送信
log "INFO: Minecraft server is running. Attempting to send 'stop' command via RCON..."

RCON_PASSWORD=$(cat "$RCON_PASSWORD_FILE")

# mcrconの実行
# "$VAR" のように変数をダブルクォートで囲むことで、予期せぬ挙動を防ぐ
if "$MCRCON_PATH" -H "$RCON_HOST" -P "$RCON_PORT" -p "$RCON_PASSWORD" "stop"; then
    log "SUCCESS: 'stop' command sent successfully to the Minecraft server."
    log "The minecraft.service will now handle the final shutdown and world save."
else
    log "ERROR: Failed to send 'stop' command via RCON. Check RCON settings and password."
    log "FALLBACK: Attempting to stop the service directly via systemctl."
    sudo systemctl stop "$MINECRAFT_SERVICE"
    exit 1
fi

log "====== Shutdown script finished successfully ======"

exit 0
