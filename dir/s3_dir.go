package dir

import (
	"bytes"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	_ "github.com/aws/aws-sdk-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Dir struct {
	region, bucketName string
	s3Client           *s3.S3
}

func NewS3Dir(region, bucket string) (*S3Dir, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	svc := s3.New(awsSession)

	return &S3Dir{
		region:     region,
		bucketName: bucket,
		s3Client:   svc,
	}, nil
}

func (s3dir S3Dir) GetTypeName() string {
	return "s3"
}

func (s3dir S3Dir) Walk(Callback WalkFunc, PostChildrenCallback PostWalkFunc) error {
	return s3dir.walkFolder("", Callback, PostChildrenCallback)
}

func (s3dir S3Dir) walkFolder(prefix string, Callback WalkFunc, PostChildrenCallback PostWalkFunc) error {
	resp, err := s3dir.s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s3dir.bucketName),
		Prefix: aws.String(prefix),
		Delimiter: aws.String("/"),
		RequestPayer: aws.String("requester"),
	})
	if err != nil {
		return errors.Wrapf(err, "Failed to list items in bucket %q", s3dir.bucketName)
	}

	// enter the folder, but make sure to remove any trailing / before calling our handlers
	relPath := prefix
	if strings.HasSuffix(relPath, "/") {
		relPath = relPath[:len(relPath)-1]
	}
	if len(relPath) == 0 {
		relPath = "."
	}
	if err := Callback(relPath, 0, time.Time{}, true); err != nil {
		return err
	}

	// first iterate deeper into the hierarchy
	for _, item := range resp.CommonPrefixes {
		if err := s3dir.walkFolder(*item.Prefix, Callback, PostChildrenCallback); err != nil {
			return err
		}
	}

	// then iterate over the contained files
	for _, item := range resp.Contents {
		if strings.Compare(prefix, *item.Key) == 0 {
			continue
		}
		if err := Callback(*item.Key, *item.Size, *item.LastModified, false); err != nil {
			return err
		}
	}

	// exit the folder
	return PostChildrenCallback(prefix)
}

type nopReadSeekCloser struct {
	io.ReadSeeker
}

func (nopReadSeekCloser) Close() error { return nil }


func (s3dir S3Dir) GetFile(path string) (time.Time, ReadSeekCloser, error) {
	response, err := s3dir.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3dir.bucketName),
		Key: aws.String(path),
		RequestPayer: aws.String("requester"),
	})
	if err != nil {
		return time.Time{}, nil, errors.Wrapf(err, "Failed to get file `%v` from S3 bucket `%v`", path, s3dir.bucketName)
	}

	// now let's read the response and buffer it to enable seeking
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return time.Time{}, nil, errors.Wrapf(err, "Failed to read file `%v` from S3 bucket `%v`", path, s3dir.bucketName)
	}

	return *response.LastModified, nopReadSeekCloser{bytes.NewReader(data)}, nil
}
