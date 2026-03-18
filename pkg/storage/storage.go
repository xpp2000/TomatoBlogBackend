package storage

import (
	"context"
	"io"
	"mime/multipart"
)

/*
	storage 包单独暴露一个全局实例和初始化函数，和db，logger的方式不同
*/

type IUploader interface {
	UploadImage(ctx context.Context, file io.Reader, header *multipart.FileHeader) (string, error)
}

var GlobalUploader IUploader
