// Copyright 2020 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/proxy-wasm-go-sdk/proxytest"
	"github.com/tetratelabs/proxy-wasm-go-sdk/proxywasm/types"
)

func TestNetwork_OnNewConnection(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewStreamContext(newNetworkContext).
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.StartVM() // call OnVMStart: init metric

	_ = host.NetworkFilterInitConnection() // OnNewConnection is called

	logs := host.GetLogs(types.LogLevelInfo) // retrieve logs emitted to Envoy
	assert.Equal(t, logs[0], "new connection!")
}

func TestNetwork_OnDownstreamClose(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewStreamContext(newNetworkContext).
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	contextID := host.NetworkFilterInitConnection()        // OnNewConnection is called
	host.NetworkFilterCloseDownstreamConnection(contextID) // OnDownstreamClose is called

	logs := host.GetLogs(types.LogLevelInfo) // retrieve logs emitted to Envoy
	require.Len(t, logs, 2)
	assert.Equal(t, logs[1], "downstream connection close!")
}

func TestNetwork_OnDownstreamData(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewStreamContext(newNetworkContext).
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	contextID := host.NetworkFilterInitConnection() // OnNewConnection is called

	msg := "this is downstream data"
	data := []byte(msg)
	host.NetworkFilterPutDownstreamData(contextID, data) // OnDownstreamData is called

	logs := host.GetLogs(types.LogLevelInfo) // retrieve logs emitted to Envoy
	assert.Equal(t, ">>>>>> downstream data received >>>>>>\n"+msg, logs[len(logs)-1])
}

func TestNetwork_OnUpstreamData(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewStreamContext(newNetworkContext).
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	contextID := host.NetworkFilterInitConnection() // OnNewConnection is called

	msg := "this is upstream data"
	data := []byte(msg)
	host.NetworkFilterPutUpstreamData(contextID, data) // OnUpstreamData is called

	logs := host.GetLogs(types.LogLevelInfo) // retrieve logs emitted to Envoy
	assert.Equal(t, "<<<<<< upstream data received <<<<<<\n"+msg, logs[len(logs)-1])
}

func TestNetwork_counter(t *testing.T) {
	opt := proxytest.NewEmulatorOption().
		WithNewStreamContext(newNetworkContext).
		WithNewRootContext(newRootContext)
	host := proxytest.NewHostEmulator(opt)
	defer host.Done() // release the host emulation lock so that other test cases can insert their own host emulation

	host.StartVM() // call OnVMStart: init metric

	contextID := host.NetworkFilterInitConnection()
	host.NetworkFilterCompleteConnection(contextID) // call OnStreamDone on contextID -> increment the connection counter

	logs := host.GetLogs(types.LogLevelInfo)
	require.Greater(t, len(logs), 0)

	assert.Equal(t, "connection complete!", logs[len(logs)-1])
	actual := counter.Get()
	assert.Equal(t, uint64(1), actual)
}
