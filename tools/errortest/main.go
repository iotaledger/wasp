package main

import (
	"bytes"
	"compress/flate"
	"compress/lzw"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/hive.go/marshalutil"
)

func serializeToJson(args ...interface{}) []byte {
	bytes, _ := json.Marshal(args)

	return bytes
}

func serializeToGob(a string, b int, c float64) []byte {
	var result bytes.Buffer
	gob.Register(&Args{})
	gob.NewEncoder(&result).Encode(Args{String: a, Int: b, Float: c})
	return result.Bytes()
}

func serializeToMarshalUtil(a string, b int, c float64) []byte {
	m := marshalutil.New()

	m.WriteUint16(uint16(len(a)))
	m.WriteBytes([]byte(a))
	m.WriteInt32(int32(b))
	m.WriteFloat64(c)

	return m.Bytes()
}

func serializeToMarshalUtilJustString(args ...interface{}) []byte {
	m := marshalutil.New()

	for _, v := range args {
		str := fmt.Sprintf("%v", v)
		m.WriteUint16(uint16(len(str)))
		m.WriteBytes([]byte(str))
	}

	return m.Bytes()
}

func DeserializeFromMarshalString(str []byte) []string {
	args := make([]string, 0)
	m := marshalutil.New(str)
	end := false

	for {
		strlen, _ := m.ReadInt16()
		str, _ := m.ReadBytes(int(strlen))

		args = append(args, string(str))

		end, _ = m.DoneReading()

		if end {
			break
		}
	}

	return args
}

func compressFlate(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w, err := flate.NewWriter(&b, 9)
	if err != nil {
		return nil, err
	}
	w.Write(data)
	w.Close()
	return b.Bytes(), nil
}

func compressLZW(data []byte) []byte {
	buf := new(bytes.Buffer)
	w := lzw.NewWriter(buf, lzw.LSB, 7)
	w.Write(data)
	w.Close()
	return buf.Bytes()
}

type Args struct {
	String string
	Int    int
	Float  float64
}

func main() {
	argA := "This is rather long"
	argB := 42
	argC := 42.13375

	fmt.Printf("test %x\n", fmt.Sprintf("%v", argC))

	a := serializeToJson(argA, argB, argC)
	b := serializeToGob(argA, argB, argC)
	c := serializeToMarshalUtil(argA, argB, argC)
	d := serializeToMarshalUtilJustString(argA, argB, argC)

	fmt.Printf("a:%v, b:%v, c:%v, d:%v \n", len(a), len(b), len(c), len(d))

	fmt.Print(string(a))
	fmt.Print(string(d))

	result := DeserializeFromMarshalString(d)

	fmt.Print(result)
}
