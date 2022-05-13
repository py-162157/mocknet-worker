// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package vxlan

import (
	"context"
	"fmt"
	"io"
	"mocknet/binapi/vpp2202/memclnt"

	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service vxlan.
type RPCService interface {
	SwInterfaceSetVxlanBypass(ctx context.Context, in *SwInterfaceSetVxlanBypass) (*SwInterfaceSetVxlanBypassReply, error)
	VxlanAddDelTunnel(ctx context.Context, in *VxlanAddDelTunnel) (*VxlanAddDelTunnelReply, error)
	VxlanAddDelTunnelV2(ctx context.Context, in *VxlanAddDelTunnelV2) (*VxlanAddDelTunnelV2Reply, error)
	VxlanAddDelTunnelV3(ctx context.Context, in *VxlanAddDelTunnelV3) (*VxlanAddDelTunnelV3Reply, error)
	VxlanOffloadRx(ctx context.Context, in *VxlanOffloadRx) (*VxlanOffloadRxReply, error)
	VxlanTunnelDump(ctx context.Context, in *VxlanTunnelDump) (RPCService_VxlanTunnelDumpClient, error)
	VxlanTunnelV2Dump(ctx context.Context, in *VxlanTunnelV2Dump) (RPCService_VxlanTunnelV2DumpClient, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) SwInterfaceSetVxlanBypass(ctx context.Context, in *SwInterfaceSetVxlanBypass) (*SwInterfaceSetVxlanBypassReply, error) {
	out := new(SwInterfaceSetVxlanBypassReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) VxlanAddDelTunnel(ctx context.Context, in *VxlanAddDelTunnel) (*VxlanAddDelTunnelReply, error) {
	out := new(VxlanAddDelTunnelReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) VxlanAddDelTunnelV2(ctx context.Context, in *VxlanAddDelTunnelV2) (*VxlanAddDelTunnelV2Reply, error) {
	out := new(VxlanAddDelTunnelV2Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) VxlanAddDelTunnelV3(ctx context.Context, in *VxlanAddDelTunnelV3) (*VxlanAddDelTunnelV3Reply, error) {
	out := new(VxlanAddDelTunnelV3Reply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) VxlanOffloadRx(ctx context.Context, in *VxlanOffloadRx) (*VxlanOffloadRxReply, error) {
	out := new(VxlanOffloadRxReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) VxlanTunnelDump(ctx context.Context, in *VxlanTunnelDump) (RPCService_VxlanTunnelDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_VxlanTunnelDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_VxlanTunnelDumpClient interface {
	Recv() (*VxlanTunnelDetails, error)
	api.Stream
}

type serviceClient_VxlanTunnelDumpClient struct {
	api.Stream
}

func (c *serviceClient_VxlanTunnelDumpClient) Recv() (*VxlanTunnelDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *VxlanTunnelDetails:
		return m, nil
	case *memclnt.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}

func (c *serviceClient) VxlanTunnelV2Dump(ctx context.Context, in *VxlanTunnelV2Dump) (RPCService_VxlanTunnelV2DumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_VxlanTunnelV2DumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&memclnt.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_VxlanTunnelV2DumpClient interface {
	Recv() (*VxlanTunnelV2Details, error)
	api.Stream
}

type serviceClient_VxlanTunnelV2DumpClient struct {
	api.Stream
}

func (c *serviceClient_VxlanTunnelV2DumpClient) Recv() (*VxlanTunnelV2Details, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *VxlanTunnelV2Details:
		return m, nil
	case *memclnt.ControlPingReply:
		err = c.Stream.Close()
		if err != nil {
			return nil, err
		}
		return nil, io.EOF
	default:
		return nil, fmt.Errorf("unexpected message: %T %v", m, m)
	}
}
