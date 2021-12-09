// Code generated by GoVPP's binapi-generator. DO NOT EDIT.
// versions:
//  binapi-generator: v0.3.5-62-g0a0c03d
//  VPP:              21.10.1-release
// source: /usr/share/vpp/api/core/bond.api.json

// Package bond contains generated bindings for API file bond.api.
//
// Contents:
//   2 enums
//  24 messages
//
package bond

import (
	"mocknet/binapi/vpp2110/ethernet_types"
	"mocknet/binapi/vpp2110/interface_types"
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
	APIFile    = "bond"
	APIVersion = "2.1.0"
	VersionCrc = 0xa03f5330
)

// BondLbAlgo defines enum 'bond_lb_algo'.
type BondLbAlgo uint32

const (
	BOND_API_LB_ALGO_L2  BondLbAlgo = 0
	BOND_API_LB_ALGO_L34 BondLbAlgo = 1
	BOND_API_LB_ALGO_L23 BondLbAlgo = 2
	BOND_API_LB_ALGO_RR  BondLbAlgo = 3
	BOND_API_LB_ALGO_BC  BondLbAlgo = 4
	BOND_API_LB_ALGO_AB  BondLbAlgo = 5
)

var (
	BondLbAlgo_name = map[uint32]string{
		0: "BOND_API_LB_ALGO_L2",
		1: "BOND_API_LB_ALGO_L34",
		2: "BOND_API_LB_ALGO_L23",
		3: "BOND_API_LB_ALGO_RR",
		4: "BOND_API_LB_ALGO_BC",
		5: "BOND_API_LB_ALGO_AB",
	}
	BondLbAlgo_value = map[string]uint32{
		"BOND_API_LB_ALGO_L2":  0,
		"BOND_API_LB_ALGO_L34": 1,
		"BOND_API_LB_ALGO_L23": 2,
		"BOND_API_LB_ALGO_RR":  3,
		"BOND_API_LB_ALGO_BC":  4,
		"BOND_API_LB_ALGO_AB":  5,
	}
)

func (x BondLbAlgo) String() string {
	s, ok := BondLbAlgo_name[uint32(x)]
	if ok {
		return s
	}
	return "BondLbAlgo(" + strconv.Itoa(int(x)) + ")"
}

// BondMode defines enum 'bond_mode'.
type BondMode uint32

const (
	BOND_API_MODE_ROUND_ROBIN   BondMode = 1
	BOND_API_MODE_ACTIVE_BACKUP BondMode = 2
	BOND_API_MODE_XOR           BondMode = 3
	BOND_API_MODE_BROADCAST     BondMode = 4
	BOND_API_MODE_LACP          BondMode = 5
)

var (
	BondMode_name = map[uint32]string{
		1: "BOND_API_MODE_ROUND_ROBIN",
		2: "BOND_API_MODE_ACTIVE_BACKUP",
		3: "BOND_API_MODE_XOR",
		4: "BOND_API_MODE_BROADCAST",
		5: "BOND_API_MODE_LACP",
	}
	BondMode_value = map[string]uint32{
		"BOND_API_MODE_ROUND_ROBIN":   1,
		"BOND_API_MODE_ACTIVE_BACKUP": 2,
		"BOND_API_MODE_XOR":           3,
		"BOND_API_MODE_BROADCAST":     4,
		"BOND_API_MODE_LACP":          5,
	}
)

func (x BondMode) String() string {
	s, ok := BondMode_name[uint32(x)]
	if ok {
		return s
	}
	return "BondMode(" + strconv.Itoa(int(x)) + ")"
}

// BondAddMember defines message 'bond_add_member'.
type BondAddMember struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	BondSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=bond_sw_if_index" json:"bond_sw_if_index,omitempty"`
	IsPassive     bool                           `binapi:"bool,name=is_passive" json:"is_passive,omitempty"`
	IsLongTimeout bool                           `binapi:"bool,name=is_long_timeout" json:"is_long_timeout,omitempty"`
}

