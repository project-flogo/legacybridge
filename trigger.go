package legacybridge

import (
	"context"
	olddata "github.com/TIBCOSoftware/flogo-lib/core/data"
	oldtrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/trigger"
)

func RegisterLegacyTriggerFactory(ref string, factory oldtrigger.Factory) {
	w := wrapTriggerFactory(factory)
	trigger.LegacyRegister(ref, w)
}

func wrapTriggerFactory(legacyFactory oldtrigger.Factory) trigger.Factory {
	w := &triggerFactoryWrapper{legacyFactory: legacyFactory}
	return w
}

type triggerFactoryWrapper struct {
	legacyFactory oldtrigger.Factory
}

func (w *triggerFactoryWrapper) New(config *trigger.Config) (trg trigger.Trigger, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	oldConfig := &oldtrigger.Config{}
	oldConfig.Name = config.Id
	oldConfig.Id = config.Id
	oldConfig.Ref = config.Ref
	oldConfig.Settings = config.Settings

	oldConfig.Handlers = make([]*oldtrigger.HandlerConfig, len(config.Handlers))

	for _, hConfig := range config.Handlers {
		oldHandleConfig := &oldtrigger.HandlerConfig{}

		oldHandleConfig.Name = hConfig.Name
		oldHandleConfig.Settings = hConfig.Settings
	}

	//translate config
	legacyTrg := w.legacyFactory.New(oldConfig)
	trg = &triggerWrapper{legacyTrg: legacyTrg}

	return trg, nil
}

type triggerWrapper struct {
	legacyTrg oldtrigger.Trigger
}

func (w *triggerWrapper) Start() error {
	return w.legacyTrg.Start()
}

func (w *triggerWrapper) Stop() error {
	return w.legacyTrg.Stop()
}

func (w *triggerWrapper) Metadata() *trigger.Metadata {

	oldMd := w.legacyTrg.Metadata()
	newMd := &trigger.Metadata{}

	newMd.Settings = make(map[string]data.TypedValue, len(oldMd.Settings))
	newMd.HandlerSettings = make(map[string]data.TypedValue, len(oldMd.Handler.Settings))
	newMd.Reply = make(map[string]data.TypedValue, len(oldMd.Reply))
	newMd.Output = make(map[string]data.TypedValue, len(oldMd.Output))

	for name, attr := range oldMd.Settings {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Settings[name] = data.NewTypedValue(newType, attr.Value())
	}

	for _, attr := range oldMd.Handler.Settings {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.HandlerSettings[attr.Name()] = data.NewTypedValue(newType, attr.Value())
	}

	for name, attr := range oldMd.Reply {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Reply[name] = data.NewTypedValue(newType, attr.Value())
	}

	for name, attr := range oldMd.Output {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Output[name] = data.NewTypedValue(newType, attr.Value())
	}

	return newMd
}

func (w *triggerWrapper) Initialize(ctx trigger.InitContext) error {
	//wrap init ctx
	if iTrg, ok := w.legacyTrg.(oldtrigger.Initializable); ok {
		wCtx := &triggerInitCtxWrapper{ctx}
		return iTrg.Initialize(wCtx)
	}

	return nil
}

type triggerInitCtxWrapper struct {
	ctx trigger.InitContext
}

func (w *triggerInitCtxWrapper) GetHandlers() []*oldtrigger.Handler {
	handlers := w.ctx.GetHandlers()

	legacyHandlers := make([]*oldtrigger.Handler, len(handlers))
	for i, handler := range handlers {

		w := &wrapperHandlerInf{handler: handler}
		legacyHandler := oldtrigger.NewHandlerAlt(w)
		legacyHandlers[i] = legacyHandler
	}

	return legacyHandlers
}

type wrapperHandlerInf struct {
	handler trigger.Handler
}

func (w *wrapperHandlerInf) Handle(ctx context.Context, triggerData map[string]interface{}) (map[string]*olddata.Attribute, error) {
	ret, err := w.handler.Handle(ctx, triggerData)
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]*olddata.Attribute)
	for name, value := range ret {
		attr, _ := olddata.NewAttribute(name, olddata.TypeAny, value)
		attrs[name] = attr
	}

	return attrs, nil
}

func (w *wrapperHandlerInf) GetSetting(setting string) (interface{}, bool) {
	return w.handler.GetSetting(setting)
}

func (w *wrapperHandlerInf) GetOutput() map[string]interface{} {
	return nil
}

func (w *wrapperHandlerInf) GetStringSetting(setting string) string {
	val, exists := w.handler.GetSetting(setting)
	if !exists {
		return ""
	}

	retVal, _ := coerce.ToString(val)

	return retVal
}

func (w *wrapperHandlerInf) String() string {
	return ""
}
