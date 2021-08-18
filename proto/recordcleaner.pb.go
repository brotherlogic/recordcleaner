// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: recordcleaner.proto

package proto

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

type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LastCleanTime map[int32]int64 `protobuf:"bytes,1,rep,name=last_clean_time,json=lastCleanTime,proto3" json:"last_clean_time,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	CurrentCount  int32           `protobuf:"varint,2,opt,name=current_count,json=currentCount,proto3" json:"current_count,omitempty"`
	LastWater     int64           `protobuf:"varint,3,opt,name=last_water,json=lastWater,proto3" json:"last_water,omitempty"`
	LastFilter    int64           `protobuf:"varint,4,opt,name=last_filter,json=lastFilter,proto3" json:"last_filter,omitempty"`
	DayOfYear     int32           `protobuf:"varint,5,opt,name=day_of_year,json=dayOfYear,proto3" json:"day_of_year,omitempty"`
	DayCount      int32           `protobuf:"varint,6,opt,name=day_count,json=dayCount,proto3" json:"day_count,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordcleaner_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_recordcleaner_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_recordcleaner_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetLastCleanTime() map[int32]int64 {
	if x != nil {
		return x.LastCleanTime
	}
	return nil
}

func (x *Config) GetCurrentCount() int32 {
	if x != nil {
		return x.CurrentCount
	}
	return 0
}

func (x *Config) GetLastWater() int64 {
	if x != nil {
		return x.LastWater
	}
	return 0
}

func (x *Config) GetLastFilter() int64 {
	if x != nil {
		return x.LastFilter
	}
	return 0
}

func (x *Config) GetDayOfYear() int32 {
	if x != nil {
		return x.DayOfYear
	}
	return 0
}

func (x *Config) GetDayCount() int32 {
	if x != nil {
		return x.DayCount
	}
	return 0
}

type GetCleanRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetCleanRequest) Reset() {
	*x = GetCleanRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordcleaner_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCleanRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCleanRequest) ProtoMessage() {}

