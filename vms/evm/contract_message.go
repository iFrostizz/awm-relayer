// Copyright (C) 2023, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package evm

import (
	"errors"

	"github.com/ava-labs/avalanchego/utils/logging"
	avalancheWarp "github.com/ava-labs/avalanchego/vms/platformvm/warp"
	warpPayload "github.com/ava-labs/avalanchego/vms/platformvm/warp/payload"
	"github.com/ava-labs/awm-relayer/config"
	"github.com/ava-labs/awm-relayer/vms/vmtypes"
	"github.com/ava-labs/subnet-evm/precompile/contracts/warp"
	"go.uber.org/zap"
)

type contractMessage struct {
	logger logging.Logger
}

func NewContractMessage(logger logging.Logger, subnetInfo config.SourceSubnet) *contractMessage {
	return &contractMessage{
		logger: logger,
	}
}

// Unpacks an AddressedCall payload from an unsigned Warp message.
// unsignedMsgBytes may be either the raw UnsignedMessage bytes or the ABI encoded bytes as emitted by the Warp precompile
func (m *contractMessage) UnpackWarpMessage(unsignedMsgBytes []byte) (*vmtypes.WarpMessageInfo, error) {
	// This function may be called with raw UnsignedMessage bytes or with ABI encoded bytes as emitted by the Warp precompile
	// The latter case is the steady state behavior, so check that first. The former only occurs on startup.
	unsignedMsg, err := warp.UnpackSendWarpEventDataToMessage(unsignedMsgBytes)
	if err != nil {
		m.logger.Debug(
			"Failed parsing unsigned message as log. Attempting to parse as standalone message",
			zap.Error(err),
		)
		var standaloneErr error
		unsignedMsg, standaloneErr = avalancheWarp.ParseUnsignedMessage(unsignedMsgBytes)
		if standaloneErr != nil {
			err = errors.Join(err, standaloneErr)
			m.logger.Error(
				"Failed parsing unsigned message as either log or standalone message",
				zap.Error(err),
			)
			return nil, err
		}
	}

	warpPayload, err := warpPayload.ParseAddressedCall(unsignedMsg.Payload)
	if err != nil {
		m.logger.Error(
			"Failed parsing addressed payload",
			zap.Error(err),
		)
		return nil, err
	}

	messageInfo := vmtypes.WarpMessageInfo{
		WarpUnsignedMessage: unsignedMsg,
		WarpPayload:         warpPayload.Payload,
	}
	return &messageInfo, nil
}
