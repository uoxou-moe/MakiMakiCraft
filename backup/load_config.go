package main

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	MinecraftWorldDir    string
	BackupOutputPath     string
	BackupFileNamePrefix string
	S3BucketName         string
	AWSRegion            string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		MinecraftWorldDir:    os.Getenv("MINECRAFT_WORLD_DIR"),
		BackupOutputPath:     os.Getenv("BACKUP_OUTPUT_PATH"),
		BackupFileNamePrefix: os.Getenv("BACKUP_FILE_NAME_PREFIX"),
		S3BucketName:         os.Getenv("S3_BUCKET_NAME"),
		AWSRegion:            os.Getenv("AWS_REGION"),
	}

	if cfg.MinecraftWorldDir == "" {
		return nil, fmt.Errorf("環境変数 MINECRAFT_WORLD_DIR が設定されていません。(例: /path/to/your/minecraft_server/world)")
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
