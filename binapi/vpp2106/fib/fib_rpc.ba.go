// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package fib

import (
	"context"
	"fmt"
	"io"
	"mocknet/binapi/vpp2106/vpe"

	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service fib.
type RPCService interface {
	FibSourceAdd(ctx context.Context, in *FibSourceAdd) (*FibSourceAddReply, error)
	FibSourceDump(ctx context.Context, in *FibSourceDump) (RPCService_FibSourceDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) FibSourceAdd(ctx context.Context, in *FibSourceAdd) (*FibSourceAddReply, error) {
	out := new(FibSourceAddReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) FibSourceDump(ctx context.Context, in *FibSourceDump) (RPCService_FibSourceDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_FibSourceDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_FibSourceDumpClient interface {
	Recv() (*FibSourceDetails, error)
	api.Stream
}

type serviceClient_FibSourceDumpClient struct {
	api.Stream
}

func (c *serviceClient_FibSourceDumpClient) Recv() (*FibSourceDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *FibSourceDetails:
		return m, nil
	case *vpe.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
