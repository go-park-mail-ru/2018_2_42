package grpc_authorisation_client

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
	"time"

	pb "github.com/go-park-mail-ru/2018_2_42/grpc_shared_code"
)

type DetailedError struct {
	UnderlyingErr error
	Code          int // http.Status*
}

func (de *DetailedError) Error() string {
	return http.StatusText(de.Code) + ": " + de.UnderlyingErr.Error()
}

const (
	address = "localhost:50051"
)

func Worker(token string) (login string, avatar string, err error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure()) //TODO : env
	if err != nil {
		err = errors.Wrap(err, "grpc_authorisation_client.go: Worker: can not connect: ")
		return
	}
	defer conn.Close()
	c := pb.NewAuthorizationCheckClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	user, err := c.GetLogin(ctx, &pb.Cookie{Sessionid: token})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.NotFound {
			err = &DetailedError{
				UnderlyingErr: err,
				Code:          http.StatusForbidden,
			}
			return
		}
		err = &DetailedError{
			UnderlyingErr: err,
			Code:          http.StatusInternalServerError,
		}
		log.Print("grpc_authorisation_client.go: Worker: " + err.Error())
		return
	}
	login = user.Login
	avatar = user.AvatarAddress
	return
}
