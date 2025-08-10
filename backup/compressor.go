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


func createTarGz(sourceDirs []string, outputFile string) error {
	log.Printf("'%s' を tar.gz 形式で '%s' に圧縮します...", sourceDirs, outputFile)

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

	for _, sourceDir := range sourceDirs {
		srcInfo , err := os.Stat(sourceDir)
		if os.IsNotExist(err) {
			return fmt.Errorf("ソースディレクトリが存在しません: %s", sourceDir)
		}
		if !srcInfo.IsDir() {
			return fmt.Errorf("ソース '%s' はディレクトリではありません", sourceDir)
		}
		// 各ディレクトリ内を再帰的に走査し、tarアーカイブに追加
		err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// ベースパスからの相対パスを作成
			relPath, err := filepath.Rel(sourceDir, path)
			if err != nil {
				return fmt.Errorf("相対パスの取得に失敗しました (%s, %s): %w", sourceDir, path, err)
			}

			// tarアーカイブ内のパスを生成 (例: world/region/r.0.0.mca)
			// ソースディレクトリのベース名 (例: 'world') と相対パスを結合する
			tarPath := filepath.Join(filepath.Base(sourceDir), relPath)

			// tarヘッダを作成
			header, err := tar.FileInfoHeader(info, tarPath)
			if err != nil {
				return fmt.Errorf("tarヘッダの作成に失敗しました (%s): %w", path, err)
			}
			header.Name = filepath.ToSlash(tarPath) // tarアーカイブはUnix形式のパス区切り文字 (/) を期待

			// ヘッダをtarライターに書き込み
			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("tarヘッダの書き込みに失敗しました (%s): %w", header.Name, err)
			}

			// ディレクトリの場合はデータは不要なのでスキップ
			if info.IsDir() {
				return nil
			}

			// ファイルの場合、内容をコピー
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
			return fmt.Errorf("ディレクトリ '%s' のtar.gz作成中にエラーが発生しました: %w", sourceDir, err)
		}
	}

	log.Printf("すべてのディレクトリの圧縮が完了しました。")
	return nil
}
