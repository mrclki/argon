package server

import (
	"bytes"
	"context"
	"io"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/peertechde/argon/api"
	"github.com/peertechde/argon/pkg/storage"
)

func NewStorageService(store storage.Storage) *StorageService {
	return &StorageService{
		store: store,
	}
}

type StorageService struct {
	api.UnimplementedStorageServer

	store storage.Storage
}

func (s *StorageService) Read(req *api.ReadRequest, stream api.Storage_ReadServer) error {
	scopedLog := log.WithFields(logrus.Fields{
		"name": req.Name,
	})
	scopedLog.Info("Handling read request")

	data, err := s.store.Read(stream.Context(), req.Name)
	if err != nil {
		if errors.Is(err, &storage.NotFoundError{Name: req.Name}) {
			return status.Errorf(codes.InvalidArgument, "file (%s) does not exist", req.Name)
		}
		return status.Errorf(codes.Internal, "failed to read file")
	}

	rd := bytes.NewReader(data)
	buf := make([]byte, defaultMaxMsgSize-1024)
	for {
		n, err := rd.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return status.Errorf(codes.Internal, "failed to read file")
		}
		if err := stream.Send(&api.ReadResponse{Data: buf[:n]}); err != nil {
			return status.Errorf(codes.Internal, "failed to send data")
		}
		scopedLog.Debugf("Send %d bytes of data", n)
	}

	scopedLog.Info("Successfully handled read request")

	return nil
}

func (s *StorageService) Write(stream api.Storage_WriteServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid argument")
	}

	name := req.GetName()
	scopedLog := log.WithFields(logrus.Fields{
		"name": name,
	})
	scopedLog.Info("Handling write request")

	if _, err := s.store.Stat(stream.Context(), name); err == nil {
		return status.Errorf(codes.AlreadyExists, "file %s already exists")
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

	if err := s.store.Write(stream.Context(), name, buf.Bytes()); err != nil {
		scopedLog.Errorf("Failed to write file (%s)", err)
		return status.Errorf(codes.Internal, "failed to write file")
	}

	if err := stream.SendAndClose(&api.WriteResponse{}); err != nil {
		log.Errorf("Failed to close the connection (%s)", err)
		return err
	}

	scopedLog.Info("Successfully handled write request")
	return nil
}

func (s *StorageService) List(ctx context.Context, req *api.ListRequest) (*api.ListResponse, error) {
	log.Info("Handling list request")

	files, err := s.store.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list files")
	}

	log.Info("Successfully handled list request")
	return &api.ListResponse{Files: files}, nil
}

func (s *StorageService) Stat(ctx context.Context, req *api.StatRequest) (*api.StatResponse, error) {
	scopedLog := log.WithFields(logrus.Fields{
		"name": req.Name,
	})
	scopedLog.Info("Handling stat request")

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file name")
	}

	fi, err := s.store.Stat(ctx, req.Name)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to stat file %s", req.Name)
	}
	fileInfo := &api.FileInfo{
		Name:    fi.Name,
		Size:    fi.Size,
		Mode:    fi.Mode,
		ModTime: timestamppb.New(fi.ModTime),
		Dir:     fi.Dir,
	}

	scopedLog.Info("Successfully handled stat request")
	return &api.StatResponse{FileInfo: fileInfo}, nil
}

func (s *StorageService) Remove(ctx context.Context, req *api.RemoveRequest) (*api.RemoveResponse, error) {
	scopedLog := log.WithFields(logrus.Fields{
		"name": req.Name,
	})
	scopedLog.Info("Handling remove request")

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid file name")
	}

	if err := s.store.Remove(ctx, req.Name); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove file %s", req.Name)
	}

	scopedLog.Info("Successfully handled remove request")
	return &api.RemoveResponse{}, nil
}

func (s *StorageService) Rename(ctx context.Context, req *api.RenameRequest) (*api.RenameResponse, error) {
	scopedLog := log.WithFields(logrus.Fields{
		"old": req.Old,
		"new": req.New,
	})
	scopedLog.Info("Handling rename request")

	if req.Old == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid old file name")
	}
	if req.New == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new file name")
	}

	if err := s.store.Rename(ctx, req.Old, req.New); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to rename file %s to %s", req.Old, req.New)
	}

	scopedLog.Info("Successfully handled rename request")
	return &api.RenameResponse{}, nil
}