func (m *BondAddMember) Reset()               { *m = BondAddMember{} }
func (*BondAddMember) GetMessageName() string { return "bond_add_member" }
func (*BondAddMember) GetCrcString() string   { return "e7d14948" }
func (*BondAddMember) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondAddMember) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 4 // m.BondSwIfIndex
	size += 1 // m.IsPassive
	size += 1 // m.IsLongTimeout
	return size
}
func (m *BondAddMember) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(uint32(m.BondSwIfIndex))
	buf.EncodeBool(m.IsPassive)
	buf.EncodeBool(m.IsLongTimeout)
	return buf.Bytes(), nil
}
func (m *BondAddMember) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.BondSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsPassive = buf.DecodeBool()
	m.IsLongTimeout = buf.DecodeBool()
	return nil
}

// BondAddMemberReply defines message 'bond_add_member_reply'.
type BondAddMemberReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *BondAddMemberReply) Reset()               { *m = BondAddMemberReply{} }
func (*BondAddMemberReply) GetMessageName() string { return "bond_add_member_reply" }
func (*BondAddMemberReply) GetCrcString() string   { return "e8d4e804" }
func (*BondAddMemberReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondAddMemberReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *BondAddMemberReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *BondAddMemberReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// BondCreate defines message 'bond_create'.
// Deprecated: the message will be removed in the future versions
type BondCreate struct {
	ID           uint32                    `binapi:"u32,name=id,default=4294967295" json:"id,omitempty"`
	UseCustomMac bool                      `binapi:"bool,name=use_custom_mac" json:"use_custom_mac,omitempty"`
	MacAddress   ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	Mode         BondMode                  `binapi:"bond_mode,name=mode" json:"mode,omitempty"`
	Lb           BondLbAlgo                `binapi:"bond_lb_algo,name=lb" json:"lb,omitempty"`
	NumaOnly     bool                      `binapi:"bool,name=numa_only" json:"numa_only,omitempty"`
}

func (m *BondCreate) Reset()               { *m = BondCreate{} }
func (*BondCreate) GetMessageName() string { return "bond_create" }
func (*BondCreate) GetCrcString() string   { return "f1dbd4ff" }
func (*BondCreate) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondCreate) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.ID
	size += 1     // m.UseCustomMac
	size += 1 * 6 // m.MacAddress
	size += 4     // m.Mode
	size += 4     // m.Lb
	size += 1     // m.NumaOnly
	return size
}
func (m *BondCreate) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(m.ID)
	buf.EncodeBool(m.UseCustomMac)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeUint32(uint32(m.Lb))
	buf.EncodeBool(m.NumaOnly)
	return buf.Bytes(), nil
}
func (m *BondCreate) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.ID = buf.DecodeUint32()
	m.UseCustomMac = buf.DecodeBool()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.Mode = BondMode(buf.DecodeUint32())
	m.Lb = BondLbAlgo(buf.DecodeUint32())
	m.NumaOnly = buf.DecodeBool()
	return nil
}

// BondCreate2 defines message 'bond_create2'.
type BondCreate2 struct {
	Mode         BondMode                  `binapi:"bond_mode,name=mode" json:"mode,omitempty"`
	Lb           BondLbAlgo                `binapi:"bond_lb_algo,name=lb" json:"lb,omitempty"`
	NumaOnly     bool                      `binapi:"bool,name=numa_only" json:"numa_only,omitempty"`
	EnableGso    bool                      `binapi:"bool,name=enable_gso" json:"enable_gso,omitempty"`
	UseCustomMac bool                      `binapi:"bool,name=use_custom_mac" json:"use_custom_mac,omitempty"`
	MacAddress   ethernet_types.MacAddress `binapi:"mac_address,name=mac_address" json:"mac_address,omitempty"`
	ID           uint32                    `binapi:"u32,name=id,default=4294967295" json:"id,omitempty"`
}

