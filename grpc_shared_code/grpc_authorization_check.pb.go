// Code generated by protoc-gen-go. DO NOT EDIT.
// source: grpc_authorization_check.proto

package authorizationcheck

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// проверяемая кука
type Cookie struct {
	Sessionid            string   `protobuf:"bytes,1,opt,name=sessionid,proto3" json:"sessionid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Cookie) Reset()         { *m = Cookie{} }
func (m *Cookie) String() string { return proto.CompactTextString(m) }
func (*Cookie) ProtoMessage()    {}
func (*Cookie) Descriptor() ([]byte, []int) {
	return fileDescriptor_823a6576501bf23e, []int{0}
}

func (m *Cookie) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Cookie.Unmarshal(m, b)
}
func (m *Cookie) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Cookie.Marshal(b, m, deterministic)
}
func (m *Cookie) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Cookie.Merge(m, src)
}
func (m *Cookie) XXX_Size() int {
	return xxx_messageInfo_Cookie.Size(m)
}
func (m *Cookie) XXX_DiscardUnknown() {
	xxx_messageInfo_Cookie.DiscardUnknown(m)
}

var xxx_messageInfo_Cookie proto.InternalMessageInfo

func (m *Cookie) GetSessionid() string {
	if m != nil {
		return m.Sessionid
	}
	return ""
}

// пользователь, которому кука принадлежит
type User struct {
	Login                string   `protobuf:"bytes,1,opt,name=login,proto3" json:"login,omitempty"`
	AvatarAddress        string   `protobuf:"bytes,2,opt,name=avatar_address,json=avatarAddress,proto3" json:"avatar_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *User) Reset()         { *m = User{} }
func (m *User) String() string { return proto.CompactTextString(m) }
func (*User) ProtoMessage()    {}
func (*User) Descriptor() ([]byte, []int) {
	return fileDescriptor_823a6576501bf23e, []int{1}
}

func (m *User) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_User.Unmarshal(m, b)
}
func (m *User) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_User.Marshal(b, m, deterministic)
}
func (m *User) XXX_Merge(src proto.Message) {
	xxx_messageInfo_User.Merge(m, src)
}
func (m *User) XXX_Size() int {
	return xxx_messageInfo_User.Size(m)
}
func (m *User) XXX_DiscardUnknown() {
	xxx_messageInfo_User.DiscardUnknown(m)
}

var xxx_messageInfo_User proto.InternalMessageInfo

func (m *User) GetLogin() string {
	if m != nil {
		return m.Login
	}
	return ""
}

func (m *User) GetAvatarAddress() string {
	if m != nil {
		return m.AvatarAddress
	}
	return ""
}

func init() {
	proto.RegisterType((*Cookie)(nil), "authorizationcheck.Cookie")
	proto.RegisterType((*User)(nil), "authorizationcheck.User")
}

func init() { proto.RegisterFile("grpc_authorization_check.proto", fileDescriptor_823a6576501bf23e) }

var fileDescriptor_823a6576501bf23e = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0x2f, 0x2a, 0x48,
	0x8e, 0x4f, 0x2c, 0x2d, 0xc9, 0xc8, 0x2f, 0xca, 0xac, 0x4a, 0x2c, 0xc9, 0xcc, 0xcf, 0x8b, 0x4f,
	0xce, 0x48, 0x4d, 0xce, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x42, 0x91, 0x02, 0xcb,
	0x28, 0xa9, 0x71, 0xb1, 0x39, 0xe7, 0xe7, 0x67, 0x67, 0xa6, 0x0a, 0xc9, 0x70, 0x71, 0x16, 0xa7,
	0x16, 0x17, 0x67, 0xe6, 0xe7, 0x65, 0xa6, 0x48, 0x30, 0x2a, 0x30, 0x6a, 0x70, 0x06, 0x21, 0x04,
	0x94, 0x9c, 0xb9, 0x58, 0x42, 0x8b, 0x53, 0x8b, 0x84, 0x44, 0xb8, 0x58, 0x73, 0xf2, 0xd3, 0x33,
	0xf3, 0xa0, 0x2a, 0x20, 0x1c, 0x21, 0x55, 0x2e, 0xbe, 0xc4, 0xb2, 0xc4, 0x92, 0xc4, 0xa2, 0xf8,
	0xc4, 0x94, 0x94, 0xa2, 0xd4, 0xe2, 0x62, 0x09, 0x26, 0xb0, 0x34, 0x2f, 0x44, 0xd4, 0x11, 0x22,
	0x68, 0x14, 0xc1, 0x25, 0xe4, 0x88, 0xec, 0x04, 0x67, 0x90, 0x13, 0x84, 0x9c, 0xb8, 0x38, 0xdc,
	0x53, 0x4b, 0x7c, 0xc0, 0x06, 0x49, 0xe9, 0x61, 0xba, 0x51, 0x0f, 0xe2, 0x40, 0x29, 0x09, 0x6c,
	0x72, 0x20, 0x47, 0x29, 0x31, 0x24, 0xb1, 0x81, 0x7d, 0x68, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff,
	0x04, 0x69, 0x9a, 0x90, 0x03, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// AuthorizationCheckClient is the client API for AuthorizationCheck service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type AuthorizationCheckClient interface {
	GetLogin(ctx context.Context, in *Cookie, opts ...grpc.CallOption) (*User, error)
}

type authorizationCheckClient struct {
	cc *grpc.ClientConn
}

func NewAuthorizationCheckClient(cc *grpc.ClientConn) AuthorizationCheckClient {
	return &authorizationCheckClient{cc}
}

func (c *authorizationCheckClient) GetLogin(ctx context.Context, in *Cookie, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, "/authorizationcheck.AuthorizationCheck/GetLogin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthorizationCheckServer is the server API for AuthorizationCheck service.
type AuthorizationCheckServer interface {
	GetLogin(context.Context, *Cookie) (*User, error)
}

func RegisterAuthorizationCheckServer(s *grpc.Server, srv AuthorizationCheckServer) {
	s.RegisterService(&_AuthorizationCheck_serviceDesc, srv)
}

func _AuthorizationCheck_GetLogin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Cookie)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthorizationCheckServer).GetLogin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/authorizationcheck.AuthorizationCheck/GetLogin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthorizationCheckServer).GetLogin(ctx, req.(*Cookie))
	}
	return interceptor(ctx, in, info, handler)
}

var _AuthorizationCheck_serviceDesc = grpc.ServiceDesc{
	ServiceName: "authorizationcheck.AuthorizationCheck",
	HandlerType: (*AuthorizationCheckServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetLogin",
			Handler:    _AuthorizationCheck_GetLogin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc_authorization_check.proto",
}
