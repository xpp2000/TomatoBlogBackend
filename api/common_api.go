package api

import (
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/model"
	"tomatoBlogDB/pkg/storage"

	"github.com/kataras/iris/v12"
)

type CommonApi struct {
	*BaseApi
}

func NewCommonApi() *CommonApi {
	return &CommonApi{
		BaseApi: NewBaseApi(),
	}
}

// @Tags Common
// @Summary Upload Image
// @Description 统一图片上传接口 (支持 Markdown 编辑器自动上传),处理blog图片
// @Accept multipart/form-data
// @Param file formData file true "图片文件"
// @Success 200 {object} model.ResponseJson{data=dto.UploadImageResp}
// @Router /api/v1/private/common/uploadimg [post]
func (m *CommonApi) UploadImage(ctx iris.Context) {
	m.SetContext(ctx)
	// 1. 硬性限制单次请求最大 5MB，防止恶意刷爆内存和带宽
	// 注册路由时，也应该限制最大size
	ctx.SetMaxRequestBodySize(6 * iris.MB)

	// 2. 接收文件流 (表单字段名约定为 "file")
	file, header, err := ctx.FormFile("file")
	if err != nil {
		// 如果前端没传 file 字段，或者文件超大，在这里拦截
		m.HandleError(ctx, errcode.ErrReadUploadFileFail)
		return
	}
	// 🌟 极其重要：处理完必须关闭文件句柄，否则会导致服务器内存和文件描述符泄漏！
	defer file.Close()

	// 3. MIME format check
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".svg": true,
	}
	if !validExts[ext] {
		m.HandleError(ctx, errcode.NewBizErr(40002, "不支持的文件后缀格式，仅支持常用图片"))
		return
	}

	// 校验 MIME 类型 (防伪造：防止把 .exe 改成 .png)
	// 读取 HTTP Header 中浏览器嗅探到的类型
	contentType := header.Header.Get("Content-Type")
	// 确保它确实是以 "image/" 开头的类型
	if !strings.HasPrefix(contentType, "image/") {
		m.HandleError(ctx, errcode.NewBizErr(40003, "非法的文件内容，伪造的图片类型"))
		return
	}

	// 4. Compress. 分流
	var finalReader io.Reader = file           // 默认使用原始文件流
	var finalSize int64 = header.Size          // 默认原始大小
	var finalFilename string = header.Filename // 默认原始文件名
	var finalContentType string = contentType  // 默认原始类型
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		webpBuffer, err := storage.SmartCompressToWebP(file, 200)
		if err != nil {
			m.HandleError(ctx, errcode.NewSysErr(err))
			return
		}
		// 覆盖变量，切换到压缩后的数据
		finalReader = webpBuffer
		finalSize = int64(webpBuffer.Len())
		// 替换后缀名，例如把 avatar.png 变成 avatar.png.webp
		// (如果想完美替换，可以用 strings.TrimSuffix(header.Filename, ext) + ".webp")
		finalFilename = strings.TrimSuffix(header.Filename, ext) + ".webp"
		finalContentType = "image/webp"
	} else {
		// SVG, GIF, 原生 WebP
	}
	if err != nil {
		m.HandleError(ctx, errcode.NewSysErr(err))
		return
	}
	// 4. 构造伪造的 FileHeader，骗过底层的 OSS 上传引擎
	// 因为我们把图片变了，后缀名和大小都变了，需要重新告诉 OSS
	newHeader := &multipart.FileHeader{
		Filename: finalFilename,
		Size:     finalSize,
		Header:   header.Header,
	}
	newHeader.Header.Set("Content-Type", finalContentType)
	// 5. 调用我们封装好的高可用上传引擎
	// 传入 ctx.Request().Context() 可以让底层 SDK 继承 HTTP 请求的超时控制
	url, err := storage.GlobalUploader.UploadImage(ctx.Request().Context(), finalReader, newHeader)
	if err != nil {
		// 上传到 OSS 失败，记入系统日志
		m.HandleError(ctx, errcode.NewSysErr(err))
		return
	}

	// 4. 返回前端喜闻乐见的格式
	// 提示：很多 Markdown 编辑器 (如 Vditor, Bytemd) 对返回值有特定要求
	// 通常把 url 放在 data 里面是最标准的做法
	m.Ok(model.ResponseJson{
		Code: 20000,
		Msg:  "上传成功",
		Data: dto.UploadImageResp{
			Url: url,
		},
	})
}
