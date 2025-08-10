package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	MinecraftWorldDirs   []string
	BackupOutputPath     string
	BackupFileNamePrefix string
	S3BucketName         string
	AWSRegion            string
}

func LoadConfig() (*Config, error) {
	worldDirsStr := os.Getenv("MINECRAFT_WORLD_DIRS")
	if worldDirsStr == "" {
		return nil, fmt.Errorf("環境変数 MINECRAFT_WORLD_DIRS が設定されていません。(例: /path/to/world,/path/to/world_nether)")
	}

	// カンマで分割し、各パスの前後の空白を削除
	var worldDirs []string
	for _, p := range strings.Split(worldDirsStr, ",") {
		trimmedPath := strings.TrimSpace(p)
		if trimmedPath != "" {
			worldDirs = append(worldDirs, trimmedPath)
		}
	}
	if len(worldDirs) == 0 {
		return nil, fmt.Errorf("環境変数 MINECRAFT_WORLD_DIRS に有効なパスが指定されていません。")
	}

	cfg := &Config{
		MinecraftWorldDirs:   worldDirs,
		BackupOutputPath:     os.Getenv("BACKUP_OUTPUT_PATH"),
		BackupFileNamePrefix: os.Getenv("BACKUP_FILE_NAME_PREFIX"),
		S3BucketName:         os.Getenv("S3_BUCKET_NAME"),
		AWSRegion:            os.Getenv("AWS_REGION"),
	}

	if cfg.BackupOutputPath == "" {
		return nil, fmt.Errorf("環境変数 BACKUP_OUTPUT_PATH が設定されていません。(例: /path/to/your/backups)")
	}
	if cfg.BackupFileNamePrefix == "" {
		log.Println("警告: 環境変数 BACKUP_FILE_NAME_PREFIX が設定されていません。デフォルト値 'minecraft_world' を使用します。")
		cfg.BackupFileNamePrefix = "minecraft_world"
	}
	if cfg.S3BucketName == "" {
		return nil, fmt.Errorf("環境変数 S3_BUCKET_NAME が設定されていません。")
	}
	if cfg.AWSRegion == "" {
		return nil, fmt.Errorf("環境変数 AWS_REGION が設定されていません。")
	}

	return cfg, nil
}
