//go:build !darwin && !android
// +build !darwin,!android

package ble

import (
	proximity "github.com/marssuren/gomobile_ipfs_0/go/pkg/proximitytransport"

	"go.uber.org/zap"
)

const Supported = false

// Noop implementation for platform that are not Darwin

func NewDriver(logger *zap.Logger) proximity.ProximityDriver {
	logger = logger.Named("BLE")
	logger.Info("NewDriver(): incompatible system")

	return proximity.NewNoopProximityDriver(ProtocolCode, ProtocolName, DefaultAddr)
}
