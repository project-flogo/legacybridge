package legacybridge

import (
	"context"
	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
	olddata "github.com/TIBCOSoftware/flogo-lib/core/data"
	oldtrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/project-flogo/core/data/schema"

	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/trigger"
)

func RegisterLegacyTriggerFactory(ref string, factory oldtrigger.Factory) {
	w := wrapTriggerFactory(factory)
	trigger.LegacyRegister(ref, w)
}

func GetTrigger(trg oldtrigger.Trigger) trigger.Trigger {
	ref := support.GetRef(trg)
	return &triggerWrapper{legacyTrigger: trg, ref: ref}
}

func wrapTriggerFactory(legacyFactory oldtrigger.Factory) trigger.Factory {

	oldTrigger := legacyFactory.New(nil)
	newMd := toNewMetadata(oldTrigger.Metadata())

	w := &triggerFactoryWrapper{legacyFactory: legacyFactory, newMd: newMd, legacyTrigger: oldTrigger}
	return w
}

type triggerFactoryWrapper struct {
	legacyFactory oldtrigger.Factory
	newMd         *trigger.Metadata
	legacyTrigger oldtrigger.Trigger
}

func (w *triggerFactoryWrapper) Metadata() *trigger.Metadata {
	return w.newMd
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

	for i, hConfig := range config.Handlers {
		oldHandleConfig := &oldtrigger.HandlerConfig{}

		oldHandleConfig.Name = hConfig.Name
		oldHandleConfig.Settings = hConfig.Settings
		oldHandleConfig.Output = make(map[string]interface{})

		for name, sche := range hConfig.OutputSchemas {
			attr, ok := w.legacyTrigger.Metadata().Output[name]
			if ok && attr.Type() == legacyData.TypeComplexObject {
				s, err := schema.FindOrCreate(sche)
				if err != nil {
					return nil, err
				}
				oldHandleConfig.Output[name] = &legacyData.ComplexObject{Metadata: s.Value(), Value: nil}
			}
		}

		oldConfig.Handlers[i] = oldHandleConfig
	}

	//translate config
	legacyTrg := w.legacyFactory.New(oldConfig)
	trg = &triggerWrapper{legacyTrigger: legacyTrg, ref: oldConfig.Ref, legacyHandlers: oldConfig.Handlers}

	return trg, nil
}

type triggerWrapper struct {
	ref            string
	legacyTrigger  oldtrigger.Trigger
	legacyHandlers []*oldtrigger.HandlerConfig
}

func (w *triggerWrapper) Ref() string {
	return w.ref
}

func (w *triggerWrapper) Start() error {
	return w.legacyTrigger.Start()
}

func (w *triggerWrapper) Stop() error {
	return w.legacyTrigger.Stop()
}

func toNewMetadata(oldMd *oldtrigger.Metadata) *trigger.Metadata {

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
	if iTrg, ok := w.legacyTrigger.(oldtrigger.Initializable); ok {
		wCtx := &triggerInitCtxWrapper{ctx: ctx, legacyHandlers: w.legacyHandlers}
		return iTrg.Initialize(wCtx)
	}

	return nil
}

type triggerInitCtxWrapper struct {
	ctx            trigger.InitContext
	legacyHandlers []*oldtrigger.HandlerConfig
}

func (w *triggerInitCtxWrapper) GetHandlers() []*oldtrigger.Handler {
	handlers := w.ctx.GetHandlers()

	legacyHandlers := make([]*oldtrigger.Handler, len(handlers))
	for i, handler := range handlers {
		wrapHandler := &wrapperHandlerInf{handler: handler, legacyHandler: w.legacyHandlers[i]}
		legacyHandler := oldtrigger.NewHandlerAlt(wrapHandler)
		legacyHandlers[i] = legacyHandler
	}

	return legacyHandlers
}

type wrapperHandlerInf struct {
	handler       trigger.Handler
	legacyHandler *oldtrigger.HandlerConfig
}

func (w *wrapperHandlerInf) Handle(ctx context.Context, triggerData map[string]interface{}) (map[string]*olddata.Attribute, error) {

	for name, data := range triggerData {
		value, _, ok := GetComplexObjectInfo(data)
		if ok {
			triggerData[name] = value
		}
	}

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
	val, exists := w.handler.Settings()[setting]
	return val, exists
}

func (w *wrapperHandlerInf) GetOutput() map[string]interface{} {
	return w.legacyHandler.Output
}

func (w *wrapperHandlerInf) GetStringSetting(setting string) string {
	val, exists := w.handler.Settings()[setting]
	if !exists {
		return ""
	}

	retVal, _ := coerce.ToString(val)

	return retVal
}

func (w *wrapperHandlerInf) String() string {
	return ""
}
