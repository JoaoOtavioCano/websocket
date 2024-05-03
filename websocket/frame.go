package websocket

import (
	"fmt"
	"strconv"
	"strings"
)

type Frame struct {
	fin           bool   // 1 bit
	rsv1          bool   // 1 bit
	rsv2          bool   // 1 bit
	rsv3          bool   // 1 bit
	opcode        uint8  // 4 bits
	mask          bool   // 1 bit
	payloadLength uint64 // 7 bits or 7 + 16 bits or 7 + 64 bits
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
		for i := 0; i < 4; i++{
			maskingKey, err := strconv.ParseUint(binary[maskingKeyBit+(8*i):maskingKeyBit+(8*(i+1))], 2, 64)
			if err != nil {
				return nil, err
			}
			f.maskingKey[i] = byte(maskingKey)
		}
	}

	f.payloadData = make([]byte, int(f.payloadLength))

	fmt.Printf("fin: %t\n", f.fin)
	fmt.Printf("rsv1: %t\n", f.rsv1)
	fmt.Printf("rsv2: %t\n", f.rsv2)
	fmt.Printf("rsv3: %t\n", f.rsv3)
	fmt.Printf("opcode: %d\n", f.opcode)
	fmt.Printf("mask: %t\n", f.mask)
	fmt.Printf("payload length: %d\n", f.payloadLength)
	fmt.Printf("masking key: %d\n", f.maskingKey)

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

func maskData(maskingKey [4]byte, originalData []byte) []byte{
	transformedData := make([]byte, len(originalData))

	for i, _ := range originalData {
		transformedData[i] =  originalData[i] ^ maskingKey[i%4]
	}

	return transformedData
}
