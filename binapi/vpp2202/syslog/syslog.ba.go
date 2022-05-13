// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.4.0-dev
//  VPP:              22.02.0-8~g3d8bab088-dirty
// source: /usr/share/vpp/api/core/syslog.api.json

// Package syslog contains generated bindings for API file syslog.api.
//
// Contents:
//   1 enum
//   8 messages
//
package syslog

import (
	"mocknet/binapi/vpp2202/ip_types"
	"strconv"

	api "git.fd.io/govpp.git/api"
	codec "git.fd.io/govpp.git/codec"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the GoVPP api package it is being compiled against.
// A compilation error at this line likely means your copy of the
// GoVPP api package needs to be updated.
const _ = api.GoVppAPIPackageIsVersion2

const (
	APIFile    = "syslog"
	APIVersion = "1.0.0"
	VersionCrc = 0x5ad12a74
)

// SyslogSeverity defines enum 'syslog_severity'.
type SyslogSeverity uint32

const (
	SYSLOG_API_SEVERITY_EMERG  SyslogSeverity = 0
	SYSLOG_API_SEVERITY_ALERT  SyslogSeverity = 1
	SYSLOG_API_SEVERITY_CRIT   SyslogSeverity = 2
	SYSLOG_API_SEVERITY_ERR    SyslogSeverity = 3
	SYSLOG_API_SEVERITY_WARN   SyslogSeverity = 4
	SYSLOG_API_SEVERITY_NOTICE SyslogSeverity = 5
	SYSLOG_API_SEVERITY_INFO   SyslogSeverity = 6
	SYSLOG_API_SEVERITY_DBG    SyslogSeverity = 7
)

var (
	SyslogSeverity_name = map[uint32]string{
		0: "SYSLOG_API_SEVERITY_EMERG",
		1: "SYSLOG_API_SEVERITY_ALERT",
		2: "SYSLOG_API_SEVERITY_CRIT",
		3: "SYSLOG_API_SEVERITY_ERR",
		4: "SYSLOG_API_SEVERITY_WARN",
		5: "SYSLOG_API_SEVERITY_NOTICE",
		6: "SYSLOG_API_SEVERITY_INFO",
		7: "SYSLOG_API_SEVERITY_DBG",
	}
	SyslogSeverity_value = map[string]uint32{
		"SYSLOG_API_SEVERITY_EMERG":  0,
		"SYSLOG_API_SEVERITY_ALERT":  1,
		"SYSLOG_API_SEVERITY_CRIT":   2,
		"SYSLOG_API_SEVERITY_ERR":    3,
		"SYSLOG_API_SEVERITY_WARN":   4,
		"SYSLOG_API_SEVERITY_NOTICE": 5,
		"SYSLOG_API_SEVERITY_INFO":   6,
		"SYSLOG_API_SEVERITY_DBG":    7,
	}
)

func (x SyslogSeverity) String() string {
	s, ok := SyslogSeverity_name[uint32(x)]
	if ok {
		return s
	}
	return "SyslogSeverity(" + strconv.Itoa(int(x)) + ")"
}

// SyslogGetFilter defines message 'syslog_get_filter'.
type SyslogGetFilter struct{}

func (m *SyslogGetFilter) Reset()               { *m = SyslogGetFilter{} }
func (*SyslogGetFilter) GetMessageName() string { return "syslog_get_filter" }
func (*SyslogGetFilter) GetCrcString() string   { return "51077d14" }
func (*SyslogGetFilter) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SyslogGetFilter) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *SyslogGetFilter) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *SyslogGetFilter) Unmarshal(b []byte) error {
	return nil
}

// SyslogGetFilterReply defines message 'syslog_get_filter_reply'.
type SyslogGetFilterReply struct {
	Retval   int32          `binapi:"i32,name=retval" json:"retval,omitempty"`
	Severity SyslogSeverity `binapi:"syslog_severity,name=severity" json:"severity,omitempty"`
}

