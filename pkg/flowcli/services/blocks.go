/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package services

import (
	"fmt"
	"strconv"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-go-sdk/client"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-go-sdk"
)

// Blocks service handles all interactions for blocks
type Blocks struct {
	gateway gateway.Gateway
	project *project.Project
	logger  output.Logger
}

// NewBlocks create new block service
func NewBlocks(
	gateway gateway.Gateway,
	project *project.Project,
	logger output.Logger,
) *Blocks {
	return &Blocks{
		gateway: gateway,
		project: project,
		logger:  logger,
	}
}

// Get the block
func (e *Blocks) GetBlock(
	query string,
	eventType string,
	verbose bool,
) (*flow.Block, []client.BlockEvents, []*flow.Collection, error) {
	e.logger.StartProgress("Fetching Block...")

	// smart parsing of query
	var err error
	var block *flow.Block
	if query == "latest" {
		block, err = e.gateway.GetLatestBlock()
	} else if height, ce := strconv.ParseUint(query, 10, 64); ce == nil {
		block, err = e.gateway.GetBlockByHeight(height)
	} else {
		block, err = e.gateway.GetBlockByID(flow.HexToID(query))
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("error fetching block: %s", err.Error())
	}

	if block == nil {
		return nil, nil, nil, fmt.Errorf("block not found")
	}

	// if we specify event get events by the type
	var events []client.BlockEvents
	if eventType != "" {
		events, err = e.gateway.GetEvents(eventType, block.Height, block.Height)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	// if verbose fetch all collections from block too
	collections := make([]*flow.Collection, 0)
	if verbose {
		for _, guarantee := range block.CollectionGuarantees {
			collection, err := e.gateway.GetCollection(guarantee.CollectionID)
			if err != nil {
				return nil, nil, nil, err
			}
			collections = append(collections, collection)
		}
	}

	e.logger.StopProgress()

	return block, events, collections, err
}