func (m *BondCreate2) Reset()               { *m = BondCreate2{} }
func (*BondCreate2) GetMessageName() string { return "bond_create2" }
func (*BondCreate2) GetCrcString() string   { return "912fda76" }
func (*BondCreate2) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondCreate2) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4     // m.Mode
	size += 4     // m.Lb
	size += 1     // m.NumaOnly
	size += 1     // m.EnableGso
	size += 1     // m.UseCustomMac
	size += 1 * 6 // m.MacAddress
	size += 4     // m.ID
	return size
}
func (m *BondCreate2) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeUint32(uint32(m.Lb))
	buf.EncodeBool(m.NumaOnly)
	buf.EncodeBool(m.EnableGso)
	buf.EncodeBool(m.UseCustomMac)
	buf.EncodeBytes(m.MacAddress[:], 6)
	buf.EncodeUint32(m.ID)
	return buf.Bytes(), nil
}
func (m *BondCreate2) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Mode = BondMode(buf.DecodeUint32())
	m.Lb = BondLbAlgo(buf.DecodeUint32())
	m.NumaOnly = buf.DecodeBool()
	m.EnableGso = buf.DecodeBool()
	m.UseCustomMac = buf.DecodeBool()
	copy(m.MacAddress[:], buf.DecodeBytes(6))
	m.ID = buf.DecodeUint32()
	return nil
}

// BondCreate2Reply defines message 'bond_create2_reply'.
type BondCreate2Reply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *BondCreate2Reply) Reset()               { *m = BondCreate2Reply{} }
func (*BondCreate2Reply) GetMessageName() string { return "bond_create2_reply" }
func (*BondCreate2Reply) GetCrcString() string   { return "5383d31f" }
func (*BondCreate2Reply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondCreate2Reply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *BondCreate2Reply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *BondCreate2Reply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// BondCreateReply defines message 'bond_create_reply'.
type BondCreateReply struct {
	Retval    int32                          `binapi:"i32,name=retval" json:"retval,omitempty"`
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *BondCreateReply) Reset()               { *m = BondCreateReply{} }
func (*BondCreateReply) GetMessageName() string { return "bond_create_reply" }
func (*BondCreateReply) GetCrcString() string   { return "5383d31f" }
func (*BondCreateReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondCreateReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	size += 4 // m.SwIfIndex
	return size
}
func (m *BondCreateReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *BondCreateReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// BondDelete defines message 'bond_delete'.
type BondDelete struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *BondDelete) Reset()               { *m = BondDelete{} }
func (*BondDelete) GetMessageName() string { return "bond_delete" }
func (*BondDelete) GetCrcString() string   { return "f9e6675e" }
func (*BondDelete) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondDelete) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *BondDelete) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *BondDelete) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// BondDeleteReply defines message 'bond_delete_reply'.
type BondDeleteReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *BondDeleteReply) Reset()               { *m = BondDeleteReply{} }
func (*BondDeleteReply) GetMessageName() string { return "bond_delete_reply" }
func (*BondDeleteReply) GetCrcString() string   { return "e8d4e804" }
func (*BondDeleteReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondDeleteReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *BondDeleteReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *BondDeleteReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// BondDetachMember defines message 'bond_detach_member'.
type BondDetachMember struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *BondDetachMember) Reset()               { *m = BondDetachMember{} }
func (*BondDetachMember) GetMessageName() string { return "bond_detach_member" }
func (*BondDetachMember) GetCrcString() string   { return "f9e6675e" }
func (*BondDetachMember) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondDetachMember) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *BondDetachMember) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *BondDetachMember) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// BondDetachMemberReply defines message 'bond_detach_member_reply'.
type BondDetachMemberReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *BondDetachMemberReply) Reset()               { *m = BondDetachMemberReply{} }
func (*BondDetachMemberReply) GetMessageName() string { return "bond_detach_member_reply" }
func (*BondDetachMemberReply) GetCrcString() string   { return "e8d4e804" }
func (*BondDetachMemberReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondDetachMemberReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *BondDetachMemberReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *BondDetachMemberReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// BondDetachSlave defines message 'bond_detach_slave'.
// Deprecated: the message will be removed in the future versions
type BondDetachSlave struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *BondDetachSlave) Reset()               { *m = BondDetachSlave{} }
func (*BondDetachSlave) GetMessageName() string { return "bond_detach_slave" }
func (*BondDetachSlave) GetCrcString() string   { return "f9e6675e" }
func (*BondDetachSlave) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondDetachSlave) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *BondDetachSlave) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *BondDetachSlave) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// BondDetachSlaveReply defines message 'bond_detach_slave_reply'.
// Deprecated: the message will be removed in the future versions
type BondDetachSlaveReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *BondDetachSlaveReply) Reset()               { *m = BondDetachSlaveReply{} }
func (*BondDetachSlaveReply) GetMessageName() string { return "bond_detach_slave_reply" }
func (*BondDetachSlaveReply) GetCrcString() string   { return "e8d4e804" }
func (*BondDetachSlaveReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondDetachSlaveReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *BondDetachSlaveReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *BondDetachSlaveReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// BondEnslave defines message 'bond_enslave'.
// Deprecated: the message will be removed in the future versions
type BondEnslave struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	BondSwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=bond_sw_if_index" json:"bond_sw_if_index,omitempty"`
	IsPassive     bool                           `binapi:"bool,name=is_passive" json:"is_passive,omitempty"`
	IsLongTimeout bool                           `binapi:"bool,name=is_long_timeout" json:"is_long_timeout,omitempty"`
}

