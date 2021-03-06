// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.1
// source: slimarray.proto

package slimarray

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

// SlimArray compresses a uint32 array with overall trend by describing the trend
// with a polynomial, e.g., to store a sorted array is very common in practice.
// Such as an block-list of IP addresses, or a series of var-length record
// position on disk.
//
// E.g. a uint32 costs only 5 bits in average in a sorted array of a million
// number in range [0, 1000*1000].
//
// In addition to the unbelievable low memory footprint,
// a `Get` access is also very fast: it takes only 10 nano second in our
// benchmark.
//
// SlimArray is also ready for transport since it is defined with protobuf. E.g.:
//    a := slimarray.NewU32([]uint32{1, 2, 3})
//    bytes, err := proto.Marshal(a)
//
// Since 0.1.1
type SlimArray struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// N is the count of elts
	N    int32    `protobuf:"varint,10,opt,name=N,proto3" json:"N,omitempty"`
	Rank []uint64 `protobuf:"varint,19,rep,packed,name=Rank,proto3" json:"Rank,omitempty"`
	// Every 1024 elts segment has a 64-bit bitmap to describe the spans in it,
	// and another 64-bit rank: the count of `1` in preceding bitmaps.
	Bitmap []uint64 `protobuf:"varint,20,rep,packed,name=Bitmap,proto3" json:"Bitmap,omitempty"`
	// Polynomial and config of every span.
	// 3 doubles to represent a polynomial;
	Polynomials []float64 `protobuf:"fixed64,21,rep,packed,name=Polynomials,proto3" json:"Polynomials,omitempty"`
	// Config stores the offset of residuals in Residuals and the bit width to
	// store a residual in a span.
	Configs []int64 `protobuf:"varint,22,rep,packed,name=Configs,proto3" json:"Configs,omitempty"`
	// packed residuals for every elt.
	Residuals []uint64 `protobuf:"varint,23,rep,packed,name=Residuals,proto3" json:"Residuals,omitempty"`
}

func (x *SlimArray) Reset() {
	*x = SlimArray{}
	if protoimpl.UnsafeEnabled {
		mi := &file_slimarray_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SlimArray) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SlimArray) ProtoMessage() {}

func (x *SlimArray) ProtoReflect() protoreflect.Message {
	mi := &file_slimarray_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SlimArray.ProtoReflect.Descriptor instead.
func (*SlimArray) Descriptor() ([]byte, []int) {
	return file_slimarray_proto_rawDescGZIP(), []int{0}
}

func (x *SlimArray) GetN() int32 {
	if x != nil {
		return x.N
	}
	return 0
}

func (x *SlimArray) GetRank() []uint64 {
	if x != nil {
		return x.Rank
	}
	return nil
}

func (x *SlimArray) GetBitmap() []uint64 {
	if x != nil {
		return x.Bitmap
	}
	return nil
}

func (x *SlimArray) GetPolynomials() []float64 {
	if x != nil {
		return x.Polynomials
	}
	return nil
}

func (x *SlimArray) GetConfigs() []int64 {
	if x != nil {
		return x.Configs
	}
	return nil
}

func (x *SlimArray) GetResiduals() []uint64 {
	if x != nil {
		return x.Residuals
	}
	return nil
}

// SlimBytes is a var-length []byte array.
//
// Internally it use a SlimArray to store record positions.
// Thus the memory overhead is about 8 bit / record.
//
// Since 0.1.4
type SlimBytes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Positions is the array of start position of every record.
	// There are n + 1 int32 in it.
	// The last one equals len(Records)
	Positions *SlimArray `protobuf:"bytes,21,opt,name=Positions,proto3" json:"Positions,omitempty"`
	// Records is byte slice of all record packed together.
	Records []byte `protobuf:"bytes,22,opt,name=Records,proto3" json:"Records,omitempty"`
}

func (x *SlimBytes) Reset() {
	*x = SlimBytes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_slimarray_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SlimBytes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SlimBytes) ProtoMessage() {}

