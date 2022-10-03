package threshcryptopbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	events "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/events"
	types1 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for emitting events.

func SignShare(m dsl.Module, destModule types.ModuleID, data [][]uint8, origin *types1.SignShareOrigin) {
	dsl.EmitMirEvent(m, events.SignShare(destModule, data, origin))
}

func SignShareResult(m dsl.Module, destModule types.ModuleID, signatureShare []uint8, origin *types1.SignShareOrigin) {
	dsl.EmitMirEvent(m, events.SignShareResult(destModule, signatureShare, origin))
}

func VerifyShare(m dsl.Module, destModule types.ModuleID, data [][]uint8, signatureShare []uint8, nodeId types.NodeID, origin *types1.VerifyShareOrigin) {
	dsl.EmitMirEvent(m, events.VerifyShare(destModule, data, signatureShare, nodeId, origin))
}

func VerifyShareResult(m dsl.Module, destModule types.ModuleID, ok bool, error string, origin *types1.VerifyShareOrigin) {
	dsl.EmitMirEvent(m, events.VerifyShareResult(destModule, ok, error, origin))
}

func VerifyFull(m dsl.Module, destModule types.ModuleID, data [][]uint8, fullSignature []uint8, origin *types1.VerifyFullOrigin) {
	dsl.EmitMirEvent(m, events.VerifyFull(destModule, data, fullSignature, origin))
}

func VerifyFullResult(m dsl.Module, destModule types.ModuleID, ok bool, error string, origin *types1.VerifyFullOrigin) {
	dsl.EmitMirEvent(m, events.VerifyFullResult(destModule, ok, error, origin))
}

func Recover(m dsl.Module, destModule types.ModuleID, data [][]uint8, signatureShares [][]uint8, origin *types1.RecoverOrigin) {
	dsl.EmitMirEvent(m, events.Recover(destModule, data, signatureShares, origin))
}

func RecoverResult(m dsl.Module, destModule types.ModuleID, fullSignature []uint8, ok bool, error string, origin *types1.RecoverOrigin) {
	dsl.EmitMirEvent(m, events.RecoverResult(destModule, fullSignature, ok, error, origin))
}
