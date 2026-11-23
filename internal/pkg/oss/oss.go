package oss

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"redbook-go/internal/config"
)

type Client struct {
	cfg config.OSSConfig
}

func New(cfg config.OSSConfig) *Client {
	return &Client{cfg: cfg}
}

func (c *Client) Configured() error {
	if c.cfg.AccessKeyID == "" || c.cfg.AccessKeySecret == "" || c.cfg.Bucket == "" {
		return errors.New("请在配置或环境变量中填写有效的 OSS 参数")
	}
	return nil
}

func (c *Client) PresignedPut(ext, contentType string) (map[string]any, error) {
	if err := c.Configured(); err != nil {
		return nil, err
	}
	if ext == "" {
		ext = ".jpg"
	} else if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	key := "uploads/" + randomID() + ext
	expires := time.Now().Add(time.Hour).Unix()
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%d\n%s", contentType, expires, canonicalResource(c.cfg.Bucket, key))
	signature := sign(c.cfg.AccessKeySecret, stringToSign)

	base := c.bucketEndpoint() + "/" + key
	u, err := url.Parse(base)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("OSSAccessKeyId", c.cfg.AccessKeyID)
	q.Set("Expires", fmt.Sprintf("%d", expires))
	q.Set("Signature", signature)
	u.RawQuery = q.Encode()

	return map[string]any{
		"uploadUrl":       u.String(),
		"objectKey":       key,
		"publicUrl":       c.publicURL(key),
		"putAclPublicRead": false,
	}, nil
}

func (c *Client) UploadViaAPI(file multipart.File, header *multipart.FileHeader) (map[string]any, error) {
	if err := c.Configured(); err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("请选择文件")
	}
	if header.Size > 10*1024*1024 {
		return nil, errors.New("文件过大，单文件不超过 10MB")
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	key := "uploads/" + randomID() + extFromFilename(header.Filename)
	data, err := io.ReadAll(io.LimitReader(file, 10*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 10*1024*1024 {
		return nil, errors.New("文件过大，单文件不超过 10MB")
	}
	date := time.Now().UTC().Format(http.TimeFormat)
	stringToSign := fmt.Sprintf("PUT\n\n%s\n%s\n%s", contentType, date, canonicalResource(c.cfg.Bucket, key))
	signature := sign(c.cfg.AccessKeySecret, stringToSign)
	url := c.bucketEndpoint() + "/" + key
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "OSS "+c.cfg.AccessKeyID+":"+signature)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("上传失败: %s %s", resp.Status, string(body))
	}
	return map[string]any{"objectKey": key, "publicUrl": c.publicURL(key), "putAclPublicRead": false}, nil
}

func (c *Client) endpoint() string {
	ep := c.cfg.Endpoint
	if !strings.HasPrefix(ep, "http://") && !strings.HasPrefix(ep, "https://") {
		ep = "https://" + ep
	}
	return strings.TrimRight(ep, "/")
}

func (c *Client) bucketEndpoint() string {
	ep := c.cfg.Endpoint
	ep = strings.TrimPrefix(strings.TrimPrefix(ep, "https://"), "http://")
	host := c.cfg.Bucket + "." + ep
	scheme := "https"
	if strings.HasPrefix(c.cfg.Endpoint, "http://") {
		scheme = "http"
	}
	return scheme + "://" + host
}

func (c *Client) publicURL(key string) string {
	if c.cfg.URLPrefix != "" {
		return strings.TrimRight(c.cfg.URLPrefix, "/") + "/" + key
	}
	ep := c.cfg.Endpoint
	ep = strings.TrimPrefix(strings.TrimPrefix(ep, "https://"), "http://")
	return "https://" + c.cfg.Bucket + "." + ep + "/" + key
}

func canonicalResource(bucket, key string) string {
	return "/" + bucket + "/" + key
}

func sign(secret, stringToSign string) string {
	mac := hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func extFromFilename(name string) string {
	i := strings.LastIndex(name, ".")
	if i < 0 || i == len(name)-1 {
		return ".jpg"
	}
	return name[i:]
}
