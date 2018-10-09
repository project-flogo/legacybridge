package legacybridge

import (
	"context"

	oldaction "github.com/TIBCOSoftware/flogo-lib/core/action"
	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/data/metadata"
)

func RegisterLegacyAction(ref string, legacyFactory oldaction.Factory) {
	action.LegacyRegister(ref, &legacyFactoryWrapper{legacyFactory: legacyFactory})
}

type legacyFactoryWrapper struct {
	legacyFactory oldaction.Factory
}

func (legacyFactoryWrapper) Initialize(ctx action.InitContext) error {
	//ignore
	return nil
}

func (w *legacyFactoryWrapper) New(config *action.Config) (action.Action, error) {

	oldCfg := &oldaction.Config{}
	oldCfg.Ref = config.Ref
	oldCfg.Settings = config.Settings
	oldCfg.Id = config.Id

	legacyAct, err := w.legacyFactory.New(oldCfg)
	if err != nil {
		return nil, err
	}

	var wa action.Action

	if act, ok := legacyAct.(oldaction.AsyncAction); ok {
		wa = wrapAsyncAction(act)
	} else if act, ok := legacyAct.(oldaction.SyncAction); ok {
		wa = wrapSyncAction(act)
	}

	return wa, nil
}

func wrapAsyncAction(legacyAct oldaction.AsyncAction) action.AsyncAction {

	aw := &asyncActWrapper{legacyAct: legacyAct}
	//todo wrap metadata
	return aw
}

func wrapSyncAction(legacyAct oldaction.SyncAction) action.SyncAction {

	aw := &syncActWrapper{legacyAct: legacyAct}
	//todo wrap metadata
	return aw
}

type asyncActWrapper struct {
	legacyAct oldaction.AsyncAction
}

func (asyncActWrapper) Metadata() *action.Metadata {
	panic("implement me")
}

func (asyncActWrapper) IOMetadata() *metadata.IOMetadata {
	panic("implement me")
}

func (asyncActWrapper) Run(context context.Context, inputs map[string]interface{}, handler action.ResultHandler) error {
	panic("implement me")
}

type syncActWrapper struct {
	legacyAct oldaction.SyncAction
}

func (syncActWrapper) Metadata() *action.Metadata {
	panic("implement me")
}

func (syncActWrapper) IOMetadata() *metadata.IOMetadata {
	panic("implement me")
}

func (syncActWrapper) Run(context context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	panic("implement me")
}
