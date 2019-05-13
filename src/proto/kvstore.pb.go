// Code generated by protoc-gen-go. DO NOT EDIT.
// source: kvstore.proto

package proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// You'll likely need to define more specific return codes than these!
type ReturnCode int32

const (
	ReturnCode_SUCCESS                       ReturnCode = 0
	ReturnCode_FAILURE_GET_NOKEY             ReturnCode = 1
	ReturnCode_FAILURE_GET_NOTLEADER         ReturnCode = 2
	ReturnCode_FAILURE_PUT                   ReturnCode = 3
	ReturnCode_FAILURE_PUT_CANTPARSECLIENTIP ReturnCode = 4
)

var ReturnCode_name = map[int32]string{
	0: "SUCCESS",
	1: "FAILURE_GET_NOKEY",
	2: "FAILURE_GET_NOTLEADER",
	3: "FAILURE_PUT",
	4: "FAILURE_PUT_CANTPARSECLIENTIP",
}

var ReturnCode_value = map[string]int32{
	"SUCCESS":                       0,
	"FAILURE_GET_NOKEY":             1,
	"FAILURE_GET_NOTLEADER":         2,
	"FAILURE_PUT":                   3,
	"FAILURE_PUT_CANTPARSECLIENTIP": 4,
}

func (x ReturnCode) String() string {
	return proto.EnumName(ReturnCode_name, int32(x))
}

func (ReturnCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{0}
}

type ExitRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ExitRequest) Reset()         { *m = ExitRequest{} }
func (m *ExitRequest) String() string { return proto.CompactTextString(m) }
func (*ExitRequest) ProtoMessage()    {}
func (*ExitRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{0}
}

