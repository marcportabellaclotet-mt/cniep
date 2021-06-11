package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
)

var s3TemplateMap = map[string]s3Template{}

type s3Template struct {
	S3FilePath   string
	TemplateName string
	LastModified string
}

type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

func GetObjects(c context.Context, api S3ListObjectsAPI, input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return api.ListObjectsV2(c, input)
}

func listS3Objects(s3Path string) {
	if !strings.HasPrefix(s3Path, "s3://") {
		logrus.Error(fmt.Sprintf("%v is an invalid s3 path", s3Path))
		return
	}
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {})
	bucket := strings.Split(s3Path, "/")[2]
	if bucket == "" {
		logrus.Error(fmt.Sprintf("%v is an invalid s3 path", s3Path))
		return
	}
	prefix := strings.Split(s3Path, fmt.Sprintf("s3://%v/", bucket))[1]
	if prefix == "" {
		logrus.Error(fmt.Sprintf("%v is an invalid s3 path", s3Path))
		return
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = fmt.Sprintf("%v/", prefix)
	}
	input := &s3.ListObjectsV2Input{
		Bucket: String(bucket),
		Prefix: String(prefix),
	}
	resp, err := GetObjects(context.TODO(), client, input)
	if err != nil {
		logrus.Error("AWS S3 Error. Error getting Objects")
		logrus.Error(err.Error())
		return
	}
	for _, item := range resp.Contents {
		if item.Size != 0 {
			checkS3File(bucket, *item.Key, fmt.Sprintf("%v", *item.LastModified))
		}
	}
}

func checkS3File(bucket string, key string, LastModified string) {
	filePathSlice := strings.Split(key, "/")
	templateName := filePathSlice[(len(filePathSlice) - 2)]
	destFileName := filePathSlice[(len(filePathSlice) - 1)]
	s3FilePath := fmt.Sprintf("%v/%v", bucket, key)
	desiredS3Template := s3Template{
		S3FilePath:   s3FilePath,
		TemplateName: templateName,
		LastModified: LastModified,
	}
	if s3TemplateMap[s3FilePath] != desiredS3Template {
		s3TemplateMap[s3FilePath] = desiredS3Template
		downloadS3File(bucket, key, templateName, destFileName)
	}
}

func downloadS3File(bucket string, key string, templateName string, destFileName string) {

	dirName := fmt.Sprintf("%v/%v/", templatePath, templateName)
	if !dirExists(dirName) {
		os.MkdirAll(dirName, os.ModePerm)
	}
	destination := fmt.Sprintf("%v%v", dirName, destFileName)

	f, err := os.Create(destination)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	defer f.Close()
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {})
	object, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: String(bucket),
		Key:    String(key),
	})
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	r := ioutil.NopCloser(object.Body)
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	r.Close()
	_, err = f.Write(buf.Bytes())
	if err != nil {
		logrus.Error(fmt.Sprintf("Error writing local file : %v", destination))
		logrus.Error(err.Error())
	}
	logrus.Info(fmt.Sprintf("%v file succesfully downloaded from s3 bucket %v", destination, bucket))
}
