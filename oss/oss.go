package oss

import (
	"bytes"
	"errors"
	"io/ioutil"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Config struct {
	Endpoint  string `toml:"endpoint" json:"endpoint"`
	AccessKey string `toml:"access_key" json:"access_key"`
	SecretKey string `toml:"secret_key" json:"secret_key"`
	Bucket    string `toml:"bucket" json:"bucket"`
}

type Client struct {
	c      *Config
	bucket *oss.Bucket
}

func (c *Client) Bucket() string {
	return c.c.Bucket
}

func (c *Client) Put(objectKey string, data []byte, contentType string) error {
	return c.bucket.PutObject(objectKey, bytes.NewReader(data), oss.ContentType(contentType))
}

func (c *Client) PutFromFile(objectKey, filePath string) error {
	return c.bucket.PutObjectFromFile(objectKey, filePath)
}

func (c *Client) Exist(objectKey string) (exist bool, err error) {
	return c.bucket.IsObjectExist(objectKey)
}

func (c *Client) Get(objectKey string) (data []byte, contentType string, err error) {
	meta, err := c.bucket.GetObjectDetailedMeta(objectKey)

	if err != nil {
		return
	}

	contentType = meta.Get(oss.HTTPHeaderContentType)
	reader, err := c.bucket.GetObject(objectKey)

	if err != nil {
		return
	}

	defer func() { _ = reader.Close() }()

	data, err = ioutil.ReadAll(reader)
	return
}

func (c *Client) GetToFile(objectKey, filePath string) error {
	return c.bucket.GetObjectToFile(objectKey, filePath, oss.AcceptEncoding("gzip"))
}

func (c *Client) Delete(objectKey string) error {
	return c.bucket.DeleteObject(objectKey)
}

func (c *Client) Deletes(objectKeys []string) (deletedKeys []string, err error) {
	res, err := c.bucket.DeleteObjects(objectKeys, oss.DeleteObjectsQuiet(true))

	if err != nil {
		return
	}

	deletedKeys = res.DeletedObjects
	return
}

func NewClient(c *Config) (client *Client, err error) {
	client = &Client{c: c}

	ossCon, err := oss.New(c.Endpoint, c.AccessKey, c.SecretKey)

	if err != nil {
		return
	}

	client.bucket, err = ossCon.Bucket(c.Bucket)
	return
}

type Pool struct {
	locker  sync.RWMutex
	clients map[string]*Client
}

func (p *Pool) Add(name string, conf *Config) (err error) {
	p.locker.Lock()
	defer p.locker.Unlock()

	client, err := NewClient(conf)

	if err != nil {
		return
	}

	p.clients[name] = client
	return
}

func (p *Pool) Get(name string) (client *Client, err error) {
	p.locker.RLock()
	defer p.locker.RUnlock()

	client, ok := p.clients[name]

	if !ok {
		err = errors.New("no oss client")
	}

	return
}

func NewPool() *Pool {
	return &Pool{clients: make(map[string]*Client, 16)}
}
