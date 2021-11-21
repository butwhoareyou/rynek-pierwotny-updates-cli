package s3

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine"
	"github.com/butwhoareyou/rynek-pierwotny-updates-cli/store/engine/file"
	"io/ioutil"
)

var contentType = "application/json"

type Engine struct {
	bucket string
	s3     *s3.S3
}

func NewEngine(bucket string, svc *s3.S3) (engine.Engine, error) {
	eng := &Engine{bucket: bucket, s3: svc}
	return eng, nil
}

func (e *Engine) Read(path string) ([]byte, error) {
	obj, err := e.s3.GetObject(&s3.GetObjectInput{
		Bucket: &e.bucket,
		Key:    &path,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return make([]byte, 0), file.NoPathError(aerr.Message())
			default:
				return make([]byte, 0), err
			}
		}
		return make([]byte, 0), err
	}
	return ioutil.ReadAll(obj.Body)
}

func (e *Engine) Write(path string, b []byte) error {
	r := bytes.NewReader(b)
	cl := int64(len(b))
	_, err := e.s3.PutObject(&s3.PutObjectInput{
		Bucket:        &e.bucket,
		Key:           &path,
		Body:          r,
		ContentLength: &cl,
		ContentType:   &contentType,
	})
	return err
}

func (e *Engine) Exists(path string) (bool, error) {
	_, err := e.s3.HeadObject(&s3.HeadObjectInput{
		Bucket: &e.bucket,
		Key:    &path,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return false, nil
			default:
				return false, err
			}
		}
	}
	return true, nil
}
