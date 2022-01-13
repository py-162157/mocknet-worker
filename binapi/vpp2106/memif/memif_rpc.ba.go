// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package memif

import (
	"context"
	"fmt"
	"io"

	api "git.fd.io/govpp.git/api"
	vpe "mocknet/binapi/vpp2106/vpe"
)

// RPCService defines RPC service memif.
type RPCService interface {
	MemifCreate(ctx context.Context, in *MemifCreate) (*MemifCreateReply, error)
	MemifDelete(ctx context.Context, in *MemifDelete) (*MemifDeleteReply, error)
	MemifDump(ctx context.Context, in *MemifDump) (RPCService_MemifDumpClient, error)
	MemifSocketFilenameAddDel(ctx context.Context, in *MemifSocketFilenameAddDel) (*MemifSocketFilenameAddDelReply, error)
	MemifSocketFilenameDump(ctx context.Context, in *MemifSocketFilenameDump) (RPCService_MemifSocketFilenameDumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) MemifCreate(ctx context.Context, in *MemifCreate) (*MemifCreateReply, error) {
	out := new(MemifCreateReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MemifDelete(ctx context.Context, in *MemifDelete) (*MemifDeleteReply, error) {
	out := new(MemifDeleteReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MemifDump(ctx context.Context, in *MemifDump) (RPCService_MemifDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MemifDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MemifDumpClient interface {
	Recv() (*MemifDetails, error)
	api.Stream
}

type serviceClient_MemifDumpClient struct {
	api.Stream
}

func (c *serviceClient_MemifDumpClient) Recv() (*MemifDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MemifDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) MemifSocketFilenameAddDel(ctx context.Context, in *MemifSocketFilenameAddDel) (*MemifSocketFilenameAddDelReply, error) {
	out := new(MemifSocketFilenameAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) MemifSocketFilenameDump(ctx context.Context, in *MemifSocketFilenameDump) (RPCService_MemifSocketFilenameDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_MemifSocketFilenameDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_MemifSocketFilenameDumpClient interface {
	Recv() (*MemifSocketFilenameDetails, error)
	api.Stream
}

type serviceClient_MemifSocketFilenameDumpClient struct {
	api.Stream
}

func (c *serviceClient_MemifSocketFilenameDumpClient) Recv() (*MemifSocketFilenameDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *MemifSocketFilenameDetails:
		return m, nil
	case *vpe.ControlPingReply:
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}