// Code generated by protoc-gen-go. DO NOT EDIT.
// source: guideocelot.proto

/*
Package models is a generated protocol buffer package.

It is generated from these files:
	guideocelot.proto

It has these top-level messages:
	CredWrapper
	Credentials
*/
package models

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
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

type CredWrapper struct {
	Credentials []*Credentials `protobuf:"bytes,1,rep,name=credentials" json:"credentials,omitempty"`
}

func (m *CredWrapper) Reset()                    { *m = CredWrapper{} }
func (m *CredWrapper) String() string            { return proto.CompactTextString(m) }
func (*CredWrapper) ProtoMessage()               {}
func (*CredWrapper) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CredWrapper) GetCredentials() []*Credentials {
	if m != nil {
		return m.Credentials
	}
	return nil
}

type Credentials struct {
	ClientId     string `protobuf:"bytes,1,opt,name=clientId" json:"clientId,omitempty"`
	ClientSecret string `protobuf:"bytes,2,opt,name=clientSecret" json:"clientSecret,omitempty"`
	TokenURL     string `protobuf:"bytes,3,opt,name=tokenURL" json:"tokenURL,omitempty"`
	AcctName     string `protobuf:"bytes,4,opt,name=acctName" json:"acctName,omitempty"`
	Type         string `protobuf:"bytes,5,opt,name=type" json:"type,omitempty"`
}

func (m *Credentials) Reset()                    { *m = Credentials{} }
func (m *Credentials) String() string            { return proto.CompactTextString(m) }
func (*Credentials) ProtoMessage()               {}
func (*Credentials) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Credentials) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *Credentials) GetClientSecret() string {
	if m != nil {
		return m.ClientSecret
	}
	return ""
}

func (m *Credentials) GetTokenURL() string {
	if m != nil {
		return m.TokenURL
	}
	return ""
}

func (m *Credentials) GetAcctName() string {
	if m != nil {
		return m.AcctName
	}
	return ""
}

func (m *Credentials) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func init() {
	proto.RegisterType((*CredWrapper)(nil), "models.CredWrapper")
	proto.RegisterType((*Credentials)(nil), "models.Credentials")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for GuideOcelot service

type GuideOcelotClient interface {
	GetCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*CredWrapper, error)
	SetCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
}

type guideOcelotClient struct {
	cc *grpc.ClientConn
}

func NewGuideOcelotClient(cc *grpc.ClientConn) GuideOcelotClient {
	return &guideOcelotClient{cc}
}

func (c *guideOcelotClient) GetCreds(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*CredWrapper, error) {
	out := new(CredWrapper)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/GetCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *guideOcelotClient) SetCreds(ctx context.Context, in *Credentials, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/models.GuideOcelot/SetCreds", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for GuideOcelot service

type GuideOcelotServer interface {
	GetCreds(context.Context, *google_protobuf.Empty) (*CredWrapper, error)
	SetCreds(context.Context, *Credentials) (*google_protobuf.Empty, error)
}

func RegisterGuideOcelotServer(s *grpc.Server, srv GuideOcelotServer) {
	s.RegisterService(&_GuideOcelot_serviceDesc, srv)
}

func _GuideOcelot_GetCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).GetCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/GetCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).GetCreds(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _GuideOcelot_SetCreds_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Credentials)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GuideOcelotServer).SetCreds(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/models.GuideOcelot/SetCreds",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GuideOcelotServer).SetCreds(ctx, req.(*Credentials))
	}
	return interceptor(ctx, in, info, handler)
}

