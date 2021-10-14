package legacybridge

import (
	"context"

	"github.com/project-flogo/core/data/schema"

	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
	legacyTrigger "github.com/TIBCOSoftware/flogo-lib/core/trigger"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/coerce"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/core/trigger"
)

func RegisterLegacyTriggerFactory(ref string, factory legacyTrigger.Factory) {
	err := trigger.LegacyRegister(ref, wrapTriggerFactory(factory))
	if err != nil {
		log.RootLogger().Warnf("Error registering legacy trigger '%s': %v", ref, err)
	}
}

func GetTrigger(trg legacyTrigger.Trigger) trigger.Trigger {
	ref := support.GetRef(trg)
	return &triggerWrapper{legacyTrigger: trg, ref: ref}
}

func wrapTriggerFactory(lFactory legacyTrigger.Factory) trigger.Factory {

	lTrigger := lFactory.New(nil)
	newMd := toNewMetadata(lTrigger.Metadata())

	w := &triggerFactoryWrapper{lFactory: lFactory, newMd: newMd, lTrigger: lTrigger}
	return w
}

type triggerFactoryWrapper struct {
	lFactory legacyTrigger.Factory
	lTrigger legacyTrigger.Trigger
	newMd    *trigger.Metadata
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

	lConfig := &legacyTrigger.Config{}
	lConfig.Name = config.Id
	lConfig.Id = config.Id
	lConfig.Ref = config.Ref
	lConfig.Settings = config.Settings

	lConfig.Handlers = make([]*legacyTrigger.HandlerConfig, len(config.Handlers))

	for i, hConfig := range config.Handlers {
		lHandleConfig := &legacyTrigger.HandlerConfig{}

		lHandleConfig.Name = hConfig.Name
		lHandleConfig.Settings = hConfig.Settings
		lHandleConfig.Output = make(map[string]interface{})

		if hConfig.Schemas != nil {
			for name, sche := range hConfig.Schemas.Output {
				attr, ok := w.lTrigger.Metadata().Output[name]
				if ok && attr.Type() == legacyData.TypeComplexObject {
					s, err := schema.FindOrCreate(sche)
					if err != nil {
						return nil, err
					}
					lHandleConfig.Output[name] = &legacyData.ComplexObject{Metadata: s.Value(), Value: nil}
				}
			}

		}

		lConfig.Handlers[i] = lHandleConfig
	}

	//translate config
	legacyTrg := w.lFactory.New(lConfig)
	trg = &triggerWrapper{legacyTrigger: legacyTrg, ref: lConfig.Ref, legacyHandlers: lConfig.Handlers}

	return trg, nil
}

type triggerWrapper struct {
	ref            string
	legacyTrigger  legacyTrigger.Trigger
	legacyHandlers []*legacyTrigger.HandlerConfig
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

func (w *triggerWrapper) Resume() error {
	fc, ok := w.legacyTrigger.(trigger.EventFlowControlAware)
	if ok {
		return fc.Resume()
	}
	return nil
}

func (w *triggerWrapper) Pause() error {
	fc, ok := w.legacyTrigger.(trigger.EventFlowControlAware)
	if ok {
		return fc.Pause()
	}
	return nil
}

func toNewMetadata(lMd *legacyTrigger.Metadata) *trigger.Metadata {

	newMd := &trigger.Metadata{}

	newMd.Settings = make(map[string]data.TypedValue, len(lMd.Settings))
	newMd.Reply = make(map[string]data.TypedValue, len(lMd.Reply))
	newMd.Output = make(map[string]data.TypedValue, len(lMd.Output))

	for name, attr := range lMd.Settings {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Settings[name] = data.NewTypedValue(newType, attr.Value())
	}

	if lMd.Handler != nil {
		newMd.HandlerSettings = make(map[string]data.TypedValue, len(lMd.Handler.Settings))
		for _, attr := range lMd.Handler.Settings {
			newType, _ := ToNewTypeFromLegacy(attr.Type())
			newMd.HandlerSettings[attr.Name()] = data.NewTypedValue(newType, attr.Value())
		}
	}

	for name, attr := range lMd.Reply {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Reply[name] = data.NewTypedValue(newType, attr.Value())
	}

	for name, attr := range lMd.Output {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Output[name] = data.NewTypedValue(newType, attr.Value())
	}

	return newMd
}

func (w *triggerWrapper) Initialize(ctx trigger.InitContext) error {
	//wrap init ctx
	if iTrg, ok := w.legacyTrigger.(legacyTrigger.Initializable); ok {
		wCtx := &triggerInitCtxWrapper{ctx: ctx, lHandlers: w.legacyHandlers}
		return iTrg.Initialize(wCtx)
	}

	return nil
}

type triggerInitCtxWrapper struct {
	ctx       trigger.InitContext
	lHandlers []*legacyTrigger.HandlerConfig
}

func (w *triggerInitCtxWrapper) GetHandlers() []*legacyTrigger.Handler {
	handlers := w.ctx.GetHandlers()

	lHandlers := make([]*legacyTrigger.Handler, len(handlers))
	for i, handler := range handlers {
		wrapHandler := &wrapperHandlerInf{handler: handler, lHandler: w.lHandlers[i]}
		lHandler := legacyTrigger.NewHandlerAlt(wrapHandler)
		lHandlers[i] = lHandler
	}

	return lHandlers
}

type wrapperHandlerInf struct {
	handler  trigger.Handler
	lHandler *legacyTrigger.HandlerConfig
}

func (w *wrapperHandlerInf) Handle(ctx context.Context, triggerData map[string]interface{}) (map[string]*legacyData.Attribute, error) {

	for name, td := range triggerData {
		value, _, ok := GetComplexObjectInfo(td)
		if ok {
			triggerData[name] = value
		}
	}

	ret, err := w.handler.Handle(ctx, triggerData)
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]*legacyData.Attribute)
	for name, value := range ret {
		attr, _ := legacyData.NewAttribute(name, legacyData.TypeAny, value)
		attrs[name] = attr
	}

	return attrs, nil
}

func (w *wrapperHandlerInf) GetSetting(setting string) (interface{}, bool) {
	val, exists := w.handler.Settings()[setting]
	return val, exists
}

func (w *wrapperHandlerInf) GetOutput() map[string]interface{} {
	return w.lHandler.Output
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
