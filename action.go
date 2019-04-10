package legacybridge

import (
	"context"
	"github.com/project-flogo/core/support/log"

	legacyAction "github.com/TIBCOSoftware/flogo-lib/core/action"

	"github.com/project-flogo/core/action"
	"github.com/project-flogo/core/data/metadata"
)

func RegisterLegacyAction(ref string, legacyFactory legacyAction.Factory) {
	err := action.LegacyRegister(ref, &legacyFactoryWrapper{legacyFactory: legacyFactory})
	if err != nil {
		log.RootLogger().Warnf("Error registering legacy action '%s': %v", ref, err)
	}
}

type legacyFactoryWrapper struct {
	legacyFactory legacyAction.Factory
}

func (legacyFactoryWrapper) Initialize(ctx action.InitContext) error {
	//ignore
	return nil
}

func (w *legacyFactoryWrapper) New(config *action.Config) (action.Action, error) {

	lConfig := &legacyAction.Config{}
	lConfig.Ref = config.Ref
	lConfig.Settings = config.Settings
	lConfig.Id = config.Id

	legacyAct, err := w.legacyFactory.New(lConfig)
	if err != nil {
		return nil, err
	}

	var wa action.Action

	if act, ok := legacyAct.(legacyAction.AsyncAction); ok {
		wa = wrapAsyncAction(act)
	} else if act, ok := legacyAct.(legacyAction.SyncAction); ok {
		wa = wrapSyncAction(act)
	}

	return wa, nil
}

func wrapAsyncAction(legacyAct legacyAction.AsyncAction) action.AsyncAction {

	aw := &asyncActWrapper{legacyAct: legacyAct}
	//todo wrap metadata
	return aw
}

func wrapSyncAction(legacyAct legacyAction.SyncAction) action.SyncAction {

	aw := &syncActWrapper{legacyAct: legacyAct}
	//todo wrap metadata
	return aw
}

type asyncActWrapper struct {
	legacyAct legacyAction.AsyncAction
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
	legacyAct legacyAction.SyncAction
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
