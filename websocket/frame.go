package websocket

import (
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

type Frame struct {
	fin           bool    // 1 bit
	rsv1          bool    // 1 bit
	rsv2          bool    // 1 bit
	rsv3          bool    // 1 bit
	opcode        uint8   // 4 bits
	mask          bool    // 1 bit
	payloadLength uint64  // 7 bits or 7 + 16 bits or 7 + 64 bits
	maskingKey    [4]byte //This field is present if the mask bit is set to 1 and is absent if the mask bit is set to 0.
	payloadData   []byte
}

const (
	continuationFrame = byte(0x0)
	textFrame         = byte(0x1)
	binaryFrame       = byte(0x2)
	//nonControlFrame3  = byte(0x3)
	//nonControlFrame4  = byte(0x4)
	//nonControlFrame5  = byte(0x5)
	//nonControlFrame6  = byte(0x6)
	//nonControlFrame7  = byte(0x7)
	connectionClose = byte(0x8)
	Ping            = byte(0x9)
	controlFrameA   = byte(0xA)
	controlFrameB   = byte(0xB)
	controlFrameC   = byte(0xC)
	controlFrameD   = byte(0xD)
	controlFrameE   = byte(0xE)
	controlFrameF   = byte(0xF)
)

func newFrame(opcode uint8, mask bool, data []byte) *Frame {
	var maskingKey [4]byte
	if mask {
		maskingKey = newMaskingKey()
	}

	return &Frame{
		fin:           false,
		rsv1:          false,
		rsv2:          false,
		rsv3:          false,
		opcode:        opcode,
		mask:          mask,
		payloadLength: uint64(len(data)),
		maskingKey:    maskingKey,
		payloadData:   data,
	}
}

func ParseFrame(data []byte) (*Frame, error) {

	f := &Frame{}

	binary := convertToBinaryRep(data)

	if strings.Compare(string(binary[0]), "1") == 0 {
		f.fin = true
	} else {
		f.fin = false
	}

	if strings.Compare(string(binary[1]), "1") == 0 {
		f.rsv1 = true
	} else {
		f.rsv1 = false
	}

	if strings.Compare(string(binary[2]), "1") == 0 {
		f.rsv2 = true
	} else {
		f.rsv2 = false
	}

	if strings.Compare(string(binary[3]), "1") == 0 {
		f.rsv3 = true
	} else {
		f.rsv3 = false
	}

	opcode64, err := strconv.ParseUint("0000"+binary[4:8], 2, 64)
	if err != nil {
		return nil, err
	}

	f.opcode = uint8(opcode64)

	if strings.Compare(string(binary[8]), "1") == 0 {
		f.mask = true
	} else {
		f.mask = false
	}

	maskingKeyBit := 16

	f.payloadLength, err = strconv.ParseUint("0"+binary[9:16], 2, 64)
	if err != nil {
		return nil, err
	}

	if f.payloadLength == 127 {
		f.payloadLength, err = strconv.ParseUint(binary[16:32], 2, 64)
		if err != nil {
			return nil, err
		}
		maskingKeyBit = 32
	} else if f.payloadLength == 128 {
		f.payloadLength, err = strconv.ParseUint(binary[16:80], 2, 64)
		if err != nil {
			return nil, err
		}
		maskingKeyBit = 80
	}

	if f.mask {
		for i := 0; i < 4; i++ {
			maskingKey, err := strconv.ParseUint(binary[maskingKeyBit+(8*i):maskingKeyBit+(8*(i+1))], 2, 64)
			if err != nil {
				return nil, err
			}
			f.maskingKey[i] = byte(maskingKey)
		}
	}

	f.payloadData = make([]byte, int(f.payloadLength))

	start := maskingKeyBit
	if f.mask {
		start = start + 32
	}

	end := start + 8

	for i := 0; i < int(f.payloadLength); i++ {
		byte64, err := strconv.ParseUint("0"+binary[start:end], 2, 64)
		if err != nil {
			return nil, err
		}
		f.payloadData[i] = byte(byte64)
		start = end
		end = start + 8
	}

	f.payloadData = maskData(f.maskingKey, f.payloadData)

	return f, nil
}

func convertToBinaryRep(data []byte) string {

	var binaryRep string

	for _, byte := range data {
		binaryRep = binaryRep + fmt.Sprintf("%08b", uint8(byte))
	}

	return binaryRep
}

func maskData(maskingKey [4]byte, originalData []byte) []byte {
	transformedData := make([]byte, len(originalData))

	for i, _ := range originalData {
		transformedData[i] = originalData[i] ^ maskingKey[i%4]
	}

	return transformedData
}

func newMaskingKey() (maskingKey [4]byte) {
	for i := 0; i < 4; i++ {
		maskingKey[i] = byte(rand.Intn(255))
	}

	return maskingKey
}

func (f *Frame) encode() ([]byte, error){
	var bitRespresentation string
	
	//if f.fin {
	//	bitRespresentation += "1"
	//} else {
	//	bitRespresentation += "0"
	//}
	bitRespresentation += "1"

	if f.rsv1 {
		bitRespresentation += "1"
	} else {
		bitRespresentation += "0"
	}

	if f.rsv2 {
		bitRespresentation += "1"
	} else {
		bitRespresentation += "0"
	}

	if f.rsv3 {
		bitRespresentation += "1"
	} else {
		bitRespresentation += "0"
	}

	bitRespresentation += fmt.Sprintf("%08b", f.opcode)[4:8]

	if f.mask {
		bitRespresentation += "1"
	} else {
		bitRespresentation += "0"
	}

	if f.payloadLength <= 126 {
		bitRespresentation += fmt.Sprintf("%08b", f.payloadLength)[1:8]
	} else if f.payloadLength <= 65535 {
		bitRespresentation += fmt.Sprintf("%08b", uint8(127))[1:8]
		bitRespresentation += fmt.Sprintf("%016b", f.payloadLength)
	} else {
		bitRespresentation += fmt.Sprintf("%08b", uint8(128))[1:8]
		bitRespresentation += fmt.Sprintf("%064b", f.payloadLength)
	}

	if f.mask {
		for i := 0; i < 4; i++ {
			bitRespresentation += fmt.Sprintf("%08b", f.maskingKey[i])
		}
	}

	var payloadData []byte
	if f.mask {
		payloadData = maskData(f.maskingKey, f.payloadData) 
	}else {
		payloadData = f.payloadData
	}

	byteSlice, err := bitStringToBytes(bitRespresentation)
	if err != nil {
		return nil, err
	} 

	encodedData := bytes.Join([][]byte{byteSlice, payloadData}, nil)

	return encodedData, nil
}


func bitStringToBytes(bitString string) ([]byte, error) {
	numberOfBytes := len(bitString)/8

	var result []byte
	
	firstBit := 0
	lastBit := firstBit + 8

	for i := 0; i < numberOfBytes; i++ {
		a, err := strconv.ParseUint(bitString[firstBit:lastBit], 2, 8)
		if err != nil {
			return nil, err
		}
		result = append(result, byte(a))
		firstBit = lastBit
		lastBit = firstBit + 8
	}

	return result, nil
}