func (x *GetCleanRequest) ProtoReflect() protoreflect.Message {
	mi := &file_recordcleaner_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCleanRequest.ProtoReflect.Descriptor instead.
func (*GetCleanRequest) Descriptor() ([]byte, []int) {
	return file_recordcleaner_proto_rawDescGZIP(), []int{1}
}

type GetCleanResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceId int32   `protobuf:"varint,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	Seen       []int32 `protobuf:"varint,2,rep,packed,name=seen,proto3" json:"seen,omitempty"`
}

func (x *GetCleanResponse) Reset() {
	*x = GetCleanResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordcleaner_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCleanResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCleanResponse) ProtoMessage() {}

func (x *GetCleanResponse) ProtoReflect() protoreflect.Message {
	mi := &file_recordcleaner_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCleanResponse.ProtoReflect.Descriptor instead.
func (*GetCleanResponse) Descriptor() ([]byte, []int) {
	return file_recordcleaner_proto_rawDescGZIP(), []int{2}
}

func (x *GetCleanResponse) GetInstanceId() int32 {
	if x != nil {
		return x.InstanceId
	}
	return 0
}

func (x *GetCleanResponse) GetSeen() []int32 {
	if x != nil {
		return x.Seen
	}
	return nil
}

type ServiceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Water   bool `protobuf:"varint,1,opt,name=water,proto3" json:"water,omitempty"`
	Fileter bool `protobuf:"varint,2,opt,name=fileter,proto3" json:"fileter,omitempty"`
}

func (x *ServiceRequest) Reset() {
	*x = ServiceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordcleaner_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceRequest) ProtoMessage() {}

func (x *ServiceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_recordcleaner_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceRequest.ProtoReflect.Descriptor instead.
func (*ServiceRequest) Descriptor() ([]byte, []int) {
	return file_recordcleaner_proto_rawDescGZIP(), []int{3}
}

func (x *ServiceRequest) GetWater() bool {
	if x != nil {
		return x.Water
	}
	return false
}

func (x *ServiceRequest) GetFileter() bool {
	if x != nil {
		return x.Fileter
	}
	return false
}

type ServiceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ServiceResponse) Reset() {
	*x = ServiceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_recordcleaner_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceResponse) ProtoMessage() {}

func (x *ServiceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_recordcleaner_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceResponse.ProtoReflect.Descriptor instead.
func (*ServiceResponse) Descriptor() ([]byte, []int) {
	return file_recordcleaner_proto_rawDescGZIP(), []int{4}
}

var File_recordcleaner_proto protoreflect.FileDescriptor

var file_recordcleaner_proto_rawDesc = []byte{
	0x0a, 0x13, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x63, 0x6c, 0x65,
	0x61, 0x6e, 0x65, 0x72, 0x22, 0xbe, 0x02, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12,
	0x50, 0x0a, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x5f, 0x74, 0x69,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x72, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e,
	0x4c, 0x61, 0x73, 0x74, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x54, 0x69, 0x6d,
	0x65, 0x12, 0x23, 0x0a, 0x0d, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e,
	0x74, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x77,
	0x61, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x6c, 0x61, 0x73, 0x74,
	0x57, 0x61, 0x74, 0x65, 0x72, 0x12, 0x1f, 0x0a, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x66, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x6c, 0x61, 0x73, 0x74,
	0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0b, 0x64, 0x61, 0x79, 0x5f, 0x6f, 0x66,
	0x5f, 0x79, 0x65, 0x61, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x64, 0x61, 0x79,
	0x4f, 0x66, 0x59, 0x65, 0x61, 0x72, 0x12, 0x1b, 0x0a, 0x09, 0x64, 0x61, 0x79, 0x5f, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x64, 0x61, 0x79, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x1a, 0x40, 0x0a, 0x12, 0x4c, 0x61, 0x73, 0x74, 0x43, 0x6c, 0x65, 0x61, 0x6e,
	0x54, 0x69, 0x6d, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x11, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x65, 0x61,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x47, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x43,
	0x6c, 0x65, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x65, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x03, 0x28, 0x05, 0x52, 0x04, 0x73, 0x65, 0x65,
	0x6e, 0x22, 0x40, 0x0a, 0x0e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x61, 0x74, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x05, 0x77, 0x61, 0x74, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x66, 0x69, 0x6c,
	0x65, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x65,
	0x74, 0x65, 0x72, 0x22, 0x11, 0x0a, 0x0f, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0xb1, 0x01, 0x0a, 0x14, 0x52, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x4d, 0x0a, 0x08, 0x47, 0x65, 0x74, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x12, 0x1e, 0x2e, 0x72, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x6c, 0x65, 0x61, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x72, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x6c, 0x65, 0x61, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4a,
	0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x1d, 0x2e, 0x72, 0x65, 0x63, 0x6f,
	0x72, 0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x72, 0x65, 0x63, 0x6f, 0x72,
	0x64, 0x63, 0x6c, 0x65, 0x61, 0x6e, 0x65, 0x72, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x62, 0x72, 0x6f, 0x74, 0x68, 0x65, 0x72,
	0x6c, 0x6f, 0x67, 0x69, 0x63, 0x2f, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x63, 0x6c, 0x65, 0x61,
	0x6e, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_recordcleaner_proto_rawDescOnce sync.Once
	file_recordcleaner_proto_rawDescData = file_recordcleaner_proto_rawDesc
)

func file_recordcleaner_proto_rawDescGZIP() []byte {
	file_recordcleaner_proto_rawDescOnce.Do(func() {
		file_recordcleaner_proto_rawDescData = protoimpl.X.CompressGZIP(file_recordcleaner_proto_rawDescData)
	})
	return file_recordcleaner_proto_rawDescData
}

var file_recordcleaner_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_recordcleaner_proto_goTypes = []interface{}{
	(*Config)(nil),           // 0: recordcleaner.Config
	(*GetCleanRequest)(nil),  // 1: recordcleaner.GetCleanRequest
	(*GetCleanResponse)(nil), // 2: recordcleaner.GetCleanResponse
	(*ServiceRequest)(nil),   // 3: recordcleaner.ServiceRequest
	(*ServiceResponse)(nil),  // 4: recordcleaner.ServiceResponse
	nil,                      // 5: recordcleaner.Config.LastCleanTimeEntry
}
var file_recordcleaner_proto_depIdxs = []int32{
	5, // 0: recordcleaner.Config.last_clean_time:type_name -> recordcleaner.Config.LastCleanTimeEntry
	1, // 1: recordcleaner.RecordCleanerService.GetClean:input_type -> recordcleaner.GetCleanRequest
	3, // 2: recordcleaner.RecordCleanerService.Service:input_type -> recordcleaner.ServiceRequest
	2, // 3: recordcleaner.RecordCleanerService.GetClean:output_type -> recordcleaner.GetCleanResponse
	4, // 4: recordcleaner.RecordCleanerService.Service:output_type -> recordcleaner.ServiceResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_recordcleaner_proto_init() }
func file_recordcleaner_proto_init() {
	if File_recordcleaner_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_recordcleaner_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_recordcleaner_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCleanRequest); i {
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
		file_recordcleaner_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCleanResponse); i {
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
		file_recordcleaner_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceRequest); i {
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
		file_recordcleaner_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceResponse); i {
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
			RawDescriptor: file_recordcleaner_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_recordcleaner_proto_goTypes,
		DependencyIndexes: file_recordcleaner_proto_depIdxs,
		MessageInfos:      file_recordcleaner_proto_msgTypes,
	}.Build()
	File_recordcleaner_proto = out.File
	file_recordcleaner_proto_rawDesc = nil
	file_recordcleaner_proto_goTypes = nil
	file_recordcleaner_proto_depIdxs = nil
}
