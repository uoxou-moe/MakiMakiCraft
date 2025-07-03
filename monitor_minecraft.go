package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec" // for 'aws s3 sync' if needed
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/HimbeerserverDE/rcon"
)

// 設定用の変数（デフォルト値）
const (
	defaultRCONHost          = "localhost"
	defaultRCONPort          = 25575
	defaultStopThresholdMinutes = 15 // X分間プレイヤーがいない場合に停止
	counterFile              = "/tmp/minecraft_zero_players_counter"
	lastStopSentFile         = "/tmp/minecraft_last_stop_sent_timestamp"
)

var (
	rconHost          string
	rconPort          int
	rconPassword      string
	sqsQueueURL       string
	stopThresholdMinutes int
)

func init() {
	rconHost = getEnvOrDefault("RCON_HOST", defaultRCONHost)
	rconPortStr := getEnvOrDefault("RCON_PORT", strconv.Itoa(defaultRCONPort))
	rconPort, _ = strconv.Atoi(rconPortStr)
	if rconPort == 0 { rconPort = defaultRCONPort }

	rconPassword = os.Getenv("RCON_PASSWORD")
	sqsQueueURL = os.Getenv("SQS_QUEUE_URL")

	stopThresholdStr := getEnvOrDefault("STOP_THRESHOLD_MINUTES", strconv.Itoa(defaultStopThresholdMinutes))
	stopThresholdMinutes, _ = strconv.Atoi(stopThresholdStr)
	if stopThresholdMinutes == 0 { stopThresholdMinutes = defaultStopThresholdMinutes } // 閾値変換失敗時
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getPlayerCount(ctx context.Context) (int, error) {
	if rconPassword == "" {
		return -1, fmt.Errorf("RCON_PASSWORD environment variable is not set")
	}

	// 接続タイムアウトを設定
	conn, err := rcon.DialWithTimeout(fmt.Sprintf("%s:%d", rconHost, rconPort), time.Second*5)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial timeout") {
			return -2, fmt.Errorf("RCON connection failed: %w (server might be down or not responding)", err)
		}
		return -3, fmt.Errorf("failed to dial RCON: %w", err)
	}
	defer conn.Close()

	if err := conn.Authenticate(rconPassword); err != nil {
		return -4, fmt.Errorf("RCON authentication failed: %w", err)
	}

	response, err := conn.Execute("list")
	if err != nil {
		return -5, fmt.Errorf("failed to execute 'list' command: %w", err)
	}
	fmt.Printf("RCON Response: %s\n", response)

	// レスポンス例: "There are 2/20 players online: Player1, Player2"
	if strings.Contains(response, "There are") && strings.Contains(response, "players online:") {
		parts := strings.Split(response, " ")
		if len(parts) > 2 {
			playerCountStr := strings.Split(parts[2], "/")[0]
			count, err := strconv.Atoi(playerCountStr)
			if err != nil {
				return 0, fmt.Errorf("failed to parse player count from '%s': %w", playerCountStr, err)
			}
			return count, nil
		}
	}
	fmt.Println("Unexpected RCON 'list' response format. Assuming 0 players.")
	return 0, nil // RCON接続は成功したが、プレイヤーリストが取得できない場合は0人とみなす
}

// float64 の値を読み込む
func readFileContent(filePath string) (float64, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0.0, nil // ファイルが存在しない場合は0を返す
		}
		return 0.0, fmt.Errorf("error reading file %s: %w", filePath, err)
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(content)), 64)
	if err != nil {
		return 0.0, fmt.Errorf("error parsing file content '%s' from %s: %w", strings.TrimSpace(string(content)), filePath, err)
	}
	return val, nil
}

// float64 の値を書き込む
func writeFileContent(filePath string, content float64) error {
	return ioutil.WriteFile(filePath, []byte(fmt.Sprintf("%.0f", content)), 0644) // 整数として保存
}

func getInstanceID() (string, error) {
	resp, err := http.Get("http://169.254.169.254/latest/meta-data/instance-id")
	if err != nil {
		return "", fmt.Errorf("failed to get instance ID from metadata: %w", err)
	}
	defer resp.Body.Close()
	id, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read instance ID from metadata: %w", err)
	}
	return strings.TrimSpace(string(id)), nil
}

func sendSQSMessage(ctx context.Context, instanceID, queueURL string) error {
	if queueURL == "" {
		return fmt.Errorf("SQS_QUEUE_URL environment variable is not set")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS SDK config: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	messageBody := map[string]string{
		"instanceId": instanceID,
		"action":     "stop_minecraft_server",
	}
	messageBodyBytes, _ := json.Marshal(messageBody)

	_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: aws.String(string(messageBodyBytes)),
	})
	if err != nil {
		return fmt.Errorf("failed to send SQS message: %w", err)
	}
	fmt.Printf("SQS message sent: %s\n", string(messageBodyBytes))
	return nil
}

