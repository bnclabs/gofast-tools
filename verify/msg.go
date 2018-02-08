package main

import "fmt"
import "encoding/binary"

import "github.com/bnclabs/gofast"
import "github.com/bnclabs/gson"

//-- post message for benchmarking

type msgPost struct {
	data []byte
}

func newMsgPost(data []byte) *msgPost {
	return &msgPost{data: data}
}

func (msg *msgPost) ID() uint64 {
	return 111
}

func (msg *msgPost) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	binary.BigEndian.PutUint64(out, uint64(len(msg.data)))
	n := 8
	n += copy(out[n:], msg.data)
	return out[:n]
}

func (msg *msgPost) Decode(in []byte) int64 {
	ln, n := int64(binary.BigEndian.Uint64(in)), int64(8)
	msg.data = fixbuffer(msg.data, ln)
	n += int64(copy(msg.data, in[n:n+ln]))
	return n
}

func (msg *msgPost) Size() int64 {
	return 8 + int64(len(msg.data))
}

func (msg *msgPost) String() string {
	return "msgPost"
}

//-- reqsp message for benchmarking

type msgReqsp struct {
	data []byte
}

func (msg *msgReqsp) ID() uint64 {
	return 112
}

func (msg *msgReqsp) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	binary.BigEndian.PutUint64(out, uint64(len(msg.data)))
	n := 8
	n += copy(out[n:], msg.data)
	return out[:n]
}

func (msg *msgReqsp) Decode(in []byte) int64 {
	ln, n := int64(binary.BigEndian.Uint64(in)), int64(8)
	msg.data = fixbuffer(msg.data, ln)
	n += int64(copy(msg.data, in[n:n+ln]))
	return n
}

func (msg *msgReqsp) Size() int64 {
	return 8 + int64(len(msg.data))
}

func (msg *msgReqsp) String() string {
	return "msgReqsp"
}

//-- stream-start message for benchmarking

type msgStreamStart struct {
	data []byte
}

func (msg *msgStreamStart) ID() uint64 {
	return 113
}

func (msg *msgStreamStart) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	binary.BigEndian.PutUint64(out, uint64(len(msg.data)))
	n := 8
	n += copy(out[n:], msg.data)
	return out[:n]
}

func (msg *msgStreamStart) Decode(in []byte) int64 {
	ln, n := int64(binary.BigEndian.Uint64(in)), int64(8)
	msg.data = fixbuffer(msg.data, ln)
	n += int64(copy(msg.data, in[n:n+ln]))
	return n
}

func (msg *msgStreamStart) Size() int64 {
	return 8 + int64(len(msg.data))
}

func (msg *msgStreamStart) String() string {
	return "msgStreamStart"
}

//-- stream message for benchmarking

type msgStream struct {
	data []byte
}

func (msg *msgStream) ID() uint64 {
	return 114
}

func (msg *msgStream) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	binary.BigEndian.PutUint64(out, uint64(len(msg.data)))
	n := 8
	n += copy(out[n:], msg.data)
	return out[:n]
}

func (msg *msgStream) Decode(in []byte) int64 {
	ln, n := int64(binary.BigEndian.Uint64(in)), int64(8)
	msg.data = fixbuffer(msg.data, ln)
	n += int64(copy(msg.data, in[n:n+ln]))
	return n
}

func (msg *msgStream) Size() int64 {
	return 8 + int64(len(msg.data))
}

func (msg *msgStream) String() string {
	return "msgStream"
}

//-- done message for benchmarking

type msgDone struct {
	data []byte
}

func newMsgDone(data []byte) *msgDone {
	return &msgDone{data: data}
}

func (msg *msgDone) ID() uint64 {
	return 115
}

func (msg *msgDone) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	binary.BigEndian.PutUint64(out, uint64(len(msg.data)))
	n := 8
	n += copy(out[n:], msg.data)
	return out[:n]
}

func (msg *msgDone) Decode(in []byte) int64 {
	ln, n := int64(binary.BigEndian.Uint64(in)), int64(8)
	msg.data = fixbuffer(msg.data, ln)
	n += int64(copy(msg.data, in[n:n+ln]))
	return n
}

func (msg *msgDone) Size() int64 {
	return 8 + int64(len(msg.data))
}

func (msg *msgDone) String() string {
	return "msgDone"
}

//-- stat message for benchmarking

type msgStat struct {
	data map[string]uint64
}

func newMsgStat(data map[string]uint64) *msgStat {
	return &msgStat{data: data}
}

func (msg *msgStat) ID() uint64 {
	return 116
}

func (msg *msgStat) Encode(out []byte) []byte {
	out = fixbuffer(out, msg.Size())
	config := gson.NewDefaultConfig().SetNumberKind(gson.SmartNumber)
	cbr := config.NewCbor(out[:0])
	data := make(map[string]interface{})
	for k, v := range msg.data {
		data[k] = v
	}
	config.NewValue(data).Tocbor(cbr)
	return cbr.Bytes()
}

func (msg *msgStat) Decode(in []byte) (n int64) {
	config := gson.NewDefaultConfig().SetNumberKind(gson.SmartNumber)
	cbr := config.NewCbor(in)
	msg.data = make(map[string]uint64)
	for k, v := range cbr.Tovalue().(map[string]interface{}) {
		msg.data[k] = v.(uint64)
	}
	return int64(len(in))
}

func (msg *msgStat) Size() int64 {
	return 10 * 1024
}

func (msg *msgStat) String() string {
	return "msgStat"
}

//-- version

type testVersion int

func (v *testVersion) Less(ver gofast.Version) bool {
	return (*v) < (*ver.(*testVersion))
}

func (v *testVersion) Equal(ver gofast.Version) bool {
	return (*v) == (*ver.(*testVersion))
}

func (v *testVersion) String() string {
	return fmt.Sprintf("%v", int(*v))
}

func (v *testVersion) Encode(out []byte) []byte {
	out = fixbuffer(out, 32)
	n := valuint642cbor(uint64(*v), out)
	return out[:n]
}

func (v *testVersion) Size() int64 {
	return 9
}

func (v *testVersion) Decode(in []byte) int64 {
	ln, n := cborItemLength(in)
	*v = testVersion(ln)
	return int64(n)
}
