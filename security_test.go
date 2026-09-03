// Copyright (c) 2026 ABUZUCOM LLC
// SPDX-License-Identifier: BSD-3-Clause

package prolink

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestDeviceFromAnnouncePacketRejectsShortPacket(t *testing.T) {
	_, err := deviceFromAnnouncePacket(prolinkHeader)
	if err == nil {
		t.Fatal("expected short announcement to be rejected")
	}
}

func TestReadFieldRejectsOversizedString(t *testing.T) {
	packet := []byte{fieldTypeString, 0, 0, 0, 0}
	_, err := readField(bytes.NewReader(packet))
	if err == nil {
		t.Fatal("expected zero-length string to be rejected")
	}

	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, maxRemoteStringUnits+1)
	packet = append([]byte{fieldTypeString}, length...)
	_, err = readField(bytes.NewReader(packet))
	if err == nil {
		t.Fatal("expected oversized string to be rejected")
	}
}

func TestReadFieldRejectsOversizedBinary(t *testing.T) {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, maxRemoteBinaryBytes+1)
	packet := append([]byte{fieldTypeBinary}, length...)
	_, err := readField(bytes.NewReader(packet))
	if err == nil {
		t.Fatal("expected oversized binary to be rejected")
	}
}

func TestReadMessagePacketRejectsInvalidFieldTypes(t *testing.T) {
	packet := []byte{}
	appendNumber := func(fieldType byte, value []byte) {
		packet = append(packet, fieldType)
		packet = append(packet, value...)
	}
	appendNumber(fieldTypeNumber04, []byte{0x87, 0x23, 0x49, 0xae})
	appendNumber(fieldTypeNumber04, []byte{0, 0, 0, 1})
	appendNumber(fieldTypeNumber01, []byte{0})
	appendNumber(fieldTypeNumber01, []byte{0})
	appendNumber(fieldTypeBinary, []byte{0, 0, 0, 0})

	_, err := readMessagePacket(bytes.NewReader(packet))
	if err == nil {
		t.Fatal("expected invalid message type to be rejected")
	}
}
