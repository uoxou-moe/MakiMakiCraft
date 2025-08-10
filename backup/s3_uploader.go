package main

import (
	"context" // contextパッケージはAWS SDKの操作で必要
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws" // SDKの型を扱うために必要
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// uploadToS3 は指定されたローカルファイルをS3にアップロードします。
func uploadToS3(ctx context.Context, filePath, bucketName, objectKey, region string) error {
    log.Printf("S3にアップロード中: '%s' から s3://%s/%s (リージョン: %s)", filePath, bucketName, objectKey, region)

    // AWS SDK設定をロード (IAMロール、環境変数などを自動的に検出)
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return fmt.Errorf("AWS設定のロードに失敗しました: %w", err)
    }
    // S3クライアントとアップロードマネージャーを初期化
    s3Client := s3.NewFromConfig(cfg)
    uploader := manager.NewUploader(s3Client)

    // アップロードするローカルファイルを開く
    f, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("ローカルファイル '%s' を開くことができませんでした: %w", filePath, err)
    }
    defer f.Close() // 関数終了時にファイルを確実にクローズ

    // S3へのアップロードを実行
    _, err = uploader.Upload(ctx, &s3.PutObjectInput{
        Bucket: aws.String(bucketName), // S3バケット名
        Key:    aws.String(objectKey),  // S3オブジェクトキー (バケット内のパス+ファイル名)
        Body:   f,                      // ファイルの内容をio.Readerとして渡す
    })
    if err != nil {
        return fmt.Errorf("S3へのアップロードに失敗しました: %w", err)
    }

    log.Printf("S3へのアップロードが完了しました: s3://%s/%s", bucketName, objectKey)
    return nil
}
