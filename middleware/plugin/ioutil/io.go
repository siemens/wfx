package ioutil

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"encoding/binary"
	"errors"
	"io"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/siemens/wfx/generated/plugin"
)

// buffer size large enough for typical requests
const initialSize = 1 << 14

type Packer interface {
	Pack(builder *flatbuffers.Builder) flatbuffers.UOffsetT
}

// ReadRequest reads a request from the provided io.Reader.
func ReadRequest(r io.Reader) (*plugin.PluginRequestT, error) {
	buf, err := readBytes(r)
	if err != nil {
		return nil, err
	}
	return plugin.GetRootAsPluginRequest(buf, 0).UnPack(), nil
}

// ReadResponse reads a response from the provided io.Reader.
func ReadResponse(r io.Reader) (*plugin.PluginResponseT, error) {
	buf, err := readBytes(r)
	if err != nil {
		return nil, err
	}
	return plugin.GetRootAsPluginResponse(buf, 0).UnPack(), nil
}

// WriteRequest writes the given request to an io.Writer.
func WriteRequest(w io.Writer, req *plugin.PluginRequestT) error {
	return writeHelper(w, req)
}

// WriteResponse writes the given request to an io.Writer.
func WriteResponse(w io.Writer, resp *plugin.PluginResponseT) error {
	return writeHelper(w, resp)
}

func readPrefix(r io.Reader) (uint32, error) {
	// see https://github.com/dvidelabs/flatcc/blob/master/doc/binary-format.md
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func readBytes(r io.Reader) ([]byte, error) {
	size, err := readPrefix(r)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeHelper(w io.Writer, packer Packer) error {
	builder := flatbuffers.NewBuilder(initialSize)
	end := packer.Pack(builder)
	builder.FinishSizePrefixed(end)
	buf := builder.FinishedBytes()
	n, err := w.Write(buf)
	if n != len(buf) {
		return errors.New("incomplete write")
	}
	return err
}
