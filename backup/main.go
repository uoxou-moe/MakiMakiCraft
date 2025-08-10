package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// 設定を読み込む
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	log.Println("Minecraftワールドのバックアッププロセスを開始します。")

	// バックアップファイル名を生成
	currentTime := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("%s_%s.tar.gz", cfg.BackupFileNamePrefix, currentTime)
	fullBackupPath := filepath.Join(cfg.BackupOutputPath, outputFileName)

	log.Printf("ワールドディレクトリ '%s' を '%s' に圧縮中...", cfg.MinecraftWorldDirs, fullBackupPath)

	// 圧縮処理を呼び出す
	if err := createTarGz(cfg.MinecraftWorldDirs, fullBackupPath); err != nil {
		log.Fatalf("ワールドの圧縮に失敗しました: %v", err)
	}

	log.Printf("ワールドの圧縮が完了しました: %s", fullBackupPath)

	// S3へのアップロード
	s3ObjectKey := fmt.Sprintf("minecraft_backups/%s", outputFileName) // S3バケット内のパスとファイル名
	ctx := context.Background()                                    // AWS SDK操作のためのContext

	// S3アップロード処理を呼び出す
	if err := uploadToS3(ctx, fullBackupPath, cfg.S3BucketName, s3ObjectKey, cfg.AWSRegion); err != nil {
		log.Fatalf("S3へのアップロードに失敗しました: %v", err)
	}

	// S3へのアップロードが成功したらローカルファイルを削除
	log.Printf("S3へのアップロードが成功したため、ローカルファイル '%s' を削除します。", fullBackupPath)
	if err := os.Remove(fullBackupPath); err != nil {
		log.Printf("ローカルファイル '%s' の削除に失敗しました: %v", fullBackupPath, err)
	} else {
		log.Println("ローカルファイルが削除されました。")
	}

	log.Println("Minecraftのバックアッププロセスが完了しました。")
}