func (x *SlimBytes) ProtoReflect() protoreflect.Message {
	mi := &file_slimarray_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SlimBytes.ProtoReflect.Descriptor instead.
func (*SlimBytes) Descriptor() ([]byte, []int) {
	return file_slimarray_proto_rawDescGZIP(), []int{1}
}

func (x *SlimBytes) GetPositions() *SlimArray {
	if x != nil {
		return x.Positions
	}
	return nil
}

func (x *SlimBytes) GetRecords() []byte {
	if x != nil {
		return x.Records
	}
	return nil
}

var File_slimarray_proto protoreflect.FileDescriptor

var file_slimarray_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x73, 0x6c, 0x69, 0x6d, 0x61, 0x72, 0x72, 0x61, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x9f, 0x01, 0x0a, 0x09, 0x53, 0x6c, 0x69, 0x6d, 0x41, 0x72, 0x72, 0x61, 0x79, 0x12,
	0x0c, 0x0a, 0x01, 0x4e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x4e, 0x12, 0x12, 0x0a,
	0x04, 0x52, 0x61, 0x6e, 0x6b, 0x18, 0x13, 0x20, 0x03, 0x28, 0x04, 0x52, 0x04, 0x52, 0x61, 0x6e,
	0x6b, 0x12, 0x16, 0x0a, 0x06, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70, 0x18, 0x14, 0x20, 0x03, 0x28,
	0x04, 0x52, 0x06, 0x42, 0x69, 0x74, 0x6d, 0x61, 0x70, 0x12, 0x20, 0x0a, 0x0b, 0x50, 0x6f, 0x6c,
	0x79, 0x6e, 0x6f, 0x6d, 0x69, 0x61, 0x6c, 0x73, 0x18, 0x15, 0x20, 0x03, 0x28, 0x01, 0x52, 0x0b,
	0x50, 0x6f, 0x6c, 0x79, 0x6e, 0x6f, 0x6d, 0x69, 0x61, 0x6c, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x16, 0x20, 0x03, 0x28, 0x03, 0x52, 0x07, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x52, 0x65, 0x73, 0x69, 0x64, 0x75, 0x61,
	0x6c, 0x73, 0x18, 0x17, 0x20, 0x03, 0x28, 0x04, 0x52, 0x09, 0x52, 0x65, 0x73, 0x69, 0x64, 0x75,
	0x61, 0x6c, 0x73, 0x22, 0x4f, 0x0a, 0x09, 0x53, 0x6c, 0x69, 0x6d, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x12, 0x28, 0x0a, 0x09, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x15, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x53, 0x6c, 0x69, 0x6d, 0x41, 0x72, 0x72, 0x61, 0x79, 0x52,
	0x09, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x52, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x16, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x73, 0x42, 0x0b, 0x5a, 0x09, 0x73, 0x6c, 0x69, 0x6d, 0x61, 0x72, 0x72, 0x61,
	0x79, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_slimarray_proto_rawDescOnce sync.Once
	file_slimarray_proto_rawDescData = file_slimarray_proto_rawDesc
)

func file_slimarray_proto_rawDescGZIP() []byte {
	file_slimarray_proto_rawDescOnce.Do(func() {
		file_slimarray_proto_rawDescData = protoimpl.X.CompressGZIP(file_slimarray_proto_rawDescData)
	})
	return file_slimarray_proto_rawDescData
}

var file_slimarray_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_slimarray_proto_goTypes = []interface{}{
	(*SlimArray)(nil), // 0: SlimArray
	(*SlimBytes)(nil), // 1: SlimBytes
}
var file_slimarray_proto_depIdxs = []int32{
	0, // 0: SlimBytes.Positions:type_name -> SlimArray
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_slimarray_proto_init() }
func file_slimarray_proto_init() {
	if File_slimarray_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_slimarray_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SlimArray); i {
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
		file_slimarray_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SlimBytes); i {
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
			RawDescriptor: file_slimarray_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_slimarray_proto_goTypes,
		DependencyIndexes: file_slimarray_proto_depIdxs,
		MessageInfos:      file_slimarray_proto_msgTypes,
	}.Build()
	File_slimarray_proto = out.File
	file_slimarray_proto_rawDesc = nil
	file_slimarray_proto_goTypes = nil
	file_slimarray_proto_depIdxs = nil
}
