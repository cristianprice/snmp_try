package main

import (
	"io"
	"math/big"
)

const (
	ASN_BOOLEAN     byte = 0x01
	ASN_INTEGER     byte = 0x02
	ASN_BIT_STR     byte = 0x03
	ASN_OCTET_STR   byte = 0x04
	ASN_NULL        byte = 0x05
	ASN_OBJECT_ID   byte = 0x06
	ASN_SEQUENCE    byte = 0x10
	ASN_SET         byte = 0x11
	ASN_UNIVERSAL   byte = 0x00
	ASN_APPLICATION byte = 0x40
	ASN_CONTEXT     byte = 0x80
	ASN_PRIVATE     byte = 0xC0
	ASN_PRIMITIVE   byte = 0x00
	ASN_CONSTRUCTOR byte = 0x20

	ASN_LONG_LEN     byte = 0x80
	ASN_EXTENSION_ID byte = 0x1F
	ASN_BIT8         byte = 0x80

	INTEGER     byte = ASN_UNIVERSAL | 0x02
	INTEGER32   byte = ASN_UNIVERSAL | 0x02
	BITSTRING   byte = ASN_UNIVERSAL | 0x03
	OCTETSTRING byte = ASN_UNIVERSAL | 0x04
	NULL        byte = ASN_UNIVERSAL | 0x05
	OID         byte = ASN_UNIVERSAL | 0x06
	SEQUENCE    byte = ASN_CONSTRUCTOR | 0x10

	IPADDRESS byte = ASN_APPLICATION
	COUNTER   byte = ASN_APPLICATION | 0x01
	COUNTER32 byte = ASN_APPLICATION | 0x01
	GAUGE     byte = ASN_APPLICATION | 0x02
	GAUGE32   byte = ASN_APPLICATION | 0x02
	TIMETICKS byte = ASN_APPLICATION | 0x03
	OPAQUE    byte = ASN_APPLICATION | 0x04
	COUNTER64 byte = ASN_APPLICATION | 0x06

	NOSUCHOBJECT   byte = 0x80
	NOSUCHINSTANCE byte = 0x81
	ENDOFMIBVIEW   byte = 0x82

	LENMASK        byte = 0x0ff
	MAX_OID_LENGTH byte = 127
)

type MutableByte struct {
	Value byte
}

func EncodeLength(os io.Writer, length int) error {
	if length < 0 {
		if _, err := os.Write(
			[]byte{
				0x04 | ASN_LONG_LEN,
				byte((length >> 24) & 0xFF),
				byte((length >> 16) & 0xFF),
				byte((length >> 8) & 0xFF),
				byte(length & 0xFF),
			}); err != nil {
			return err
		}
	} else if length < 0x80 {
		if _, err := os.Write([]byte{byte(length)}); err != nil {
			return err
		}
	} else if length <= 0xFF {
		if _, err := os.Write([]byte{0x01 | ASN_LONG_LEN, byte(length)}); err != nil {
			return err
		}
	} else if length <= 0xFFFF { /* 0xFF < length <= 0xFFFF */
		if _, err := os.Write([]byte{0x02 | ASN_LONG_LEN, byte((length >> 8) & 0xFF), byte(length & 0xFF)}); err != nil {
			return err
		}
	} else if length <= 0xFFFFFF { /* 0xFFFF < length <= 0xFFFFFF */
		if _, err := os.Write([]byte{0x03 | ASN_LONG_LEN, byte((length >> 16) & 0xFF), byte((length >> 8) & 0xFF), byte(length & 0xFF)}); err != nil {
			return err
		}
	} else {
		if _, err := os.Write([]byte{0x04 | ASN_LONG_LEN, byte((length >> 24) & 0xFF), byte((length >> 16) & 0xFF), byte((length >> 8) & 0xFF), byte(length & 0xFF)}); err != nil {
			return err
		}
	}
	return nil
}

func EncodeLengthWithNumLenBytes(os io.Writer, length int, numLengthBytes int) error {
	if _, err := os.Write([]byte{byte(numLengthBytes) | ASN_LONG_LEN}); err != nil {
		return err
	}
	for i := (numLengthBytes - 1) * 8; i >= 0; i -= 8 {
		if _, err := os.Write([]byte{byte((length >> i) & 0xFF)}); err != nil {
			return err
		}
	}
	return nil
}

