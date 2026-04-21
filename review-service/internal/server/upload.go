package server

import (
	"bytes"
	"fmt"
	"io"
	stdhttp "net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	maxUploadBytes = 10 << 20
)

type uploadImageReply struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

func registerUploadRoutes(srv *khttp.Server, logger log.Logger) {
	uploadRoot := resolveUploadRoot()
	_ = os.MkdirAll(filepath.Join(uploadRoot, "review"), 0755)

	srv.HandlePrefix("/static/", stdhttp.StripPrefix("/static/", stdhttp.FileServer(stdhttp.Dir(uploadRoot))))

	r := srv.Route("/")
	r.POST("/v1/upload/review-image", func(ctx khttp.Context) error {
		req := ctx.Request()
		if err := req.ParseMultipartForm(maxUploadBytes); err != nil {
			return ctx.Result(400, map[string]string{"message": "上传文件解析失败"})
		}

		file, header, err := req.FormFile("file")
		if err != nil {
			return ctx.Result(400, map[string]string{"message": "请上传字段名为 file 的图片"})
		}
		defer file.Close()

		if header.Size > maxUploadBytes {
			return ctx.Result(400, map[string]string{"message": "图片大小不能超过 10MB"})
		}

		data, err := io.ReadAll(io.LimitReader(file, maxUploadBytes+1))
		if err != nil {
			return ctx.Result(500, map[string]string{"message": "读取上传文件失败"})
		}
		if int64(len(data)) > maxUploadBytes {
			return ctx.Result(400, map[string]string{"message": "图片大小不能超过 10MB"})
		}

		ext, ok := detectImageExt(data, header.Filename)
		if !ok {
			return ctx.Result(400, map[string]string{"message": "仅支持 jpg/jpeg/png/webp/gif 图片"})
		}

		now := time.Now()
		relativeDir := filepath.Join("review", now.Format("2006"), now.Format("01"), now.Format("02"))
		targetDir := filepath.Join(uploadRoot, relativeDir)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			log.NewHelper(logger).WithContext(ctx).Errorf("create upload dir failed: %v", err)
			return ctx.Result(500, map[string]string{"message": "创建上传目录失败"})
		}

		filename := fmt.Sprintf("%d_%d%s", now.UnixMilli(), time.Now().Nanosecond()%100000, ext)
		targetPath := filepath.Join(targetDir, filename)
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			log.NewHelper(logger).WithContext(ctx).Errorf("write upload file failed: %v", err)
			return ctx.Result(500, map[string]string{"message": "保存图片失败"})
		}

		reply := &uploadImageReply{
			URL:  path.Join("/static", filepath.ToSlash(relativeDir), filename),
			Name: filename,
			Size: int64(len(data)),
		}
		return ctx.Result(200, reply)
	})
}

func resolveUploadRoot() string {
	if p := os.Getenv("REVIEW_UPLOAD_DIR"); p != "" {
		return p
	}

	wd, err := os.Getwd()
	if err != nil {
		return filepath.Join(".", "uploads")
	}

	dir := wd
	for range 16 {
		if filepath.Base(dir) == "review-service" {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return filepath.Join(dir, "uploads")
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return filepath.Join(wd, "uploads")
}

func detectImageExt(data []byte, filename string) (string, bool) {
	contentType := stdhttp.DetectContentType(data)
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	case "image/gif":
		return ".gif", true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) || ext != ".jpg" && ext != ".jpeg" {
			return ext, true
		}
	}
	return "", false
}