func (m *ExitRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitRequest.Unmarshal(m, b)
}
func (m *ExitRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitRequest.Marshal(b, m, deterministic)
}
func (m *ExitRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitRequest.Merge(m, src)
}
func (m *ExitRequest) XXX_Size() int {
	return xxx_messageInfo_ExitRequest.Size(m)
}
func (m *ExitRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ExitRequest proto.InternalMessageInfo

type ExitResponse struct {
	Ret                  ReturnCode `protobuf:"varint,1,opt,name=ret,proto3,enum=proto.ReturnCode" json:"ret,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *ExitResponse) Reset()         { *m = ExitResponse{} }
func (m *ExitResponse) String() string { return proto.CompactTextString(m) }
func (*ExitResponse) ProtoMessage()    {}
func (*ExitResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{1}
}

func (m *ExitResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ExitResponse.Unmarshal(m, b)
}
func (m *ExitResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ExitResponse.Marshal(b, m, deterministic)
}
func (m *ExitResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExitResponse.Merge(m, src)
}
func (m *ExitResponse) XXX_Size() int {
	return xxx_messageInfo_ExitResponse.Size(m)
}
func (m *ExitResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ExitResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ExitResponse proto.InternalMessageInfo

func (m *ExitResponse) GetRet() ReturnCode {
	if m != nil {
		return m.Ret
	}
	return ReturnCode_SUCCESS
}

type GetRequest struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{2}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type GetResponse struct {
	Value                string     `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	Ret                  ReturnCode `protobuf:"varint,2,opt,name=ret,proto3,enum=proto.ReturnCode" json:"ret,omitempty"`
	LeaderID             int32      `protobuf:"varint,3,opt,name=leaderID,proto3" json:"leaderID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *GetResponse) Reset()         { *m = GetResponse{} }
func (m *GetResponse) String() string { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()    {}
func (*GetResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{3}
}

func (m *GetResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetResponse.Unmarshal(m, b)
}
func (m *GetResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetResponse.Marshal(b, m, deterministic)
}
func (m *GetResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetResponse.Merge(m, src)
}
func (m *GetResponse) XXX_Size() int {
	return xxx_messageInfo_GetResponse.Size(m)
}
func (m *GetResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetResponse proto.InternalMessageInfo

func (m *GetResponse) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *GetResponse) GetRet() ReturnCode {
	if m != nil {
		return m.Ret
	}
	return ReturnCode_SUCCESS
}

func (m *GetResponse) GetLeaderID() int32 {
	if m != nil {
		return m.LeaderID
	}
	return 0
}

type PutRequest struct {
	Key                  string   `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	SequenceNo           int64    `protobuf:"varint,3,opt,name=sequenceNo,proto3" json:"sequenceNo,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutRequest) Reset()         { *m = PutRequest{} }
func (m *PutRequest) String() string { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()    {}
func (*PutRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{4}
}

func (m *PutRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutRequest.Unmarshal(m, b)
}
func (m *PutRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutRequest.Marshal(b, m, deterministic)
}
func (m *PutRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutRequest.Merge(m, src)
}
func (m *PutRequest) XXX_Size() int {
	return xxx_messageInfo_PutRequest.Size(m)
}
func (m *PutRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PutRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PutRequest proto.InternalMessageInfo

func (m *PutRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *PutRequest) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *PutRequest) GetSequenceNo() int64 {
	if m != nil {
		return m.SequenceNo
	}
	return 0
}

type PutResponse struct {
	Ret                  ReturnCode `protobuf:"varint,1,opt,name=ret,proto3,enum=proto.ReturnCode" json:"ret,omitempty"`
	LeaderID             int32      `protobuf:"varint,2,opt,name=leaderID,proto3" json:"leaderID,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *PutResponse) Reset()         { *m = PutResponse{} }
func (m *PutResponse) String() string { return proto.CompactTextString(m) }
func (*PutResponse) ProtoMessage()    {}
func (*PutResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{5}
}

func (m *PutResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutResponse.Unmarshal(m, b)
}
func (m *PutResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutResponse.Marshal(b, m, deterministic)
}
func (m *PutResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutResponse.Merge(m, src)
}
func (m *PutResponse) XXX_Size() int {
	return xxx_messageInfo_PutResponse.Size(m)
}
func (m *PutResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PutResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PutResponse proto.InternalMessageInfo

func (m *PutResponse) GetRet() ReturnCode {
	if m != nil {
		return m.Ret
	}
	return ReturnCode_SUCCESS
}

func (m *PutResponse) GetLeaderID() int32 {
	if m != nil {
		return m.LeaderID
	}
	return 0
}

type CLRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CLRequest) Reset()         { *m = CLRequest{} }
func (m *CLRequest) String() string { return proto.CompactTextString(m) }
func (*CLRequest) ProtoMessage()    {}
func (*CLRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{6}
}

func (m *CLRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CLRequest.Unmarshal(m, b)
}
func (m *CLRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CLRequest.Marshal(b, m, deterministic)
}
func (m *CLRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CLRequest.Merge(m, src)
}
func (m *CLRequest) XXX_Size() int {
	return xxx_messageInfo_CLRequest.Size(m)
}
func (m *CLRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CLRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CLRequest proto.InternalMessageInfo

type CLResponse struct {
	IsLeader             bool     `protobuf:"varint,1,opt,name=isLeader,proto3" json:"isLeader,omitempty"`
	LeaderId             int32    `protobuf:"varint,2,opt,name=leaderId,proto3" json:"leaderId,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CLResponse) Reset()         { *m = CLResponse{} }
func (m *CLResponse) String() string { return proto.CompactTextString(m) }
func (*CLResponse) ProtoMessage()    {}
func (*CLResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_088d7f6aff848d9e, []int{7}
}

func (m *CLResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CLResponse.Unmarshal(m, b)
}
func (m *CLResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CLResponse.Marshal(b, m, deterministic)
}
func (m *CLResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CLResponse.Merge(m, src)
}
func (m *CLResponse) XXX_Size() int {
	return xxx_messageInfo_CLResponse.Size(m)
}
func (m *CLResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_CLResponse.DiscardUnknown(m)
}

var xxx_messageInfo_CLResponse proto.InternalMessageInfo

func (m *CLResponse) GetIsLeader() bool {
	if m != nil {
		return m.IsLeader
	}
	return false
}

func (m *CLResponse) GetLeaderId() int32 {
	if m != nil {
		return m.LeaderId
	}
	return 0
}

func init() {
	proto.RegisterEnum("proto.ReturnCode", ReturnCode_name, ReturnCode_value)
	proto.RegisterType((*ExitRequest)(nil), "proto.ExitRequest")
	proto.RegisterType((*ExitResponse)(nil), "proto.ExitResponse")
	proto.RegisterType((*GetRequest)(nil), "proto.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "proto.GetResponse")
	proto.RegisterType((*PutRequest)(nil), "proto.PutRequest")
	proto.RegisterType((*PutResponse)(nil), "proto.PutResponse")
	proto.RegisterType((*CLRequest)(nil), "proto.CLRequest")
	proto.RegisterType((*CLResponse)(nil), "proto.CLResponse")
}

func init() { proto.RegisterFile("kvstore.proto", fileDescriptor_088d7f6aff848d9e) }

var fileDescriptor_088d7f6aff848d9e = []byte{
	// 408 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x52, 0x4d, 0xaf, 0xd2, 0x40,
	0x14, 0xa5, 0xed, 0x43, 0xfb, 0x6e, 0x45, 0xcb, 0xe8, 0x4b, 0x9e, 0x4d, 0x24, 0x38, 0x6e, 0x88,
	0x0b, 0x12, 0xe1, 0x17, 0x34, 0x65, 0x24, 0x0d, 0x4d, 0x6d, 0xa6, 0xc5, 0xc4, 0x15, 0x41, 0x3b,
	0x0b, 0x02, 0xa1, 0xd8, 0x0f, 0x22, 0x2b, 0x7f, 0xa5, 0xff, 0xc7, 0xcc, 0xb4, 0xd3, 0x0e, 0x24,
	0x9a, 0xbc, 0x15, 0xdc, 0x73, 0xcf, 0x3d, 0xe7, 0xf6, 0xcc, 0x85, 0xc1, 0xfe, 0x5c, 0x94, 0x59,
	0xce, 0xa6, 0xa7, 0x3c, 0x2b, 0x33, 0xd4, 0x17, 0x3f, 0x78, 0x00, 0x16, 0xf9, 0xb5, 0x2b, 0x29,
	0xfb, 0x59, 0xb1, 0xa2, 0xc4, 0x73, 0x78, 0x51, 0x97, 0xc5, 0x29, 0x3b, 0x16, 0x0c, 0x7d, 0x00,
	0x23, 0x67, 0xe5, 0xa3, 0x36, 0xd6, 0x26, 0x2f, 0x67, 0xc3, 0x7a, 0x74, 0x4a, 0x59, 0x59, 0xe5,
	0x47, 0x2f, 0x4b, 0x19, 0xe5, 0x5d, 0x3c, 0x02, 0x58, 0x32, 0x29, 0x81, 0x6c, 0x30, 0xf6, 0xec,
	0x22, 0x46, 0xee, 0x29, 0xff, 0x8b, 0x53, 0xb0, 0x44, 0xbf, 0xd1, 0x7c, 0x03, 0xfd, 0xf3, 0xf6,
	0x50, 0xb1, 0x86, 0x52, 0x17, 0xd2, 0x49, 0xff, 0x9f, 0x13, 0x72, 0xc0, 0x3c, 0xb0, 0x6d, 0xca,
	0x72, 0x7f, 0xf1, 0x68, 0x8c, 0xb5, 0x49, 0x9f, 0xb6, 0x35, 0x4e, 0x00, 0xa2, 0xea, 0xdf, 0x5b,
	0x74, 0xb6, 0xba, 0x6a, 0x3b, 0x02, 0x28, 0xf8, 0xc8, 0xf1, 0x07, 0x0b, 0x33, 0xa1, 0x69, 0x50,
	0x05, 0xc1, 0x21, 0x58, 0x42, 0xf5, 0x09, 0x79, 0x5c, 0x6d, 0xa9, 0xdf, 0x6c, 0x69, 0xc1, 0xbd,
	0x17, 0xc8, 0xb4, 0x17, 0x00, 0xbc, 0x68, 0xb4, 0x1d, 0x30, 0x77, 0x45, 0x20, 0x88, 0xc2, 0xc0,
	0xa4, 0x6d, 0xad, 0x48, 0xa6, 0x37, 0x92, 0xe9, 0xc7, 0xdf, 0x00, 0xdd, 0x06, 0xc8, 0x82, 0xe7,
	0xf1, 0xda, 0xf3, 0x48, 0x1c, 0xdb, 0x3d, 0xf4, 0x00, 0xc3, 0xcf, 0xae, 0x1f, 0xac, 0x29, 0xd9,
	0x2c, 0x49, 0xb2, 0x09, 0xbf, 0xac, 0xc8, 0x37, 0x5b, 0x43, 0x6f, 0xe1, 0xe1, 0x1a, 0x4e, 0x02,
	0xe2, 0x2e, 0x08, 0xb5, 0x75, 0xf4, 0x0a, 0x2c, 0xd9, 0x8a, 0xd6, 0x89, 0x6d, 0xa0, 0xf7, 0xf0,
	0x4e, 0x01, 0x36, 0x9e, 0x1b, 0x26, 0x91, 0x4b, 0x63, 0xe2, 0x05, 0x3e, 0x09, 0x13, 0x3f, 0xb2,
	0xef, 0x66, 0x7f, 0x34, 0x18, 0xac, 0xd8, 0xe5, 0x2b, 0x0f, 0x34, 0xe6, 0x27, 0x86, 0xa6, 0x60,
	0x2c, 0x59, 0x89, 0x64, 0x40, 0xdd, 0x75, 0x38, 0x48, 0x85, 0xea, 0x0f, 0xc7, 0x3d, 0xce, 0x8f,
	0xaa, 0x8e, 0xdf, 0xbd, 0x63, 0xcb, 0x57, 0x1e, 0x01, 0xf7, 0xd0, 0x27, 0x30, 0x7d, 0x19, 0x8d,
	0xdd, 0x30, 0xda, 0x58, 0x9d, 0xa1, 0x82, 0x28, 0x23, 0x77, 0xfc, 0xb2, 0x91, 0x14, 0x54, 0xae,
	0xde, 0x79, 0x7d, 0x85, 0xc9, 0x91, 0xef, 0xcf, 0x04, 0x3a, 0xff, 0x1b, 0x00, 0x00, 0xff, 0xff,
	0x20, 0x46, 0x30, 0x94, 0x3a, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// KeyValueStoreClient is the client API for KeyValueStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type KeyValueStoreClient interface {
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error)
	IsLeader(ctx context.Context, in *CLRequest, opts ...grpc.CallOption) (*CLResponse, error)
	Exit(ctx context.Context, in *ExitRequest, opts ...grpc.CallOption) (*ExitResponse, error)
}

type keyValueStoreClient struct {
	cc *grpc.ClientConn
}

func NewKeyValueStoreClient(cc *grpc.ClientConn) KeyValueStoreClient {
	return &keyValueStoreClient{cc}
}

func (c *keyValueStoreClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/proto.KeyValueStore/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueStoreClient) Put(ctx context.Context, in *PutRequest, opts ...grpc.CallOption) (*PutResponse, error) {
	out := new(PutResponse)
	err := c.cc.Invoke(ctx, "/proto.KeyValueStore/Put", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueStoreClient) IsLeader(ctx context.Context, in *CLRequest, opts ...grpc.CallOption) (*CLResponse, error) {
	out := new(CLResponse)
	err := c.cc.Invoke(ctx, "/proto.KeyValueStore/IsLeader", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *keyValueStoreClient) Exit(ctx context.Context, in *ExitRequest, opts ...grpc.CallOption) (*ExitResponse, error) {
	out := new(ExitResponse)
	err := c.cc.Invoke(ctx, "/proto.KeyValueStore/Exit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// KeyValueStoreServer is the server API for KeyValueStore service.
type KeyValueStoreServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Put(context.Context, *PutRequest) (*PutResponse, error)
	IsLeader(context.Context, *CLRequest) (*CLResponse, error)
	Exit(context.Context, *ExitRequest) (*ExitResponse, error)
}

// UnimplementedKeyValueStoreServer can be embedded to have forward compatible implementations.
type UnimplementedKeyValueStoreServer struct {
}

func (*UnimplementedKeyValueStoreServer) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedKeyValueStoreServer) Put(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Put not implemented")
}
func (*UnimplementedKeyValueStoreServer) IsLeader(ctx context.Context, req *CLRequest) (*CLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsLeader not implemented")
}
func (*UnimplementedKeyValueStoreServer) Exit(ctx context.Context, req *ExitRequest) (*ExitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exit not implemented")
}

func RegisterKeyValueStoreServer(s *grpc.Server, srv KeyValueStoreServer) {
	s.RegisterService(&_KeyValueStore_serviceDesc, srv)
}

func _KeyValueStore_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueStoreServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.KeyValueStore/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueStoreServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValueStore_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueStoreServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.KeyValueStore/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueStoreServer).Put(ctx, req.(*PutRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValueStore_IsLeader_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CLRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueStoreServer).IsLeader(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.KeyValueStore/IsLeader",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueStoreServer).IsLeader(ctx, req.(*CLRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _KeyValueStore_Exit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(KeyValueStoreServer).Exit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.KeyValueStore/Exit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(KeyValueStoreServer).Exit(ctx, req.(*ExitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _KeyValueStore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.KeyValueStore",
	HandlerType: (*KeyValueStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _KeyValueStore_Get_Handler,
		},
		{
			MethodName: "Put",
			Handler:    _KeyValueStore_Put_Handler,
		},
		{
			MethodName: "IsLeader",
			Handler:    _KeyValueStore_IsLeader_Handler,
		},
		{
			MethodName: "Exit",
			Handler:    _KeyValueStore_Exit_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kvstore.proto",
}
