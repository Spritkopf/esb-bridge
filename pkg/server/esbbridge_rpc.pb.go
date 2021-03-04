// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: pkg/server/esbbridge_rpc.proto

package server

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Listener holds all information to listen for a specific package
type Listener struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr []byte `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Cmd  []byte `protobuf:"bytes,2,opt,name=cmd,proto3" json:"cmd,omitempty"`
}

func (x *Listener) Reset() {
	*x = Listener{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_server_esbbridge_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Listener) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Listener) ProtoMessage() {}

func (x *Listener) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_server_esbbridge_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Listener.ProtoReflect.Descriptor instead.
func (*Listener) Descriptor() ([]byte, []int) {
	return file_pkg_server_esbbridge_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *Listener) GetAddr() []byte {
	if x != nil {
		return x.Addr
	}
	return nil
}

func (x *Listener) GetCmd() []byte {
	if x != nil {
		return x.Cmd
	}
	return nil
}

// EsbMessage holds all information for an ESB transaction
type EsbMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr    []byte `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Cmd     []byte `protobuf:"bytes,2,opt,name=cmd,proto3" json:"cmd,omitempty"`
	Payload []byte `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *EsbMessage) Reset() {
	*x = EsbMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_server_esbbridge_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EsbMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EsbMessage) ProtoMessage() {}

func (x *EsbMessage) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_server_esbbridge_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EsbMessage.ProtoReflect.Descriptor instead.
func (*EsbMessage) Descriptor() ([]byte, []int) {
	return file_pkg_server_esbbridge_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *EsbMessage) GetAddr() []byte {
	if x != nil {
		return x.Addr
	}
	return nil
}

func (x *EsbMessage) GetCmd() []byte {
	if x != nil {
		return x.Cmd
	}
	return nil
}

func (x *EsbMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

var File_pkg_server_esbbridge_rpc_proto protoreflect.FileDescriptor

var file_pkg_server_esbbridge_rpc_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x65, 0x73, 0x62,
	0x62, 0x72, 0x69, 0x64, 0x67, 0x65, 0x5f, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x06, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x22, 0x30, 0x0a, 0x08, 0x4c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x63, 0x6d, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x22, 0x4c, 0x0a, 0x0a, 0x45, 0x73,
	0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x12, 0x10, 0x0a, 0x03,
	0x63, 0x6d, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x63, 0x6d, 0x64, 0x12, 0x18,
	0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x32, 0x75, 0x0a, 0x09, 0x45, 0x73, 0x62, 0x42,
	0x72, 0x69, 0x64, 0x67, 0x65, 0x12, 0x34, 0x0a, 0x08, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x66, 0x65,
	0x72, 0x12, 0x12, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x45, 0x73, 0x62, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x12, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x45,
	0x73, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x12, 0x32, 0x0a, 0x06, 0x4c,
	0x69, 0x73, 0x74, 0x65, 0x6e, 0x12, 0x10, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x4c,
	0x69, 0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x1a, 0x12, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x2e, 0x45, 0x73, 0x62, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42,
	0x36, 0x5a, 0x34, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x70,
	0x72, 0x69, 0x74, 0x6b, 0x6f, 0x70, 0x66, 0x2f, 0x65, 0x73, 0x62, 0x2d, 0x62, 0x72, 0x69, 0x64,
	0x67, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x65, 0x73, 0x62, 0x62, 0x72, 0x69, 0x64, 0x67, 0x65,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_server_esbbridge_rpc_proto_rawDescOnce sync.Once
	file_pkg_server_esbbridge_rpc_proto_rawDescData = file_pkg_server_esbbridge_rpc_proto_rawDesc
)

func file_pkg_server_esbbridge_rpc_proto_rawDescGZIP() []byte {
	file_pkg_server_esbbridge_rpc_proto_rawDescOnce.Do(func() {
		file_pkg_server_esbbridge_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_server_esbbridge_rpc_proto_rawDescData)
	})
	return file_pkg_server_esbbridge_rpc_proto_rawDescData
}

var file_pkg_server_esbbridge_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_server_esbbridge_rpc_proto_goTypes = []interface{}{
	(*Listener)(nil),   // 0: server.Listener
	(*EsbMessage)(nil), // 1: server.EsbMessage
}
var file_pkg_server_esbbridge_rpc_proto_depIdxs = []int32{
	1, // 0: server.EsbBridge.Transfer:input_type -> server.EsbMessage
	0, // 1: server.EsbBridge.Listen:input_type -> server.Listener
	1, // 2: server.EsbBridge.Transfer:output_type -> server.EsbMessage
	1, // 3: server.EsbBridge.Listen:output_type -> server.EsbMessage
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_server_esbbridge_rpc_proto_init() }
func file_pkg_server_esbbridge_rpc_proto_init() {
	if File_pkg_server_esbbridge_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_server_esbbridge_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Listener); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_server_esbbridge_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EsbMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_server_esbbridge_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_server_esbbridge_rpc_proto_goTypes,
		DependencyIndexes: file_pkg_server_esbbridge_rpc_proto_depIdxs,
		MessageInfos:      file_pkg_server_esbbridge_rpc_proto_msgTypes,
	}.Build()
	File_pkg_server_esbbridge_rpc_proto = out.File
	file_pkg_server_esbbridge_rpc_proto_rawDesc = nil
	file_pkg_server_esbbridge_rpc_proto_goTypes = nil
	file_pkg_server_esbbridge_rpc_proto_depIdxs = nil
}