func (m *SyslogGetFilterReply) Reset()               { *m = SyslogGetFilterReply{} }
func (*SyslogGetFilterReply) GetMessageName() string { return "syslog_get_filter_reply" }
func (*SyslogGetFilterReply) GetCrcString() string   { return "eb1833f8" }
func (*SyslogGetFilterReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SyslogGetFilterReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.Severity
	return size
}
func (m *SyslogGetFilterReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.Severity))
	return buf.Bytes(), nil
}
func (m *SyslogGetFilterReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.Severity = SyslogSeverity(buf.DecodeUint32())
	return nil
}

// SyslogGetSender defines message 'syslog_get_sender'.
type SyslogGetSender struct{}

func (m *SyslogGetSender) Reset()               { *m = SyslogGetSender{} }
func (*SyslogGetSender) GetMessageName() string { return "syslog_get_sender" }
func (*SyslogGetSender) GetCrcString() string   { return "51077d14" }
func (*SyslogGetSender) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SyslogGetSender) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *SyslogGetSender) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *SyslogGetSender) Unmarshal(b []byte) error {
	return nil
}

// SyslogGetSenderReply defines message 'syslog_get_sender_reply'.
type SyslogGetSenderReply struct {
	Retval           int32               `binapi:"i32,name=retval" json:"retval,omitempty"`
	SrcAddress       ip_types.IP4Address `binapi:"ip4_address,name=src_address" json:"src_address,omitempty"`
	CollectorAddress ip_types.IP4Address `binapi:"ip4_address,name=collector_address" json:"collector_address,omitempty"`
	CollectorPort    uint16              `binapi:"u16,name=collector_port" json:"collector_port,omitempty"`
	VrfID            uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	MaxMsgSize       uint32              `binapi:"u32,name=max_msg_size" json:"max_msg_size,omitempty"`
}

func (m *SyslogGetSenderReply) Reset()               { *m = SyslogGetSenderReply{} }
func (*SyslogGetSenderReply) GetMessageName() string { return "syslog_get_sender_reply" }
func (*SyslogGetSenderReply) GetCrcString() string   { return "424cfa4e" }
func (*SyslogGetSenderReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SyslogGetSenderReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.Retval
	size += 1 * 4 // m.SrcAddress
	size += 1 * 4 // m.CollectorAddress
	size += 2     // m.CollectorPort
	size += 4     // m.VrfID
	size += 4     // m.MaxMsgSize
	return size
}
func (m *SyslogGetSenderReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeBytes(m.SrcAddress[:], 4)
	buf.EncodeBytes(m.CollectorAddress[:], 4)
	buf.EncodeUint16(m.CollectorPort)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeUint32(m.MaxMsgSize)
	return buf.Bytes(), nil
}
func (m *SyslogGetSenderReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	copy(m.SrcAddress[:], buf.DecodeBytes(4))
	copy(m.CollectorAddress[:], buf.DecodeBytes(4))
	m.CollectorPort = buf.DecodeUint16()
	m.VrfID = buf.DecodeUint32()
	m.MaxMsgSize = buf.DecodeUint32()
	return nil
}

// SyslogSetFilter defines message 'syslog_set_filter'.
type SyslogSetFilter struct {
	Severity SyslogSeverity `binapi:"syslog_severity,name=severity" json:"severity,omitempty"`
}

func (m *SyslogSetFilter) Reset()               { *m = SyslogSetFilter{} }
func (*SyslogSetFilter) GetMessageName() string { return "syslog_set_filter" }
func (*SyslogSetFilter) GetCrcString() string   { return "571348c3" }
func (*SyslogSetFilter) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SyslogSetFilter) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Severity
	return size
}
func (m *SyslogSetFilter) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.Severity))
	return buf.Bytes(), nil
}
func (m *SyslogSetFilter) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Severity = SyslogSeverity(buf.DecodeUint32())
	return nil
}

