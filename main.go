package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var sess *session.Session

func init() {
	// Setup AWS S3 Session (build once use every function)
	sess = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	// Download soffice
	cmd := exec.Command("curl", "-L", "https://github.com/vladgolubev/serverless-libreoffice/releases/download/v6.1.0.0.alpha0/lo.tar.gz", "-o", "/tmp/lo.tar.gz")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("go-libreoffice: soffice download failure:%v\n", err)
	}

	cmd1 := exec.Command("tar", "-xvzf", "/tmp/lo.tar.gz", "-C", "/tmp")
	err = cmd1.Run()
	if err != nil {
		fmt.Printf("go-libreoffice: soffice unzip failure:%v\n", err)
	}
}

func handleRequest(ctx context.Context, s3Event events.S3Event) (string, error) {
	for _, record := range s3Event.Records {
		obj, err := s3.New(sess).GetObject(&s3.GetObjectInput{
			Bucket: &record.S3.Bucket.Name,
			Key:    &record.S3.Object.Key,
		})
		if err != nil {
			fmt.Printf("go-libreoffice: unexpected get S3 object failure:%v\n", err)
			return "", err
		}
		defer obj.Body.Close()

		// Download file to /tmp
		file, err := os.Create("/tmp/" + record.S3.Object.Key)
		if err != nil {
			fmt.Printf("go-libreoffice: unexpected file creation failure:%v\n", err)
			return "", err
		}
		defer file.Close()

		downloader := s3manager.NewDownloader(sess)
		_, err = downloader.Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(record.S3.Bucket.Name),
				Key:    aws.String(record.S3.Object.Key),
			})
		if err != nil {
			fmt.Printf("go-libreoffice: unexpected file download failure:%v\n", err)
			return "", err
		}

		// Convert file in /tmp directory
		cmd := exec.Command("/tmp/instdir/program/soffice", "--convert-to", "docx:writer_pdf_Export", file.Name(), "--outdir", "/tmp", "--headless")
		err = cmd.Run()
		if err != nil {
			fmt.Printf("go-libreoffice: converting failure:%v\n", err)
			return "", err
		}

		// Get converted file
		arr := strings.Split(record.S3.Object.Key, ".")
		cFile := arr[0] + ".docx"

		// Upload to S3
		err = addToS3Bucket(sess, record.S3.Bucket.Name, cFile)
		if err != nil {
			fmt.Printf("go-libreoffice: uploading to S3 failure:%v\n", err)
			return "", err
		}

		return "https://" + record.S3.Bucket.Name + ".s3." + os.Getenv("AWS_REGION") + ".amazonaws.com/" + cFile, nil
	}
	return "", nil
}

func addToS3Bucket(s *session.Session, bucketName string, fileDir string) error {
	file, err := os.Open("/tmp/"+fileDir)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size int64 = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	// Config settings: this is where you choose the bucket, filename, content-type etc.
	// of the file you're uploading.
	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(fileDir),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}

func main() {
	lambda.Start(handleRequest)
}
