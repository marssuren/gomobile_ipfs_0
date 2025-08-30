package main

import (
	proximity "github.com/marssuren/gomobile_ipfs_0/go/pkg/proximitytransport"
)

type ProximityDriver interface {
	proximity.ProximityDriver
}

type ProximityTransport interface {
	proximity.ProximityTransport
}
