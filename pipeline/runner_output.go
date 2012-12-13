/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Rob Miller (rmiller@mozilla.com)
#
# ***** END LICENSE BLOCK *****/
package pipeline

import (
	"encoding/json"
	"errors"
)

// Interface for output objects that need to share a global resource (such as
// a file handle or network connection) to actually emit the output data.
type OutputWriter interface {
	PluginGlobal

	// Setup method, called exactly once
	Init(config interface{}) error

	// This must create exactly one instance of the `outData` data object type
	// expected by the `Write` method. Will be called multiple times to create
	// a pool of reusable objects.
	MakeOutData() *interface{}

	// Will be handed a used output object which should be reset to a zero
	// state for in preparation for reuse. This method will be in use by
	// multiple goroutines simultaneously, it should modify the passed
	// `outData` object **only**.
	ZeroOutData(outData *interface{})

	// Extracts relevant information from the provided `PipelinePack`
	// (probably from the `Message` attribute) and uses it to populate the
	// provided output object. This method will be in use by multiple
	// goroutines simultaneously, it should modify the passed `emptyOutData`
	// object **only**.
	PrepOutData(pack *PipelinePack, emptyOutData *interface{})

	// Receives a populated output object, handles the actual work of writing
	// data out to an external destination.
	Write(outData *interface{}) error
}

// Output plugin that drives an OutputWriter
type RunnerOutput struct {
	Writer   OutputWriter
	outChan  chan interface{}
	backChan chan interface{}
}

func RunnerOutputMaker(writer OutputWriter) func() *RunnerOutput {
	return func() *RunnerOutput { return &RunnerOutput{Writer: writer} }
}

func (self *RunnerOutput) InitOnce(config interface{}) (global PluginGlobal, err error) {
	conf := config.(*PluginConfig)
	confLoaded, err := LoadConfigStruct(conf, self.Writer)
	if err != nil {
		return self.Writer, errors.New("WriteRunner config parsing error: ", err)
	}
	if err = self.Writer.Init(confLoaded); err != nil {
		return self.Writer, errors.New("WriteRunner initialization error: ", err)
	}
	self.outChan = make(chan interface{}, 2*PoolSize)
	self.backChan = make(chan interface{}, 2*PoolSize)
	return self.Writer, nil
}

func (self *RunnerOutput) Init(global PluginGlobal, config interface{}) error {
	return nil
}