func (m *BondEnslave) Reset()               { *m = BondEnslave{} }
func (*BondEnslave) GetMessageName() string { return "bond_enslave" }
func (*BondEnslave) GetCrcString() string   { return "e7d14948" }
func (*BondEnslave) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *BondEnslave) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 4 // m.BondSwIfIndex
	size += 1 // m.IsPassive
	size += 1 // m.IsLongTimeout
	return size
}
func (m *BondEnslave) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(uint32(m.BondSwIfIndex))
	buf.EncodeBool(m.IsPassive)
	buf.EncodeBool(m.IsLongTimeout)
	return buf.Bytes(), nil
}
func (m *BondEnslave) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.BondSwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.IsPassive = buf.DecodeBool()
	m.IsLongTimeout = buf.DecodeBool()
	return nil
}

// BondEnslaveReply defines message 'bond_enslave_reply'.
type BondEnslaveReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *BondEnslaveReply) Reset()               { *m = BondEnslaveReply{} }
func (*BondEnslaveReply) GetMessageName() string { return "bond_enslave_reply" }
func (*BondEnslaveReply) GetCrcString() string   { return "e8d4e804" }
func (*BondEnslaveReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *BondEnslaveReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *BondEnslaveReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *BondEnslaveReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// SwBondInterfaceDetails defines message 'sw_bond_interface_details'.
type SwBondInterfaceDetails struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	ID            uint32                         `binapi:"u32,name=id" json:"id,omitempty"`
	Mode          BondMode                       `binapi:"bond_mode,name=mode" json:"mode,omitempty"`
	Lb            BondLbAlgo                     `binapi:"bond_lb_algo,name=lb" json:"lb,omitempty"`
	NumaOnly      bool                           `binapi:"bool,name=numa_only" json:"numa_only,omitempty"`
	ActiveMembers uint32                         `binapi:"u32,name=active_members" json:"active_members,omitempty"`
	Members       uint32                         `binapi:"u32,name=members" json:"members,omitempty"`
	InterfaceName string                         `binapi:"string[64],name=interface_name" json:"interface_name,omitempty"`
}

func (m *SwBondInterfaceDetails) Reset()               { *m = SwBondInterfaceDetails{} }
func (*SwBondInterfaceDetails) GetMessageName() string { return "sw_bond_interface_details" }
func (*SwBondInterfaceDetails) GetCrcString() string   { return "9428a69c" }
func (*SwBondInterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwBondInterfaceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4  // m.SwIfIndex
	size += 4  // m.ID
	size += 4  // m.Mode
	size += 4  // m.Lb
	size += 1  // m.NumaOnly
	size += 4  // m.ActiveMembers
	size += 4  // m.Members
	size += 64 // m.InterfaceName
	return size
}
func (m *SwBondInterfaceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(m.ID)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeUint32(uint32(m.Lb))
	buf.EncodeBool(m.NumaOnly)
	buf.EncodeUint32(m.ActiveMembers)
	buf.EncodeUint32(m.Members)
	buf.EncodeString(m.InterfaceName, 64)
	return buf.Bytes(), nil
}
func (m *SwBondInterfaceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.ID = buf.DecodeUint32()
	m.Mode = BondMode(buf.DecodeUint32())
	m.Lb = BondLbAlgo(buf.DecodeUint32())
	m.NumaOnly = buf.DecodeBool()
	m.ActiveMembers = buf.DecodeUint32()
	m.Members = buf.DecodeUint32()
	m.InterfaceName = buf.DecodeString(64)
	return nil
}

