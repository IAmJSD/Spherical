package s3

import (
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type urlFrag struct {
	encode bool
	part   string
}

func urlConcat(start string, parts ...urlFrag) string {
	if !strings.HasPrefix(start, "http://") && !strings.HasPrefix(start, "https://") {
		start = "https://" + start
	}
	if !strings.HasSuffix(start, "/") {
		start += "/"
	}
	plm1 := len(parts) - 1
	for i, v := range parts {
		if v.encode {
			start += url.PathEscape(v.part)
		} else {
			start += v.part
		}
		if i != plm1 {
			start += "/"
		}
	}
	return start
}

// Upload is used to upload a file. Returns the URL to it.
func Upload(ctx context.Context, path string, r io.Reader, contentLength int64, contentType, contentDisposition, acl string) (string, error) {
	// Do the upload.
	c := config.Config()
	endpoint := c.S3Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	secure := strings.HasPrefix(endpoint, "https://")
	cli, err := minio.New(strings.SplitN(endpoint, "://", 2)[1], &minio.Options{
		Creds:  credentials.NewStaticV4(c.S3AccessKeyID, c.S3SecretAccessKey, ""),
		Secure: secure,
		Region: c.S3Region,
	})
	if err != nil {
		return "", err
	}
	_, err = cli.PutObject(ctx, c.S3Bucket, path, r, contentLength, minio.PutObjectOptions{
		ContentType:        contentType,
		ContentDisposition: contentDisposition,
		UserMetadata: map[string]string{
			"x-amz-acl": acl,
		},
	})
	if err != nil {
		return "", err
	}

	// Get the resulting URL if it is just a hostname concat.
	if c.S3Hostname != "" {
		return urlConcat(c.S3Hostname, urlFrag{
			encode: false,
			part:   path,
		}), nil
	}

	// Handle endpoint and bucket concats.
	return urlConcat(c.S3Endpoint, urlFrag{
		encode: true,
		part:   c.S3Bucket,
	}, urlFrag{
		encode: false,
		part:   path,
	}), nil
}

// Delete is used to delete a file at a path.
func Delete(ctx context.Context, path string) error {
	c := config.Config()
	endpoint := c.S3Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	secure := strings.HasPrefix(endpoint, "https://")
	cli, err := minio.New(strings.SplitN(endpoint, "://", 2)[1], &minio.Options{
		Creds:  credentials.NewStaticV4(c.S3AccessKeyID, c.S3SecretAccessKey, ""),
		Secure: secure,
		Region: c.S3Region,
	})
	if err != nil {
		return err
	}
	return cli.RemoveObject(ctx, c.S3Bucket, path, minio.RemoveObjectOptions{})
}
