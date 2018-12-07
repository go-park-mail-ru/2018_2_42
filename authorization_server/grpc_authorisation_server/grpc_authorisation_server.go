package grpc_authorisation_server

import (
	"context"
	pb "github.com/go-park-mail-ru/2018_2_42/grpc_shared_code"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"

	"github.com/go-park-mail-ru/2018_2_42/authorization_server/environment"
)

type ServerEnvironment struct {
	environment.Environment
}

func (se *ServerEnvironment) GetLogin(ctx context.Context, cookie *pb.Cookie) (user *pb.User, err error) {
	exist, userInfo, err := se.DB.SelectUserBySessionId(cookie.Sessionid)
	if err != nil {
		err = status.Error(codes.Internal, "grpc_authorisation_server.go: GetLogin: SelectUserBySessionId: "+err.Error())
		log.Print(err)
		return
	}
	if !exist {
		err = status.Error(codes.NotFound, "user with cookie.sessionid='"+cookie.Sessionid+"' not found")
		return
	}
	user = &pb.User{Login: userInfo.Login, AvatarAddress: userInfo.AvatarAddress}
	return
}

func Worker(se *ServerEnvironment) (err error) {
	lis, err := net.Listen("tcp", se.Config.DataSourceName)
	if err != nil {
		err = errors.Wrap(err, "grpc_authorisation_client.go: ListenAndServe: failed to listen: ")
		return
	}
	newServer := grpc.NewServer()
	pb.RegisterAuthorizationCheckServer(newServer, se)
	// Register reflection service on gRPC server.
	reflection.Register(newServer)
	if err := newServer.Serve(lis); err != nil {
		err = errors.Wrap(err, "ServerEnvironment.ListenAndServe: failed to serve: ")
	}
	return
}