// SwBondInterfaceDump defines message 'sw_bond_interface_dump'.
type SwBondInterfaceDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index,default=4294967295" json:"sw_if_index,omitempty"`
}

func (m *SwBondInterfaceDump) Reset()               { *m = SwBondInterfaceDump{} }
func (*SwBondInterfaceDump) GetMessageName() string { return "sw_bond_interface_dump" }
func (*SwBondInterfaceDump) GetCrcString() string   { return "f9e6675e" }
func (*SwBondInterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwBondInterfaceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *SwBondInterfaceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *SwBondInterfaceDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// SwInterfaceBondDetails defines message 'sw_interface_bond_details'.
type SwInterfaceBondDetails struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	ID            uint32                         `binapi:"u32,name=id" json:"id,omitempty"`
	Mode          BondMode                       `binapi:"bond_mode,name=mode" json:"mode,omitempty"`
	Lb            BondLbAlgo                     `binapi:"bond_lb_algo,name=lb" json:"lb,omitempty"`
	NumaOnly      bool                           `binapi:"bool,name=numa_only" json:"numa_only,omitempty"`
	ActiveSlaves  uint32                         `binapi:"u32,name=active_slaves" json:"active_slaves,omitempty"`
	Slaves        uint32                         `binapi:"u32,name=slaves" json:"slaves,omitempty"`
	InterfaceName string                         `binapi:"string[64],name=interface_name" json:"interface_name,omitempty"`
}

func (m *SwInterfaceBondDetails) Reset()               { *m = SwInterfaceBondDetails{} }
func (*SwInterfaceBondDetails) GetMessageName() string { return "sw_interface_bond_details" }
func (*SwInterfaceBondDetails) GetCrcString() string   { return "bb7c929b" }
func (*SwInterfaceBondDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceBondDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4  // m.SwIfIndex
	size += 4  // m.ID
	size += 4  // m.Mode
	size += 4  // m.Lb
	size += 1  // m.NumaOnly
	size += 4  // m.ActiveSlaves
	size += 4  // m.Slaves
	size += 64 // m.InterfaceName
	return size
}
func (m *SwInterfaceBondDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(m.ID)
	buf.EncodeUint32(uint32(m.Mode))
	buf.EncodeUint32(uint32(m.Lb))
	buf.EncodeBool(m.NumaOnly)
	buf.EncodeUint32(m.ActiveSlaves)
	buf.EncodeUint32(m.Slaves)
	buf.EncodeString(m.InterfaceName, 64)
	return buf.Bytes(), nil
}
func (m *SwInterfaceBondDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.ID = buf.DecodeUint32()
	m.Mode = BondMode(buf.DecodeUint32())
	m.Lb = BondLbAlgo(buf.DecodeUint32())
	m.NumaOnly = buf.DecodeBool()
	m.ActiveSlaves = buf.DecodeUint32()
	m.Slaves = buf.DecodeUint32()
	m.InterfaceName = buf.DecodeString(64)
	return nil
}

// SwInterfaceBondDump defines message 'sw_interface_bond_dump'.
// Deprecated: the message will be removed in the future versions
type SwInterfaceBondDump struct{}