func minecraftServerStopCommand(ctx context.Context) error {
	if rconPassword == "" {
		return fmt.Errorf("RCON_PASSWORD environment variable is not set")
	}

	conn, err := rcon.DialWithTimeout(fmt.Sprintf("%s:%d", rconHost, rconPort), time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to connect to RCON for stop command: %w", err)
	}
	defer conn.Close()

	if err := conn.Authenticate(rconPassword); err != nil {
		return fmt.Errorf("RCON authentication failed for stop command: %w", err)
	}

	response, err := conn.Execute("stop")
	if err != nil {
		return fmt.Errorf("failed to execute 'stop' command: %w", err)
	}
	fmt.Printf("Minecraft server 'stop' command response: %s\n", response)
	return nil
}

// s3BackupWorld はMinecraftワールドをS3にバックアップします。
func s3BackupWorld(worldPath, s3BucketPath string) error {
	fmt.Printf("Starting S3 backup of %s to %s...\n", worldPath, s3BucketPath)
	cmd := exec.Command("aws", "s3", "sync", worldPath, s3BucketPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("S3 backup failed: %w", err)
	}
	fmt.Println("S3 backup completed.")
	return nil
}


func main() {
	ctx := context.Background()

	instanceID, err := getInstanceID()
	if err != nil {
		fmt.Printf("Error getting instance ID: %v. Exiting.\n", err)
		return
	}

	if rconPassword == "" {
		fmt.Println("Error: RCON_PASSWORD environment variable is not set. Exiting.")
		return
	}
	if sqsQueueURL == "" {
		fmt.Println("Error: SQS_QUEUE_URL environment variable is not set. Exiting.")
		return
	}

	playerCount, err := getPlayerCount(ctx)
	if err != nil {
		fmt.Printf("Error getting player count: %v. Treating as 0 for counter.\n", err)
		// RCON接続エラーの場合、プレイヤーはいないと判断し0とする
		playerCount = 0
	}

	currentZeroCount, err := readFileContent(counterFile)
	if err != nil {
		fmt.Printf("Error reading counter file: %v. Starting from 0.\n", err)
		currentZeroCount = 0
	}
	lastStopSentTS, err := readFileContent(lastStopSentFile)
	if err != nil {
		fmt.Printf("Error reading last stop sent timestamp: %v. Starting from 0.\n", err)
		lastStopSentTS = 0.0
	}
	currentTime := float64(time.Now().Unix())

	if playerCount == 0 {
		currentZeroCount++
		fmt.Printf("Player count is 0. Consecutive zero count: %.0f\n", currentZeroCount)
	} else {
		currentZeroCount = 0
		fmt.Printf("Player count: %d. Resetting zero count.\n", playerCount)
	}
	writeFileContent(counterFile, currentZeroCount)

	// 停止閾値を超え、かつ最近停止リクエストを送信していないかチェック
	// SQSメッセージ送信から停止までタイムラグがあるため、多重送信を防ぐ
	if currentZeroCount >= float64(stopThresholdMinutes) && (currentTime-lastStopSentTS) > float64(stopThresholdMinutes*60) {
		fmt.Printf("Zero players for %.0f minutes. Sending stop request...\n", currentZeroCount)
		
		// オプション: Minecraftサーバーに安全なシャットダウンコマンドを送信
		// これにより、EC2が停止する前にゲームが安全に終了します。
		if err := minecraftServerStopCommand(ctx); err != nil {
			fmt.Printf("Warning: Failed to send Minecraft server stop command: %v\n", err)
		}

		// オプション: ワールドデータのS3バックアップ
		// 例: s3BackupWorld("/path/to/minecraft/server/world", "s3://your-minecraft-backup-bucket/world_data/")
		// バックアップが完了するまでブロックするため、時間がかかる場合は注意

		if err := sendSQSMessage(ctx, instanceID, sqsQueueURL); err != nil {
			fmt.Printf("Failed to send stop request to SQS: %v. Will retry on next check.\n", err)
		} else {
			writeFileContent(lastStopSentFile, currentTime) // SQSメッセージ送信時刻を記録
			fmt.Println("Stop request sent. EC2 should stop soon.")
		}
	} else if currentZeroCount >= float64(stopThresholdMinutes) {
		fmt.Println("Stop request already sent recently. Waiting for EC2 to stop.")
	}
}
