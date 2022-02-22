package server

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/peertechde/argon/api"
)

func NewStorageService(root string) *StorageService {
	return &StorageService{
		root: root,
	}
}

type StorageService struct {
	api.UnimplementedStorageServer

	root string
}

func (s *StorageService) path(name string) string {
	return filepath.Join(s.root, name)
}

func (s *StorageService) Upload(stream api.Storage_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid argument")
	}

	name := req.GetName()
	scopedLog := log.WithFields(logrus.Fields{
		"name": name,
	})
	scopedLog.Info("Received upload request")

	if _, err := os.Stat(s.path(name)); err == nil {
		scopedLog.Info("File already exists")
		return status.Errorf(codes.AlreadyExists, "file (%s) already exists", name)
	}

	var buf bytes.Buffer
	var size int
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			scopedLog.Errorf("Failed to receive data (%s)", err)
			return err
		}
		scopedLog.Debugf("Received %d bytes of data", len(req.GetData()))

		_, err = buf.Write(req.GetData())
		if err != nil {
			log.Errorf("Failed to copy data (%s)", err)
			return err
		}
		size += len(req.GetData())
	}

	n, err := s.save(name, buf.Bytes())
	if err != nil {
		return status.Errorf(codes.Internal, "failed to save file")
	}

	// TODO: check for "short writes"
	if n != size {
	}

	if err := stream.SendAndClose(&api.UploadResponse{}); err != nil {
		log.Errorf("Failed to close the connection (%s)", err)
		return err
	}

	return nil
}

func (s *StorageService) Download(req *api.DownloadRequest, stream api.Storage_DownloadServer) error {
	scopedLog := log.WithFields(logrus.Fields{
		"name": req.Name,
	})
	scopedLog.Info("Received download request")

	if _, err := os.Stat(s.path(req.Name)); errors.Is(err, os.ErrNotExist) {
		return status.Errorf(codes.AlreadyExists, "file (%s) doesn't exists", req.Name)
	}

	fd, err := os.Open(s.path(req.Name))
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open file")
	}
	defer fd.Close()

	buf := make([]byte, defaultMaxMsgSize-1024)
	for {
		n, err := fd.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return status.Errorf(codes.Internal, "failed to read file")
		}
		if err := stream.Send(&api.DownloadResponse{Data: buf}); err != nil {
			return status.Errorf(codes.Internal, "failed to send data")
		}
		scopedLog.Debugf("Send %d bytes of data", n)
	}

	scopedLog.Info("Successfully handled download request")

	return nil
}

func (s *StorageService) save(name string, data []byte) (int, error) {
	fd, err := os.Create(s.path(name))
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