func (m *SwInterfaceBondDump) Reset()               { *m = SwInterfaceBondDump{} }
func (*SwInterfaceBondDump) GetMessageName() string { return "sw_interface_bond_dump" }
func (*SwInterfaceBondDump) GetCrcString() string   { return "51077d14" }
func (*SwInterfaceBondDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceBondDump) Size() (size int) {
	if m == nil {
		return 0
	}
	return size
}
func (m *SwInterfaceBondDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	return buf.Bytes(), nil
}
func (m *SwInterfaceBondDump) Unmarshal(b []byte) error {
	return nil
}

// SwInterfaceSetBondWeight defines message 'sw_interface_set_bond_weight'.
type SwInterfaceSetBondWeight struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	Weight    uint32                         `binapi:"u32,name=weight" json:"weight,omitempty"`
}

func (m *SwInterfaceSetBondWeight) Reset()               { *m = SwInterfaceSetBondWeight{} }
func (*SwInterfaceSetBondWeight) GetMessageName() string { return "sw_interface_set_bond_weight" }
func (*SwInterfaceSetBondWeight) GetCrcString() string   { return "deb510a0" }
func (*SwInterfaceSetBondWeight) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceSetBondWeight) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	size += 4 // m.Weight
	return size
}
func (m *SwInterfaceSetBondWeight) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeUint32(m.Weight)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetBondWeight) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.Weight = buf.DecodeUint32()
	return nil
}

// SwInterfaceSetBondWeightReply defines message 'sw_interface_set_bond_weight_reply'.
type SwInterfaceSetBondWeightReply struct {
	Retval int32 `binapi:"i32,name=retval" json:"retval,omitempty"`
}

func (m *SwInterfaceSetBondWeightReply) Reset() { *m = SwInterfaceSetBondWeightReply{} }
func (*SwInterfaceSetBondWeightReply) GetMessageName() string {
	return "sw_interface_set_bond_weight_reply"
}
func (*SwInterfaceSetBondWeightReply) GetCrcString() string { return "e8d4e804" }
func (*SwInterfaceSetBondWeightReply) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceSetBondWeightReply) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.Retval
	return size
}
func (m *SwInterfaceSetBondWeightReply) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeInt32(m.Retval)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSetBondWeightReply) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.Retval = buf.DecodeInt32()
	return nil
}

// SwInterfaceSlaveDetails defines message 'sw_interface_slave_details'.
type SwInterfaceSlaveDetails struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	InterfaceName string                         `binapi:"string[64],name=interface_name" json:"interface_name,omitempty"`
	IsPassive     bool                           `binapi:"bool,name=is_passive" json:"is_passive,omitempty"`
	IsLongTimeout bool                           `binapi:"bool,name=is_long_timeout" json:"is_long_timeout,omitempty"`
	IsLocalNuma   bool                           `binapi:"bool,name=is_local_numa" json:"is_local_numa,omitempty"`
	Weight        uint32                         `binapi:"u32,name=weight" json:"weight,omitempty"`
}

func (m *SwInterfaceSlaveDetails) Reset()               { *m = SwInterfaceSlaveDetails{} }
func (*SwInterfaceSlaveDetails) GetMessageName() string { return "sw_interface_slave_details" }
func (*SwInterfaceSlaveDetails) GetCrcString() string   { return "3c4a0e23" }
func (*SwInterfaceSlaveDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwInterfaceSlaveDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4  // m.SwIfIndex
	size += 64 // m.InterfaceName
	size += 1  // m.IsPassive
	size += 1  // m.IsLongTimeout
	size += 1  // m.IsLocalNuma
	size += 4  // m.Weight
	return size
}
func (m *SwInterfaceSlaveDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeString(m.InterfaceName, 64)
	buf.EncodeBool(m.IsPassive)
	buf.EncodeBool(m.IsLongTimeout)
	buf.EncodeBool(m.IsLocalNuma)
	buf.EncodeUint32(m.Weight)
	return buf.Bytes(), nil
}
func (m *SwInterfaceSlaveDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.InterfaceName = buf.DecodeString(64)
	m.IsPassive = buf.DecodeBool()
	m.IsLongTimeout = buf.DecodeBool()
	m.IsLocalNuma = buf.DecodeBool()
	m.Weight = buf.DecodeUint32()
	return nil
}

