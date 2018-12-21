package AWS_upload_awatar

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
)

type AWSUploader struct {
	Bucket string
	Session *session.Session
	Uploader *s3manager.Uploader
}

func NewAWSUploader() (u *AWSUploader, err error) {
	awsUploader := AWSUploader{
		Bucket: "avatars-rpsarena-ru",
	}
	awsUploader.Session, err = session.NewSession()
	awsUploader.Uploader = s3manager.NewUploader(awsUploader.Session)
	return
}

// Upload the file's body to S3 bucket as an object with the key being the
// same as the filename.
func (u *AWSUploader) Upload(image io.Reader, filename string) (err error) {
	_, err = u.Uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(u.Bucket),

		// Can also use the `filepath` standard library package to modify the
		// filename as need for an S3 object key. Such as turning absolute path
		// to a relative path.
		Key: aws.String(filename),

		// The file to be uploaded. io.ReadSeeker is preferred as the Uploader
		// will be able to optimize memory when uploading large content. io.Reader
		// is supported, but will require buffering of the reader's bytes for
		// each part.
		Body: image,
	})
	return
}
