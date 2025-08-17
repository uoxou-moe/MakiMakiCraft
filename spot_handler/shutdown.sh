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

# --- Discord通知送信 ---
send_discord_notification() {
    # DISCORD_WEBHOOK_URLが設定されていない場合は何もしない
    if [ -z "${DISCORD_WEBHOOK_URL:-}" ]; then
        log "INFO: Discord webhook URL not set. Skipping notification."
        return 0
    fi

    log "INFO: Sending notification to Discord..."

    # 引数からメッセージと色を取得
    local message="$1"
    local color="$2" # Discordの色コード (10進数)
    
    # EC2インスタンスIDを取得
    local instance_id
    instance_id=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
    if [ -z "$instance_id" ]; then
        instance_id="N/A"
    fi

    # Discordに送信するJSONペイロードを heredoc で作成
    local json_payload
    json_payload=$(cat <<EOF
{
    "username": "Minecraft Server Bot",
    "avatar_url": "https://i.imgur.com/v1hGfV8.png",
    "embeds": [{
        "title": "Spot Instance Shutdown Notice",
        "description": "$message",
        "color": "$color",
        "fields": [
        {
            "name": "Instance ID",
            "value": "$instance_id",
            "inline": true
        },
        {
            "name": "Timestamp (UTC)",
            "value": "$(date -u +"%Y-%m-%d %H:%M:%S")",
            "inline": true
        }
        ],
        "footer": {
        "text": "Automated notification from spot-handler"
        }
    }]
}
EOF
)

    # curlを使ってWebhookにPOSTリクエストを送信
    # --fail: HTTPエラー時にエラーコードを返す
    # -s: プログレスバーを非表示
    # -o /dev/null: レスポンスボディを捨てる
    if curl --fail -s -o /dev/null -H "Content-Type: application/json" -d "$json_payload" "$DISCORD_WEBHOOK_URL"; then
        log "SUCCESS: Notification sent to Discord."
    else
        log "ERROR: Failed to send notification to Discord."
    fi
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

# mcrconの実行とDiscord通知
# "$VAR" のように変数をダブルクォートで囲むことで、予期せぬ挙動を防ぐ
if "$MCRCON_PATH" -H "$RCON_HOST" -P "$RCON_PORT" -p "$RCON_PASSWORD" "stop"; then
    log "SUCCESS: 'stop' command sent successfully to the Minecraft server."

    # ★★★ 通知を送信 (成功ケース) ★★★
    send_discord_notification "✅ **Shutdown initiated gracefully!**\nServer stop command sent via RCON." "3066993"

    log "The minecraft.service will now handle the final shutdown and world save."
else
    log "ERROR: Failed to send 'stop' command via RCON. Check RCON settings and password."

    # ★★★ 通知を送信 (失敗/フォールバックケース) ★★★
    send_discord_notification "⚠️ **RCON command failed!**\nAttempting fallback with systemctl stop." "15105570"

    log "FALLBACK: Attempting to stop the service directly via systemctl."
    sudo systemctl stop "$MINECRAFT_SERVICE"
    exit 1
fi

log "====== Shutdown script finished successfully ======"

exit 0