// SwInterfaceSlaveDump defines message 'sw_interface_slave_dump'.
// Deprecated: the message will be removed in the future versions
type SwInterfaceSlaveDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *SwInterfaceSlaveDump) Reset()               { *m = SwInterfaceSlaveDump{} }
func (*SwInterfaceSlaveDump) GetMessageName() string { return "sw_interface_slave_dump" }
func (*SwInterfaceSlaveDump) GetCrcString() string   { return "f9e6675e" }
func (*SwInterfaceSlaveDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwInterfaceSlaveDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *SwInterfaceSlaveDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *SwInterfaceSlaveDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

// SwMemberInterfaceDetails defines message 'sw_member_interface_details'.
type SwMemberInterfaceDetails struct {
	SwIfIndex     interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
	InterfaceName string                         `binapi:"string[64],name=interface_name" json:"interface_name,omitempty"`
	IsPassive     bool                           `binapi:"bool,name=is_passive" json:"is_passive,omitempty"`
	IsLongTimeout bool                           `binapi:"bool,name=is_long_timeout" json:"is_long_timeout,omitempty"`
	IsLocalNuma   bool                           `binapi:"bool,name=is_local_numa" json:"is_local_numa,omitempty"`
	Weight        uint32                         `binapi:"u32,name=weight" json:"weight,omitempty"`
}

func (m *SwMemberInterfaceDetails) Reset()               { *m = SwMemberInterfaceDetails{} }
func (*SwMemberInterfaceDetails) GetMessageName() string { return "sw_member_interface_details" }
func (*SwMemberInterfaceDetails) GetCrcString() string   { return "3c4a0e23" }
func (*SwMemberInterfaceDetails) GetMessageType() api.MessageType {
	return api.ReplyMessage
}

func (m *SwMemberInterfaceDetails) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4  // m.SwIfIndex
	size += 64 // m.InterfaceName
	size += 1  // m.IsPassive
	size += 1  // m.IsLongTimeout
	size += 1  // m.IsLocalNuma
	size += 4  // m.Weight
	return size
}
func (m *SwMemberInterfaceDetails) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	buf.EncodeString(m.InterfaceName, 64)
	buf.EncodeBool(m.IsPassive)
	buf.EncodeBool(m.IsLongTimeout)
	buf.EncodeBool(m.IsLocalNuma)
	buf.EncodeUint32(m.Weight)
	return buf.Bytes(), nil
}
func (m *SwMemberInterfaceDetails) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	m.InterfaceName = buf.DecodeString(64)
	m.IsPassive = buf.DecodeBool()
	m.IsLongTimeout = buf.DecodeBool()
	m.IsLocalNuma = buf.DecodeBool()
	m.Weight = buf.DecodeUint32()
	return nil
}

// SwMemberInterfaceDump defines message 'sw_member_interface_dump'.
type SwMemberInterfaceDump struct {
	SwIfIndex interface_types.InterfaceIndex `binapi:"interface_index,name=sw_if_index" json:"sw_if_index,omitempty"`
}

func (m *SwMemberInterfaceDump) Reset()               { *m = SwMemberInterfaceDump{} }
func (*SwMemberInterfaceDump) GetMessageName() string { return "sw_member_interface_dump" }
func (*SwMemberInterfaceDump) GetCrcString() string   { return "f9e6675e" }
func (*SwMemberInterfaceDump) GetMessageType() api.MessageType {
	return api.RequestMessage
}

func (m *SwMemberInterfaceDump) Size() (size int) {
	if m == nil {
		return 0
	}
	size += 4 // m.SwIfIndex
	return size
}
func (m *SwMemberInterfaceDump) Marshal(b []byte) ([]byte, error) {
	if b == nil {
		b = make([]byte, m.Size())
	}
	buf := codec.NewBuffer(b)
	buf.EncodeUint32(uint32(m.SwIfIndex))
	return buf.Bytes(), nil
}
func (m *SwMemberInterfaceDump) Unmarshal(b []byte) error {
	buf := codec.NewBuffer(b)
	m.SwIfIndex = interface_types.InterfaceIndex(buf.DecodeUint32())
	return nil
}

