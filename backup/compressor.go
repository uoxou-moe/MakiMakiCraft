package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)


func createTarGz(sourceDir, outputFile string) error {
	log.Printf("'%s' を tar.gz 形式で '%s' に圧縮します...", sourceDir, outputFile)

	// 出力ファイルのディレクトリが存在することを確認し、なければ作成
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("出力ディレクトリ '%s' の作成に失敗しました: %w", outputDir, err)
	}

	// 出力ファイルを作成
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("出力ファイル '%s' の作成に失敗しました: %w", outputFile, err)
	}
	defer file.Close()

	// gzipライターを作成
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// tarライターを作成
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// ソースディレクトリが存在し、ディレクトリであることを確認
	srcInfo, err := os.Stat(sourceDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("ソースディレクトリが存在しません: %s", sourceDir)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("ソース '%s' はディレクトリではありません", sourceDir)
	}

	// ディレクトリ内を再帰的に走査し、tarアーカイブに追加
	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("相対パスの取得に失敗しました (%s, %s): %w", sourceDir, path, err)
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return fmt.Errorf("tarヘッダの作成に失敗しました (%s): %w", path, err)
		}
		header.Name = filepath.ToSlash(relPath) // tarアーカイブはUnix形式のパス区切り文字 (/) を期待

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("tarヘッダの書き込みに失敗しました (%s): %w", header.Name, err)
		}

		if info.IsDir() {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("ファイルを開くことができませんでした (%s): %w", path, err)
		}
		defer srcFile.Close()

		if _, err := io.Copy(tw, srcFile); err != nil {
			return fmt.Errorf("ファイルのコピーに失敗しました (%s): %w", path, err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("tar.gz作成中にエラーが発生しました: %w", err)
	}

	log.Printf("'%s' の圧縮が完了しました。", sourceDir)
	return nil
}
