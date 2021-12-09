// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package policer

import (
	"context"
	"fmt"
	"io"
	"mocknet/binapi/vpp2110/vpe"

	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service policer.
type RPCService interface {
	PolicerAddDel(ctx context.Context, in *PolicerAddDel) (*PolicerAddDelReply, error)
	PolicerBind(ctx context.Context, in *PolicerBind) (*PolicerBindReply, error)
	PolicerDump(ctx context.Context, in *PolicerDump) (RPCService_PolicerDumpClient, error)
	PolicerInput(ctx context.Context, in *PolicerInput) (*PolicerInputReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) PolicerAddDel(ctx context.Context, in *PolicerAddDel) (*PolicerAddDelReply, error) {
	out := new(PolicerAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) PolicerBind(ctx context.Context, in *PolicerBind) (*PolicerBindReply, error) {
	out := new(PolicerBindReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) PolicerDump(ctx context.Context, in *PolicerDump) (RPCService_PolicerDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_PolicerDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_PolicerDumpClient interface {
	Recv() (*PolicerDetails, error)
	api.Stream
}

type serviceClient_PolicerDumpClient struct {
	api.Stream
}

func (c *serviceClient_PolicerDumpClient) Recv() (*PolicerDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *PolicerDetails:
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

func (c *serviceClient) PolicerInput(ctx context.Context, in *PolicerInput) (*PolicerInputReply, error) {
	out := new(PolicerInputReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
