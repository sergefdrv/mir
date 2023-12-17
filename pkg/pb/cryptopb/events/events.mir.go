// Code generated by Mir codegen. DO NOT EDIT.

package cryptopbevents

import (
	types "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

func SignRequest(destModule stdtypes.ModuleID, data *types.SignedData, origin *types.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_SignRequest{
					SignRequest: &types.SignRequest{
						Data:   data,
						Origin: origin,
					},
				},
			},
		},
	}
}

func SignResult(destModule stdtypes.ModuleID, signature []uint8, origin *types.SignOrigin) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_SignResult{
					SignResult: &types.SignResult{
						Signature: signature,
						Origin:    origin,
					},
				},
			},
		},
	}
}

func VerifySig(destModule stdtypes.ModuleID, data *types.SignedData, signature []uint8, origin *types.SigVerOrigin, nodeId stdtypes.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_VerifySig{
					VerifySig: &types.VerifySig{
						Data:      data,
						Signature: signature,
						Origin:    origin,
						NodeId:    nodeId,
					},
				},
			},
		},
	}
}

func SigVerified(destModule stdtypes.ModuleID, origin *types.SigVerOrigin, nodeId stdtypes.NodeID, error error) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_SigVerified{
					SigVerified: &types.SigVerified{
						Origin: origin,
						NodeId: nodeId,
						Error:  error,
					},
				},
			},
		},
	}
}

func VerifySigs(destModule stdtypes.ModuleID, data []*types.SignedData, signatures [][]uint8, origin *types.SigVerOrigin, nodeIds []stdtypes.NodeID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_VerifySigs{
					VerifySigs: &types.VerifySigs{
						Data:       data,
						Signatures: signatures,
						Origin:     origin,
						NodeIds:    nodeIds,
					},
				},
			},
		},
	}
}

func SigsVerified(destModule stdtypes.ModuleID, origin *types.SigVerOrigin, nodeIds []stdtypes.NodeID, errors []error, allOk bool) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Crypto{
			Crypto: &types.Event{
				Type: &types.Event_SigsVerified{
					SigsVerified: &types.SigsVerified{
						Origin:  origin,
						NodeIds: nodeIds,
						Errors:  errors,
						AllOk:   allOk,
					},
				},
			},
		},
	}
}