func init() { file_bond_binapi_init() }
func file_bond_binapi_init() {
	api.RegisterMessage((*BondAddMember)(nil), "bond_add_member_e7d14948")
	api.RegisterMessage((*BondAddMemberReply)(nil), "bond_add_member_reply_e8d4e804")
	api.RegisterMessage((*BondCreate)(nil), "bond_create_f1dbd4ff")
	api.RegisterMessage((*BondCreate2)(nil), "bond_create2_912fda76")
	api.RegisterMessage((*BondCreate2Reply)(nil), "bond_create2_reply_5383d31f")
	api.RegisterMessage((*BondCreateReply)(nil), "bond_create_reply_5383d31f")
	api.RegisterMessage((*BondDelete)(nil), "bond_delete_f9e6675e")
	api.RegisterMessage((*BondDeleteReply)(nil), "bond_delete_reply_e8d4e804")
	api.RegisterMessage((*BondDetachMember)(nil), "bond_detach_member_f9e6675e")
	api.RegisterMessage((*BondDetachMemberReply)(nil), "bond_detach_member_reply_e8d4e804")
	api.RegisterMessage((*BondDetachSlave)(nil), "bond_detach_slave_f9e6675e")
	api.RegisterMessage((*BondDetachSlaveReply)(nil), "bond_detach_slave_reply_e8d4e804")
	api.RegisterMessage((*BondEnslave)(nil), "bond_enslave_e7d14948")
	api.RegisterMessage((*BondEnslaveReply)(nil), "bond_enslave_reply_e8d4e804")
	api.RegisterMessage((*SwBondInterfaceDetails)(nil), "sw_bond_interface_details_9428a69c")
	api.RegisterMessage((*SwBondInterfaceDump)(nil), "sw_bond_interface_dump_f9e6675e")
	api.RegisterMessage((*SwInterfaceBondDetails)(nil), "sw_interface_bond_details_bb7c929b")
	api.RegisterMessage((*SwInterfaceBondDump)(nil), "sw_interface_bond_dump_51077d14")
	api.RegisterMessage((*SwInterfaceSetBondWeight)(nil), "sw_interface_set_bond_weight_deb510a0")
	api.RegisterMessage((*SwInterfaceSetBondWeightReply)(nil), "sw_interface_set_bond_weight_reply_e8d4e804")
	api.RegisterMessage((*SwInterfaceSlaveDetails)(nil), "sw_interface_slave_details_3c4a0e23")
	api.RegisterMessage((*SwInterfaceSlaveDump)(nil), "sw_interface_slave_dump_f9e6675e")
	api.RegisterMessage((*SwMemberInterfaceDetails)(nil), "sw_member_interface_details_3c4a0e23")
	api.RegisterMessage((*SwMemberInterfaceDump)(nil), "sw_member_interface_dump_f9e6675e")
}

// Messages returns list of all messages in this module.
func AllMessages() []api.Message {
	return []api.Message{
		(*BondAddMember)(nil),
		(*BondAddMemberReply)(nil),
		(*BondCreate)(nil),
		(*BondCreate2)(nil),
		(*BondCreate2Reply)(nil),
		(*BondCreateReply)(nil),
		(*BondDelete)(nil),
		(*BondDeleteReply)(nil),
		(*BondDetachMember)(nil),
		(*BondDetachMemberReply)(nil),
		(*BondDetachSlave)(nil),
		(*BondDetachSlaveReply)(nil),
		(*BondEnslave)(nil),
		(*BondEnslaveReply)(nil),
		(*SwBondInterfaceDetails)(nil),
		(*SwBondInterfaceDump)(nil),
		(*SwInterfaceBondDetails)(nil),
		(*SwInterfaceBondDump)(nil),
		(*SwInterfaceSetBondWeight)(nil),
		(*SwInterfaceSetBondWeightReply)(nil),
		(*SwInterfaceSlaveDetails)(nil),
		(*SwInterfaceSlaveDump)(nil),
		(*SwMemberInterfaceDetails)(nil),
		(*SwMemberInterfaceDump)(nil),
	}
}
