package client

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/peertechde/argon/api"
	"github.com/peertechde/argon/pkg/logging"
	"github.com/peertechde/argon/pkg/storage"
)

const (
	defaultMaxMsgSize = 1024*1024*4 - 1024
)

var log = logging.Logger.WithField(logging.Subsys, "client")

func New(options ...Option) *Client {
	var opts Options
	opts.Apply(options...)

	return &Client{
		options: opts,
	}
}

type Client struct {
	options Options

	grpcClient    *grpc.ClientConn
	storageClient api.StorageClient
}

func (c *Client) DialContext(ctx context.Context, target string) error {
	var dialOptions []grpc.DialOption
	if c.options.TLSConfig != nil {
		dialOptions = append(dialOptions,
			grpc.WithTransportCredentials(credentials.NewTLS(c.options.TLSConfig)),
		)
	} else {
		dialOptions = append(dialOptions,
			grpc.WithInsecure(),
		)
	}
	cc, err := grpc.DialContext(ctx, target, dialOptions...)
	if err != nil {
		return err
	}
	c.grpcClient = cc
	c.storageClient = api.NewStorageClient(c.grpcClient)

	return nil
}

func (c *Client) Read(ctx context.Context, name, dst string) error {
	stream, err := c.storageClient.Read(ctx, &api.ReadRequest{Name: name})
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	var size int
	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		_, err = buf.Write(resp.Data)
		if err != nil {
			return err
		}
		size += len(resp.Data)
	}
	n, err := save(dst, buf.Bytes())
	if err != nil {
		return err
	}

	// TODO:
	if n != size {
	}

	return nil
}

func (c *Client) Write(ctx context.Context, name string) error {
	fd, err := os.Open(name)
	if err != nil {
		return errors.Wrapf(err, "failed to open file")
	}
	defer fd.Close()

	stream, err := c.storageClient.Write(ctx)
	if err != nil {
		return err
	}

	if err := stream.Send(&api.WriteRequest{Member: &api.WriteRequest_Name{Name: filepath.Base(name)}}); err != nil {
		s := status.Convert(err)
		for _, d := range s.Details() {
			switch info := d.(type) {
			default:
				log.Printf("Unexpected type: %s", info)
			}
		}
		return errors.Wrap(err, "failed to send")
	}

	rd := bufio.NewReader(fd)
	buf := make([]byte, defaultMaxMsgSize)
	for {
		n, err := rd.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.Wrap(err, "failed to read file")
		}
		if err := stream.Send(&api.WriteRequest{Member: &api.WriteRequest_Data{Data: buf[:n]}}); err != nil {
			return errors.Wrap(err, "failed to send")
		}
	}

	if _, err := stream.CloseAndRecv(); err != nil {
		return err
	}

	return nil
}

func (c *Client) List(ctx context.Context) ([]string, error) {
	resp, err := c.storageClient.List(ctx, &api.ListRequest{})
	if err != nil {
		return nil, err
	}
	return resp.Files, nil
}

func (c *Client) Stat(ctx context.Context, name string) (*storage.FileInfo, error) {
	resp, err := c.storageClient.Stat(ctx, &api.StatRequest{Name: name})
	if err != nil {
		return nil, err
	}
	fileInfo := &storage.FileInfo{
		Name:    resp.FileInfo.Name,
		Size:    resp.FileInfo.Size,
		Mode:    resp.FileInfo.Mode,
		ModTime: resp.FileInfo.ModTime.AsTime(),
		Dir:     resp.FileInfo.Dir,
	}
	return fileInfo, nil
}

func (c *Client) Remove(ctx context.Context, name string) error {
	_, err := c.storageClient.Remove(ctx, &api.RemoveRequest{Name: name})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Rename(ctx context.Context, old, new string) error {
	_, err := c.storageClient.Rename(ctx, &api.RenameRequest{Old: old, New: new})
	if err != nil {
		return err
	}
	return nil
}

func save(name string, data []byte) (int, error) {
	fd, err := os.Create(name)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create file")
	}
	defer fd.Close()

	n, err := fd.Write(data)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write file")
	}
	return n, nil
}
