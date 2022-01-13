// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package ip_neighbor

import (
	"context"
	"fmt"
	"io"
	"mocknet/binapi/vpp2106/vpe"

	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service ip_neighbor.
type RPCService interface {
	IPNeighborAddDel(ctx context.Context, in *IPNeighborAddDel) (*IPNeighborAddDelReply, error)
	IPNeighborConfig(ctx context.Context, in *IPNeighborConfig) (*IPNeighborConfigReply, error)
	IPNeighborDump(ctx context.Context, in *IPNeighborDump) (RPCService_IPNeighborDumpClient, error)
	IPNeighborFlush(ctx context.Context, in *IPNeighborFlush) (*IPNeighborFlushReply, error)
	IPNeighborReplaceBegin(ctx context.Context, in *IPNeighborReplaceBegin) (*IPNeighborReplaceBeginReply, error)
	IPNeighborReplaceEnd(ctx context.Context, in *IPNeighborReplaceEnd) (*IPNeighborReplaceEndReply, error)
	WantIPNeighborEvents(ctx context.Context, in *WantIPNeighborEvents) (*WantIPNeighborEventsReply, error)
	WantIPNeighborEventsV2(ctx context.Context, in *WantIPNeighborEventsV2) (*WantIPNeighborEventsV2Reply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) IPNeighborAddDel(ctx context.Context, in *IPNeighborAddDel) (*IPNeighborAddDelReply, error) {
	out := new(IPNeighborAddDelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) IPNeighborConfig(ctx context.Context, in *IPNeighborConfig) (*IPNeighborConfigReply, error) {
	out := new(IPNeighborConfigReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) IPNeighborDump(ctx context.Context, in *IPNeighborDump) (RPCService_IPNeighborDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_IPNeighborDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_IPNeighborDumpClient interface {
	Recv() (*IPNeighborDetails, error)
	api.Stream
}

type serviceClient_IPNeighborDumpClient struct {
	api.Stream
}

func (c *serviceClient_IPNeighborDumpClient) Recv() (*IPNeighborDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *IPNeighborDetails:
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

func (c *serviceClient) IPNeighborFlush(ctx context.Context, in *IPNeighborFlush) (*IPNeighborFlushReply, error) {
	out := new(IPNeighborFlushReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) IPNeighborReplaceBegin(ctx context.Context, in *IPNeighborReplaceBegin) (*IPNeighborReplaceBeginReply, error) {
	out := new(IPNeighborReplaceBeginReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) IPNeighborReplaceEnd(ctx context.Context, in *IPNeighborReplaceEnd) (*IPNeighborReplaceEndReply, error) {
	out := new(IPNeighborReplaceEndReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) WantIPNeighborEvents(ctx context.Context, in *WantIPNeighborEvents) (*WantIPNeighborEventsReply, error) {
	out := new(WantIPNeighborEventsReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) WantIPNeighborEventsV2(ctx context.Context, in *WantIPNeighborEventsV2) (*WantIPNeighborEventsV2Reply, error) {
	out := new(WantIPNeighborEventsV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
