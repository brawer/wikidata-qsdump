// SPDX-FileCopyrightText: 2023 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var logger *log.Logger

func main() {
	ctx := context.Background()

	var dumps = flag.String("dumps", "/public/dumps/public", "path to Wikimedia dumps")
	var testRun = flag.Bool("testRun", false, "if true, we process only a small fraction of the data; used for testing")
	storagekey := flag.String("storage-key", "", "path to key with storage access credentials")
	flag.Parse()

	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatal(err)
	}
	logfile, err := os.OpenFile("logs/qsdump.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	var storage *minio.Client
	if *storagekey != "" {
		storage, err = NewStorageClient(*storagekey)
		if err != nil {
			logger.Fatal(err)
		}

		bucketExists, err := storage.BucketExists(ctx, "qsdump")
		if err != nil {
			logger.Fatal(err)
		}
		if !bucketExists {
			logger.Fatal("storage bucket \"qsdump\" does not exist")
		}
	}

	if err := buildDump(*dumps, *testRun, storage); err != nil {
		logger.Printf("qsdump failed: %v", err)
		log.Fatal(err)
		return
	}
}

// NewStorageClient sets up a client for accessing S3-compatible storage.
func NewStorageClient(keypath string) (*minio.Client, error) {
	data, err := os.ReadFile(keypath)
	if err != nil {
		return nil, err
	}

	var config struct{ Endpoint, Key, Secret string }
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Key, config.Secret, ""),
		Secure: true,
	})
	if err != nil {
		return nil, err
	}

	client.SetAppInfo("WikidataQSDump", "0.1")
	return client, nil
}

func buildDump(dumpsPath string, testRun bool, storage *minio.Client) error {
	edate, epath, err := findEntitiesDump(dumpsPath)
	if err != nil {
		return err
	}

	qsdumpPath := fmt.Sprintf("qsdump-%s.zst", edate.Format("20060102"))
	_ = os.Remove(qsdumpPath)
	cmd := exec.Command("zstd", "-q", "-11", "-o", qsdumpPath)
	writer, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer writer.Close()

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := extractQuickStatements(epath, writer); err != nil {
			panic(err)
		}
		if err := writer.Close(); err != nil {
			panic(err)
		}
	}()
	if err := cmd.Wait(); err != nil {
		return err
	}

	if storage != nil {
		if err := upload(edate, qsdumpPath, storage); err != nil {
			return err
		}
	}

	return nil
}

// Upload puts the final output files into an S3-compatible object storage.
func upload(date time.Time, localPath string, storage *minio.Client) error {
	ymd := date.Format("20060102")
	dest := fmt.Sprintf("public/qsdump-%s.zst", ymd)
	if err := uploadFile(dest, localPath, "application/zstd", storage); err != nil {
		return err
	}

	return nil
}

// UploadFile puts one single file into an S3-compatible object storage.
func uploadFile(dest, src, contentType string, storage *minio.Client) error {
	ctx := context.Background()
	bucket := "qsdump"

	// Check if the output file already exists in storage.
	_, err := storage.StatObject(ctx, bucket, dest, minio.StatObjectOptions{})
	if err == nil {
		logmsg := fmt.Sprintf("Already in object storage: %s/%s", bucket, dest)
		fmt.Println(logmsg)
		if logger != nil {
			logger.Println(logmsg)
		}
		return nil
	}

	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err = storage.FPutObject(ctx, bucket, dest, src, opts)
	if err != nil {
		return err
	}

	logmsg := fmt.Sprintf("Uploaded to object storage: %s/%s", bucket, dest)
	fmt.Println(logmsg)
	if logger != nil {
		logger.Println(logmsg)
	}

	return nil
}
