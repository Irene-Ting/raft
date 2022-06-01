// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.4.0
// source: pb/rpc.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_pb_rpc_proto protoreflect.FileDescriptor

var file_pb_rpc_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x70, 0x62, 0x2f, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02,
	0x70, 0x62, 0x1a, 0x10, 0x70, 0x62, 0x2f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x32, 0xd5, 0x01, 0x0a, 0x04, 0x52, 0x61, 0x66, 0x74, 0x12, 0x43, 0x0a,
	0x0c, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x17, 0x2e,
	0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x6c,
	0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x46, 0x0a, 0x0d, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72,
	0x69, 0x65, 0x73, 0x12, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45,
	0x6e, 0x74, 0x72, 0x69, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e,
	0x70, 0x62, 0x2e, 0x41, 0x70, 0x70, 0x65, 0x6e, 0x64, 0x45, 0x6e, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x40, 0x0a, 0x0b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f, 0x74, 0x65, 0x12, 0x16, 0x2e, 0x70, 0x62, 0x2e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x17, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x56, 0x6f,
	0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x1e, 0x5a, 0x1c,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a, 0x75, 0x73, 0x74, 0x69,
	0x6e, 0x30, 0x75, 0x30, 0x2f, 0x72, 0x61, 0x66, 0x74, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var file_pb_rpc_proto_goTypes = []interface{}{
	(*ApplyCommandRequest)(nil),   // 0: pb.ApplyCommandRequest
	(*AppendEntriesRequest)(nil),  // 1: pb.AppendEntriesRequest
	(*RequestVoteRequest)(nil),    // 2: pb.RequestVoteRequest
	(*ApplyCommandResponse)(nil),  // 3: pb.ApplyCommandResponse
	(*AppendEntriesResponse)(nil), // 4: pb.AppendEntriesResponse
	(*RequestVoteResponse)(nil),   // 5: pb.RequestVoteResponse
}
var file_pb_rpc_proto_depIdxs = []int32{
	0, // 0: pb.Raft.ApplyCommand:input_type -> pb.ApplyCommandRequest
	1, // 1: pb.Raft.AppendEntries:input_type -> pb.AppendEntriesRequest
	2, // 2: pb.Raft.RequestVote:input_type -> pb.RequestVoteRequest
	3, // 3: pb.Raft.ApplyCommand:output_type -> pb.ApplyCommandResponse
	4, // 4: pb.Raft.AppendEntries:output_type -> pb.AppendEntriesResponse
	5, // 5: pb.Raft.RequestVote:output_type -> pb.RequestVoteResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_rpc_proto_init() }
func file_pb_rpc_proto_init() {
	if File_pb_rpc_proto != nil {
		return
	}
	file_pb_message_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pb_rpc_proto_goTypes,
		DependencyIndexes: file_pb_rpc_proto_depIdxs,
	}.Build()
	File_pb_rpc_proto = out.File
	file_pb_rpc_proto_rawDesc = nil
	file_pb_rpc_proto_goTypes = nil
	file_pb_rpc_proto_depIdxs = nil
}
