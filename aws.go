package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3_REGION = "us-east-2"
	S3_BUCKET = "titanstockbot"
)

func UploadFileToS3() {
	fmt.Println("uploading entries to aws...")
	// Open the file for use
	file, err := os.Open(EntriesDB)
	if err != nil {
		fmt.Println("file open error for upload")
		return
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println("file read error for upload: ", err.Error())
		return
	}
	s, err := createSession()
	if err != nil {
		fmt.Println("session creation error: ", err.Error())
		return
	}
	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(EntriesDB),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	if err != nil {
		fmt.Println("aws upload err: ", err.Error())
	}
	return
}

func DownloadFileFromS3() {
	fmt.Println("downloading from AWS...")
	f, err := os.Create(EntriesDB)
	if err != nil {
		fmt.Println("file creation error in download: ", err.Error())
		return
	}
	s, err := createSession()
	if err != nil {
		fmt.Println("session creation error: ", err.Error())
		return
	}
	downloader := s3manager.NewDownloader(s)
	_, err = downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(S3_BUCKET),
		Key:    aws.String(EntriesDB),
	})
	if err != nil {
		fmt.Println("aws download error: ", err.Error())
		return
	}
}

func createSession() (*session.Session, error) {
	token := ""
	// setup creds
	creds := credentials.NewStaticCredentials(os.Getenv("aws-key"), os.Getenv("aws-secret"), token)
	_, err := creds.Get()
	if err != nil {
		return nil, err
	}
	// Create a single AWS session (we can re use this if we're uploading many files)
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION), Credentials: creds})
	return s, err
}
