// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.12.4
// source: protos/pure_price.proto

package server

import (
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

type PriceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Source string `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	Base   string `protobuf:"bytes,2,opt,name=base,proto3" json:"base,omitempty"`
	Quote  string `protobuf:"bytes,3,opt,name=quote,proto3" json:"quote,omitempty"`
}

func (x *PriceRequest) Reset() {
	*x = PriceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_pure_price_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PriceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PriceRequest) ProtoMessage() {}

func (x *PriceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_pure_price_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PriceRequest.ProtoReflect.Descriptor instead.
func (*PriceRequest) Descriptor() ([]byte, []int) {
	return file_protos_pure_price_proto_rawDescGZIP(), []int{0}
}

func (x *PriceRequest) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *PriceRequest) GetBase() string {
	if x != nil {
		return x.Base
	}
	return ""
}

func (x *PriceRequest) GetQuote() string {
	if x != nil {
		return x.Quote
	}
	return ""
}

type PriceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Price float64 `protobuf:"fixed64,1,opt,name=price,proto3" json:"price,omitempty"`
}

func (x *PriceResponse) Reset() {
	*x = PriceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_pure_price_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PriceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PriceResponse) ProtoMessage() {}

func (x *PriceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_protos_pure_price_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PriceResponse.ProtoReflect.Descriptor instead.
func (*PriceResponse) Descriptor() ([]byte, []int) {
	return file_protos_pure_price_proto_rawDescGZIP(), []int{1}
}

func (x *PriceResponse) GetPrice() float64 {
	if x != nil {
		return x.Price
	}
	return 0
}

var File_protos_pure_price_proto protoreflect.FileDescriptor

var file_protos_pure_price_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x70, 0x75, 0x72, 0x65, 0x5f, 0x70, 0x72,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x22, 0x50, 0x0a, 0x0c, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x61, 0x73,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x61, 0x73, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x71, 0x75, 0x6f, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x71, 0x75,
	0x6f, 0x74, 0x65, 0x22, 0x25, 0x0a, 0x0d, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x01, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x32, 0x55, 0x0a, 0x12, 0x43, 0x72,
	0x79, 0x70, 0x74, 0x6f, 0x50, 0x72, 0x69, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x3f, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x12, 0x14, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2e, 0x50, 0x72, 0x69, 0x63,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x73, 0x2e, 0x50, 0x72, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x42, 0x0c, 0x5a, 0x0a, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_pure_price_proto_rawDescOnce sync.Once
	file_protos_pure_price_proto_rawDescData = file_protos_pure_price_proto_rawDesc
)

func file_protos_pure_price_proto_rawDescGZIP() []byte {
	file_protos_pure_price_proto_rawDescOnce.Do(func() {
		file_protos_pure_price_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_pure_price_proto_rawDescData)
	})
	return file_protos_pure_price_proto_rawDescData
}

var file_protos_pure_price_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protos_pure_price_proto_goTypes = []interface{}{
	(*PriceRequest)(nil),  // 0: protos.PriceRequest
	(*PriceResponse)(nil), // 1: protos.PriceResponse
}
var file_protos_pure_price_proto_depIdxs = []int32{
	0, // 0: protos.CryptoPriceService.GetCryptoPrice:input_type -> protos.PriceRequest
	1, // 1: protos.CryptoPriceService.GetCryptoPrice:output_type -> protos.PriceResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_protos_pure_price_proto_init() }
func file_protos_pure_price_proto_init() {
	if File_protos_pure_price_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_pure_price_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PriceRequest); i {
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
		file_protos_pure_price_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PriceResponse); i {
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
			RawDescriptor: file_protos_pure_price_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_protos_pure_price_proto_goTypes,
		DependencyIndexes: file_protos_pure_price_proto_depIdxs,
		MessageInfos:      file_protos_pure_price_proto_msgTypes,
	}.Build()
	File_protos_pure_price_proto = out.File
	file_protos_pure_price_proto_rawDesc = nil
	file_protos_pure_price_proto_goTypes = nil
	file_protos_pure_price_proto_depIdxs = nil
}