var _GuideOcelot_serviceDesc = grpc.ServiceDesc{
	ServiceName: "models.GuideOcelot",
	HandlerType: (*GuideOcelotServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCreds",
			Handler:    _GuideOcelot_GetCreds_Handler,
		},
		{
			MethodName: "SetCreds",
			Handler:    _GuideOcelot_SetCreds_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "guideocelot.proto",
}

func init() { proto.RegisterFile("guideocelot.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 314 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0xcf, 0x4a, 0x33, 0x31,
	0x14, 0xc5, 0x99, 0xb6, 0x5f, 0x69, 0x33, 0x1f, 0x42, 0xa3, 0x48, 0x18, 0x5d, 0x94, 0xac, 0x8a,
	0x8b, 0x0c, 0x56, 0xdc, 0xb8, 0xad, 0x52, 0x14, 0xff, 0x40, 0x8b, 0xb8, 0x4e, 0x33, 0xd7, 0x32,
	0x38, 0x4d, 0x42, 0xe6, 0x56, 0xe8, 0xd6, 0x57, 0x70, 0xe5, 0x1b, 0xf8, 0x3e, 0xbe, 0x82, 0x0f,
	0x22, 0x93, 0x74, 0xac, 0x05, 0xdd, 0xdd, 0x73, 0xcf, 0xbd, 0xe7, 0x26, 0x3f, 0xd2, 0x9b, 0x2f,
	0xf3, 0x0c, 0x8c, 0x82, 0xc2, 0xa0, 0xb0, 0xce, 0xa0, 0xa1, 0xed, 0x85, 0xc9, 0xa0, 0x28, 0x93,
	0x83, 0xb9, 0x31, 0xf3, 0x02, 0x52, 0xdf, 0x9d, 0x2d, 0x1f, 0x53, 0x58, 0x58, 0x5c, 0x85, 0xa1,
	0xe4, 0x70, 0x6d, 0x4a, 0x9b, 0xa7, 0x52, 0x6b, 0x83, 0x12, 0x73, 0xa3, 0xcb, 0xe0, 0xf2, 0x73,
	0x12, 0x8f, 0x1c, 0x64, 0x0f, 0x4e, 0x5a, 0x0b, 0x8e, 0x9e, 0x92, 0x58, 0x39, 0xc8, 0x40, 0x63,
	0x2e, 0x8b, 0x92, 0x45, 0xfd, 0xe6, 0x20, 0x1e, 0xee, 0x8a, 0x70, 0x47, 0x8c, 0x36, 0xd6, 0xe4,
	0xe7, 0x1c, 0x7f, 0x8b, 0x42, 0xcc, 0x5a, 0xd3, 0x84, 0x74, 0x54, 0x91, 0x83, 0xc6, 0xcb, 0x8c,
	0x45, 0xfd, 0x68, 0xd0, 0x9d, 0x7c, 0x6b, 0xca, 0xc9, 0xff, 0x50, 0x4f, 0x41, 0x39, 0x40, 0xd6,
	0xf0, 0xfe, 0x56, 0xaf, 0xda, 0x47, 0xf3, 0x04, 0xfa, 0x7e, 0x72, 0xcd, 0x9a, 0x61, 0xbf, 0xd6,
	0x95, 0x27, 0x95, 0xc2, 0x5b, 0xb9, 0x00, 0xd6, 0x0a, 0x5e, 0xad, 0x29, 0x25, 0x2d, 0x5c, 0x59,
	0x60, 0xff, 0x7c, 0xdf, 0xd7, 0xc3, 0xf7, 0x88, 0xc4, 0xe3, 0x0a, 0xdd, 0x9d, 0x47, 0x47, 0xaf,
	0x48, 0x67, 0x0c, 0x58, 0xbd, 0xb6, 0xa4, 0xfb, 0x22, 0xc0, 0x11, 0x35, 0x39, 0x71, 0x51, 0x91,
	0x4b, 0xb6, 0x7e, 0xbc, 0x66, 0xc3, 0x7b, 0x2f, 0x1f, 0x9f, 0xaf, 0x8d, 0x98, 0x76, 0xd3, 0xe7,
	0xe3, 0x54, 0xf9, 0xfd, 0x1b, 0xd2, 0x99, 0xd6, 0x59, 0xbf, 0x51, 0x4a, 0xfe, 0x38, 0xc0, 0xf7,
	0x7c, 0xd6, 0x0e, 0xdf, 0x64, 0x9d, 0x45, 0x47, 0xb3, 0xb6, 0x9f, 0x3a, 0xf9, 0x0a, 0x00, 0x00,
	0xff, 0xff, 0x37, 0x5f, 0x17, 0x5e, 0xeb, 0x01, 0x00, 0x00,
}