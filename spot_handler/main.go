package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PollingInterval string `yaml:"pollingInterval"`
	ShutdownScript  string `yaml:"shutdownScript"`
	MetadataURL     string `yaml:"metadataUrl"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config yaml: %w", err)
	}

	return &config, nil
}

// --- 中断検知関数 ---
// trueを返した場合、中断通知があったことを示す
func checkInterruption(url string) (bool, error) {
	// タイムアウトを設定したHTTPクライアントを作成
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		// ネットワークエラーなど
		return false, fmt.Errorf("failed to get metadata: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// 200 OK: 中断通知あり
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Interruption notice received. Details: %s", string(body))
		return true, nil
	case http.StatusNotFound:
		// 404 Not Found: 正常、中断なし
		log.Println("No interruption notice. Continuing to poll.")
		return false, nil
	default:
		// その他のステータスコードは予期せぬエラー
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// --- シャットダウン処理実行関数 ---
func executeShutdownScript(scriptPath string) {
	log.Printf("Executing shutdown script: %s", scriptPath)
	// コマンドを実行
	cmd := exec.Command("/bin/sh", scriptPath)

	// コマンドの標準出力と標準エラーを取得
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing shutdown script: %v. Output: %s", err, string(output))
		return
	}
	log.Printf("Shutdown script executed successfully. Output: %s", string(output))
}

func main() {
	log.Println("Starting Spot Interruption Handler...")

	// 1. 設定ファイルを読み込む
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Fatal: Could not load config. %v", err)
	}
	log.Printf("Config loaded: Polling every %s", config.PollingInterval)

	// 2. ポーリング間隔をパース
	duration, err := time.ParseDuration(config.PollingInterval)
	if err != nil {
		log.Fatalf("Fatal: Invalid polling interval format. %v", err)
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	// 3. OSのシグナルを待機するチャンネルを作成
	// SIGINT (Ctrl+C) や SIGTERM (killコマンド) を受け取ったら終了する
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 4. メインループ
	for {
		select {
		case <-ticker.C:
			// 定期的なポーリング処理
			interrupted, err := checkInterruption(config.MetadataURL)
			if err != nil {
				// エラーが発生しても処理は継続する
				log.Printf("Error checking for interruption: %v", err)
				continue
			}

			if interrupted {
				// 中断を検知したらスクリプトを実行して終了
				executeShutdownScript(config.ShutdownScript)
				log.Println("Handler finished its job. Exiting.")
				return
			}

		case sig := <-sigChan:
			// OSからの終了シグナルを受け取った場合
			log.Printf("Received signal: %s. Shutting down.", sig)
			return
		}
	}
}
