#include "ber.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

const unsigned char snmp_get_request[] = {
    0x30, 0x2a,             // Sequence (Message)
    0x02, 0x01, 0x00,       // Version: SNMPv2c (INTEGER, 0)
    0x04, 0x06,             // Community String (OCTET STRING)
    'p', 'u', 'b', 'l', 'i', 'c',
    0xa0, 0x1d,             // GetRequest PDU (0xA0 = GetRequest)
    0x02, 0x04, 0x30, 0x39, 0x38, 0x36,  // Request ID (INTEGER)
    0x02, 0x01, 0x00,       // Error Status: No error (INTEGER, 0)
    0x02, 0x01, 0x00,       // Error Index: 0 (INTEGER, 0)
    0x30, 0x0f,             // Variable Bindings Sequence
    0x30, 0x0d,             // Single VarBind
    0x06, 0x09,             // Object Identifier (OID)
    0x2b, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00, // OID: 1.3.6.1.2.1.1.1.0
    0x05, 0x00              // NULL value (End of VarBind)
};


void main()
{
    sid s;
    uchar tag;
    uint x;
    uchar* req = snmp_get_request;
    req = parse_sid(&s, snmp_get_request);
    switch (s) {
    case APPLICATION:
        printf("APPLICATION\n");
        break;
    case UNIVERSAL:
        printf("UNIVERSAL\n");
        break;
    case CONTEXTSPECIFIC:
        printf("CONTEXTSPECIFIC\n");
        break;
    case PRIVATE:
        printf("PRIVATE\n");
        break;
    default:
        printf("Dont know.\n");
        break;
    }

    printf("Hello, World!\n");
}