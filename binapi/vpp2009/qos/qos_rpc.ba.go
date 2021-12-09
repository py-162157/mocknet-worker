// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package qos

import (
	"context"
	"fmt"
	"io"
	"mocknet/binapi/vpp2009/vpe"

	api "git.fd.io/govpp.git/api"
)

// RPCService defines RPC service qos.
type RPCService interface {
	QosEgressMapDelete(ctx context.Context, in *QosEgressMapDelete) (*QosEgressMapDeleteReply, error)
	QosEgressMapDump(ctx context.Context, in *QosEgressMapDump) (RPCService_QosEgressMapDumpClient, error)
	QosEgressMapUpdate(ctx context.Context, in *QosEgressMapUpdate) (*QosEgressMapUpdateReply, error)
	QosMarkDump(ctx context.Context, in *QosMarkDump) (RPCService_QosMarkDumpClient, error)
	QosMarkEnableDisable(ctx context.Context, in *QosMarkEnableDisable) (*QosMarkEnableDisableReply, error)
	QosRecordDump(ctx context.Context, in *QosRecordDump) (RPCService_QosRecordDumpClient, error)
	QosRecordEnableDisable(ctx context.Context, in *QosRecordEnableDisable) (*QosRecordEnableDisableReply, error)
	QosStoreDump(ctx context.Context, in *QosStoreDump) (RPCService_QosStoreDumpClient, error)
	QosStoreEnableDisable(ctx context.Context, in *QosStoreEnableDisable) (*QosStoreEnableDisableReply, error)
}

type serviceClient struct {
	conn api.Connection
}

func NewServiceClient(conn api.Connection) RPCService {
	return &serviceClient{conn}
}

func (c *serviceClient) QosEgressMapDelete(ctx context.Context, in *QosEgressMapDelete) (*QosEgressMapDeleteReply, error) {
	out := new(QosEgressMapDeleteReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) QosEgressMapDump(ctx context.Context, in *QosEgressMapDump) (RPCService_QosEgressMapDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_QosEgressMapDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_QosEgressMapDumpClient interface {
	Recv() (*QosEgressMapDetails, error)
	api.Stream
}

type serviceClient_QosEgressMapDumpClient struct {
	api.Stream
}

func (c *serviceClient_QosEgressMapDumpClient) Recv() (*QosEgressMapDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *QosEgressMapDetails:
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

func (c *serviceClient) QosEgressMapUpdate(ctx context.Context, in *QosEgressMapUpdate) (*QosEgressMapUpdateReply, error) {
	out := new(QosEgressMapUpdateReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) QosMarkDump(ctx context.Context, in *QosMarkDump) (RPCService_QosMarkDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_QosMarkDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_QosMarkDumpClient interface {
	Recv() (*QosMarkDetails, error)
	api.Stream
}

type serviceClient_QosMarkDumpClient struct {
	api.Stream
}

func (c *serviceClient_QosMarkDumpClient) Recv() (*QosMarkDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *QosMarkDetails:
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

func (c *serviceClient) QosMarkEnableDisable(ctx context.Context, in *QosMarkEnableDisable) (*QosMarkEnableDisableReply, error) {
	out := new(QosMarkEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) QosRecordDump(ctx context.Context, in *QosRecordDump) (RPCService_QosRecordDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_QosRecordDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_QosRecordDumpClient interface {
	Recv() (*QosRecordDetails, error)
	api.Stream
}

type serviceClient_QosRecordDumpClient struct {
	api.Stream
}

func (c *serviceClient_QosRecordDumpClient) Recv() (*QosRecordDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *QosRecordDetails:
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

func (c *serviceClient) QosRecordEnableDisable(ctx context.Context, in *QosRecordEnableDisable) (*QosRecordEnableDisableReply, error) {
	out := new(QosRecordEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}

func (c *serviceClient) QosStoreDump(ctx context.Context, in *QosStoreDump) (RPCService_QosStoreDumpClient, error) {
	stream, err := c.conn.NewStream(ctx)
	if err != nil {
		return nil, err
	}
	x := &serviceClient_QosStoreDumpClient{stream}
	if err := x.Stream.SendMsg(in); err != nil {
		return nil, err
	}
	if err = x.Stream.SendMsg(&vpe.ControlPing{}); err != nil {
		return nil, err
	}
	return x, nil
}

type RPCService_QosStoreDumpClient interface {
	Recv() (*QosStoreDetails, error)
	api.Stream
}

type serviceClient_QosStoreDumpClient struct {
	api.Stream
}

func (c *serviceClient_QosStoreDumpClient) Recv() (*QosStoreDetails, error) {
	msg, err := c.Stream.RecvMsg()
	if err != nil {
		return nil, err
	}
	switch m := msg.(type) {
	case *QosStoreDetails:
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

func (c *serviceClient) QosStoreEnableDisable(ctx context.Context, in *QosStoreEnableDisable) (*QosStoreEnableDisableReply, error) {
	out := new(QosStoreEnableDisableReply)
	err := c.conn.Invoke(ctx, in, out)
	if err != nil {
		return nil, err
	}
	return out, api.RetvalToVPPApiError(out.Retval)
}