func EncodeInteger(os io.Writer, _type byte, value int) error {
	integer := value
	mask := 0x1FF << ((8 * 3) - 1)
	intsize := 4

	/*
	 * Truncate "unnecessary" bytes off of the most significant end of this
	 * 2's complement integer.  There should be no sequence of 9
	 * consecutive 1's or 0's at the most significant end of the
	 * integer.
	 */
	for ((integer&mask) == 0 || (integer&mask) == mask) && intsize > 1 {
		intsize--
		integer <<= 8
	}
	if err := EncodeHeader(os, _type, intsize); err != nil {
		return err
	}
	mask = 0xFF << (8 * 3)
	/* mask is 0xFF000000 on a big-endian machine */
	for intsize > 0 {
		if _, err := os.Write([]byte{byte((integer & mask) >> (8 * 3))}); err != nil {
			return err
		}
		integer <<= 8
		intsize--
	}
	return nil
}

func EncodeHeader(os io.Writer, _type byte, length int) error {
	if _, err := os.Write([]byte{_type}); err != nil {
		return err
	}
	return EncodeLength(os, length)
}

func EncodeHeaderWithNumBytesLen(os io.Writer, typ byte, length int, numBytesLength int) error {
	if _, err := os.Write([]byte{typ}); err != nil {
		return err
	}
	return EncodeLengthWithNumLenBytes(os, length, numBytesLength)
}

func GetBERLengthOfLength(length int) int {
	if length < 0 {
		return 5
	} else if length < 0x80 {
		return 1
	} else if length <= 0xFF {
		return 2
	} else if length <= 0xFFFF { /* 0xFF < length <= 0xFFFF */
		return 3
	} else if length <= 0xFFFFFF { /* 0xFFFF < length <= 0xFFFFFF */
		return 4
	}
	return 5
}

func EncodeBigInteger(os io.Writer, _type byte, value *big.Int) error {
	bytes := value.Bytes()
	if err := EncodeHeader(os, _type, len(bytes)); err != nil {
		return err
	}
	if _, err := os.Write(bytes); err != nil {
		return err
	}
	return nil
}

func GetBigIntegerBERLength(value *big.Int) int {
	length := len(value.Bytes())
	return length + GetBERLengthOfLength(length) + 1
}

func EncodeUnsignedInteger(os io.Writer, _type byte, value uint64) error {
	// figure out the len
	len := 1
	if (value>>24)&uint64(LENMASK) != 0 {
		len = 4
	} else if (value>>16)&uint64(LENMASK) != 0 {
		len = 3
	} else if (value>>8)&uint64(LENMASK) != 0 {
		len = 2
	}
	// check for 5 byte len where first byte will be a null
	if (value>>(8*(len-1)))&0x080 != 0 {
		len++
	}

	// build up the header
	if err := EncodeHeader(os, _type, len); err != nil { // length of BER encoded item
		return err
	}

	// special case, add a null byte for len of 5
	if len == 5 {
		if _, err := os.Write([]byte{0}); err != nil {
			return err
		}
		for x := 1; x < len; x++ {
			if _, err := os.Write([]byte{byte(value >> (8 * (4 - x)) & uint64(LENMASK))}); err != nil {
				return err
			}
		}
	} else {
		for x := 0; x < len; x++ {
			if _, err := os.Write([]byte{byte((value >> (8 * ((len - 1) - x))) & uint64(LENMASK))}); err != nil {
				return err
			}
		}
	}
	return nil
}

func EncodeString(os io.Writer, _type byte, str []byte) error {
	/*
	* ASN.1 octet string ::= primstring | cmpdstring
	* primstring ::= 0x04 asnlength byte {byte}*
	* cmpdstring ::= 0x24 asnlength string {string}*
	* This code will never send a compound string.
	 */
	if err := EncodeHeader(os, _type, len(str)); err != nil {
		return err
	}
	// fixed
	if _, err := os.Write(str); err != nil {
		return err
	}
	return nil
}
