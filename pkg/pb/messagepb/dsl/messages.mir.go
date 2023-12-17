// Code generated by Mir codegen. DO NOT EDIT.

package messagepbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

// Module-specific dsl functions for processing net messages.

func UponMessageReceived[W types.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from stdtypes.NodeID, msg *M) error) {
	dsl1.UponMessageReceived(m, func(from stdtypes.NodeID, msg *types.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}
