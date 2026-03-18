package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Driver struct {
	client       *minio.Client
	bucketName   string
	endpoint     string
	useSSL       bool
	customDomain string
}

// 确保具体类实现接口
var _ IUploader = (*S3Driver)(nil)

func NewS3Driver(endpoint, accessKey, secretKey, bucketName, customDomain string, useSSL bool) (*S3Driver, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// 2. 自动化基建：检查 Bucket 是否存在
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, fmt.Errorf("检查 Bucket 状态失败: %w", err)
	}
	// 3. 如果不存在，强行创建并注入只读 Policy
	if !exists {
		// 创建桶
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("自动创建 Bucket 失败: %w", err)
		}

		// 构造公开只读策略 JSON
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Action": ["s3:GetObject"],
					"Effect": "Allow",
					"Principal": {"AWS": ["*"]},
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucketName)

		// 注入策略
		err = client.SetBucketPolicy(ctx, bucketName, policy)
		if err != nil {
			return nil, fmt.Errorf("设置 Bucket 公开权限失败: %w", err)
		}
		fmt.Printf("🎉 自动化基建完成: Bucket [%s] 已创建并开启公共访问！\n", bucketName)
	}

	return &S3Driver{
		client:       client,
		bucketName:   bucketName,
		endpoint:     endpoint,
		useSSL:       useSSL,
		customDomain: customDomain,
	}, nil
}

// implement interface func
func (d *S3Driver) UploadImage(ctx context.Context, file io.Reader, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	objectName := uuid.New().String() + ext
	contentType := header.Header.Get("Content-Type")

	_, err := d.client.PutObject(ctx, d.bucketName, objectName, file, header.Size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	// 如果配置了 CDN 或自定义域名，直接返回干净的 URL
	if d.customDomain != "" {
		return fmt.Sprintf("%s/%s", strings.TrimRight(d.customDomain, "/"), objectName), nil
	}
	// 否则拼接默认的 S3 访问路径
	protocol := "http://"
	if d.useSSL {
		protocol = "https://"
	}
	return fmt.Sprintf("%s%s/%s/%s", protocol, d.endpoint, d.bucketName, objectName), nil
}