// SyslogSetFilterReply defines message 'syslog_set_filter_reply'.
type SyslogSetFilterReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SyslogSetFilterReply) Reset()               { *m = SyslogSetFilterReply{} }
func (*SyslogSetFilterReply) GetMessageName() string { return "syslog_set_filter_reply" }
func (*SyslogSetFilterReply) GetCrcString() string   { return "e8d4e804" }
func (*SyslogSetFilterReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SyslogSetFilterReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SyslogSetFilterReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SyslogSetFilterReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// SyslogSetSender defines message 'syslog_set_sender'.
type SyslogSetSender struct {
	SrcAddress       ip_types.IP4Address `binapi:"ip4_address,name=src_address" json:"src_address,omitempty"`
	CollectorAddress ip_types.IP4Address `binapi:"ip4_address,name=collector_address" json:"collector_address,omitempty"`
	CollectorPort    uint16              `binapi:"u16,name=collector_port,default=514" json:"collector_port,omitempty"`
	VrfID            uint32              `binapi:"u32,name=vrf_id" json:"vrf_id,omitempty"`
	MaxMsgSize       uint32              `binapi:"u32,name=max_msg_size,default=480" json:"max_msg_size,omitempty"`
}

func (m *SyslogSetSender) Reset()               { *m = SyslogSetSender{} }
func (*SyslogSetSender) GetMessageName() string { return "syslog_set_sender" }
func (*SyslogSetSender) GetCrcString() string   { return "b8011d0b" }
func (*SyslogSetSender) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SyslogSetSender) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 1 * 4 // m.SrcAddress
	size += 1 * 4 // m.CollectorAddress
	size += 2     // m.CollectorPort
	size += 4     // m.VrfID
	size += 4     // m.MaxMsgSize
	return size
}
func (m *SyslogSetSender) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeBytes(m.SrcAddress[:], 4)
	buf.EncodeBytes(m.CollectorAddress[:], 4)
	buf.EncodeUint16(m.CollectorPort)
	buf.EncodeUint32(m.VrfID)
	buf.EncodeUint32(m.MaxMsgSize)
	return buf.Bytes(), nil
}
func (m *SyslogSetSender) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	copy(m.SrcAddress[:], buf.DecodeBytes(4))
	copy(m.CollectorAddress[:], buf.DecodeBytes(4))
	m.CollectorPort = buf.DecodeUint16()
	m.VrfID = buf.DecodeUint32()
	m.MaxMsgSize = buf.DecodeUint32()
	return nil
}

// SyslogSetSenderReply defines message 'syslog_set_sender_reply'.
type SyslogSetSenderReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SyslogSetSenderReply) Reset()               { *m = SyslogSetSenderReply{} }
func (*SyslogSetSenderReply) GetMessageName() string { return "syslog_set_sender_reply" }
func (*SyslogSetSenderReply) GetCrcString() string   { return "e8d4e804" }
func (*SyslogSetSenderReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SyslogSetSenderReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SyslogSetSenderReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SyslogSetSenderReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

func init() { file_syslog_binapi_init() }
func file_syslog_binapi_init() {
	api.RegisterMessage((*SyslogGetFilter)(nil), "syslog_get_filter_51077d14")
	api.RegisterMessage((*SyslogGetFilterReply)(nil), "syslog_get_filter_reply_eb1833f8")
	api.RegisterMessage((*SyslogGetSender)(nil), "syslog_get_sender_51077d14")
	api.RegisterMessage((*SyslogGetSenderReply)(nil), "syslog_get_sender_reply_424cfa4e")
	api.RegisterMessage((*SyslogSetFilter)(nil), "syslog_set_filter_571348c3")
	api.RegisterMessage((*SyslogSetFilterReply)(nil), "syslog_set_filter_reply_e8d4e804")
	api.RegisterMessage((*SyslogSetSender)(nil), "syslog_set_sender_b8011d0b")
	api.RegisterMessage((*SyslogSetSenderReply)(nil), "syslog_set_sender_reply_e8d4e804")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*SyslogGetFilter)(nil),
		(*SyslogGetFilterReply)(nil),
		(*SyslogGetSender)(nil),
		(*SyslogGetSenderReply)(nil),
		(*SyslogSetFilter)(nil),
		(*SyslogSetFilterReply)(nil),
		(*SyslogSetSender)(nil),
		(*SyslogSetSenderReply)(nil),
	}
}
