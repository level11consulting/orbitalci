package admin

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SshSecret interface {
	UpdateSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
	GetSSHCred(context.Context, *pb.SSHKeyWrapper) (*pb.SSHKeyWrapper, error)
	SSHCredExists(context.Context, *pb.SSHKeyWrapper) (*pb.Exists, error)
	SetSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
	GetSSHCreds(context.Context, *empty.Empty) (*pb.SSHWrap, error)
	DeleteSSHCreds(context.Context, *pb.SSHKeyWrapper) (*empty.Empty, error)
}

func (g *guideOcelotServer) UpdateSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	return g.updateAnyCred(ctx, creds)
}

func (g *guideOcelotServer) SSHCredExists(ctx context.Context, creds *pb.SSHKeyWrapper) (*pb.Exists, error) {
	return g.checkAnyCredExists(ctx, creds)
}

func (g *guideOcelotServer) SetSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	if creds.SubType.Parent() != pb.CredType_SSH {
		return nil, status.Error(codes.InvalidArgument, "Subtype must be of ssh type: "+strings.Join(pb.CredType_SSH.SubtypesString(), " | "))
	}
	// no validation necessary, its a file upload

	err := SetupRCCCredentials(g.RemoteConfig, g.Storage, creds)
	if err != nil {
		if _, ok := err.(*pb.ValidationErr); ok {
			return &empty.Empty{}, status.Error(codes.InvalidArgument, "SSH Creds Upload failed validation. Errors are: "+err.Error())
		}
		return &empty.Empty{}, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

func (g *guideOcelotServer) GetSSHCreds(context.Context, *empty.Empty) (*pb.SSHWrap, error) {
	credWrapper := &pb.SSHWrap{}
	credz, err := g.RemoteConfig.GetCredsByType(g.Storage, pb.CredType_SSH, true)
	if err != nil {
		if _, ok := err.(*storage.ErrNotFound); ok {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return credWrapper, status.Errorf(codes.Internal, "unable to get ssh creds! error: %s", err.Error())
	}
	for _, v := range credz {
		credWrapper.Keys = append(credWrapper.Keys, v.(*pb.SSHKeyWrapper))
	}
	if len(credWrapper.Keys) == 0 {
		return credWrapper, status.Error(codes.NotFound, "no ssh keys found")
	}
	return credWrapper, nil
}

func (g *guideOcelotServer) GetSSHCred(ctx context.Context, credentials *pb.SSHKeyWrapper) (*pb.SSHKeyWrapper, error) {
	creddy, err := g.getAnyCred(credentials)
	if err != nil {
		return nil, err
	}
	ssh, ok := creddy.(*pb.SSHKeyWrapper)
	if !ok {
		return nil, status.Error(codes.Internal, "Unable to cast as SSH Creds")
	}
	return ssh, nil
}

func (g *guideOcelotServer) DeleteSSHCreds(ctx context.Context, creds *pb.SSHKeyWrapper) (*empty.Empty, error) {
	return g.deleteAnyCred(ctx, creds, pb.CredType_SSH)
}