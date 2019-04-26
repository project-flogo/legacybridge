package legacybridge

import (
	legacyActivity "github.com/TIBCOSoftware/flogo-lib/core/activity"
	legacyData "github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/data/schema"
	"github.com/project-flogo/core/support"
	"github.com/project-flogo/core/support/log"
	"github.com/project-flogo/flow/definition"
)

func RegisterLegacyActivity(act legacyActivity.Activity) {
	err := activity.LegacyRegister(act.Metadata().ID, wrapActivity(act))
	if err != nil {
		log.RootLogger().Warnf("Error registering legacy activity '%s': %v", act.Metadata().ID, err)
	}
}

func GetActivity(act legacyActivity.Activity) activity.Activity {
	return wrapActivity(act)
}

func wrapActivity(legacyAct legacyActivity.Activity) activity.Activity {
	ref := support.GetRef(legacyAct)

	newMd := convertOldMetadata(legacyAct.Metadata())

	aw := &activityWrapper{legacyAct: legacyAct, ref: ref, newMd: newMd}
	return aw
}

type activityWrapper struct {
	legacyAct legacyActivity.Activity
	newMd     *activity.Metadata
	ref       string
}

func (aw *activityWrapper) BypassValidation() bool {
	return schemaValidationEnabled
}

func (aw *activityWrapper) Ref() string {
	return aw.ref
}

func (aw *activityWrapper) Metadata() *activity.Metadata {
	return aw.newMd
}

func convertOldMetadata(oldMd *legacyActivity.Metadata) *activity.Metadata {

	newMd := &activity.Metadata{IOMetadata: &metadata.IOMetadata{}}

	newMd.Settings = make(map[string]data.TypedValue, len(oldMd.Settings))
	newMd.Input = make(map[string]data.TypedValue, len(oldMd.Input))
	newMd.Output = make(map[string]data.TypedValue, len(oldMd.Output))

	for name, attr := range oldMd.Settings {
		newMd.Settings[name] = convertOldAttribute(attr)
	}

	for name, attr := range oldMd.Input {
		newMd.Input[name] = convertOldAttribute(attr)
	}

	for name, attr := range oldMd.Output {
		newMd.Output[name] = convertOldAttribute(attr)
	}

	return newMd
}

func convertOldAttribute(attr *legacyData.Attribute) *data.Attribute {
	newType, _ := ToNewTypeFromLegacy(attr.Type())
	newVal := attr.Value()
	var newSchema schema.Schema

	//special handling for ComplexObjects
	if attr.Type() == legacyData.TypeComplexObject && attr.Value() != nil {

		if cVal, ok := attr.Value().(*legacyData.ComplexObject); ok {

			newVal = cVal.Value

			if cVal.Metadata != "" {
				//has schema
				def := &schema.Def{Type: "json", Value: cVal.Metadata}
				s, err := schema.New(def)
				if err != nil {
					//log error
				}
				newSchema = s
			}
		}
	}

	return data.NewAttributeWithSchema(attr.Name(), newType, newVal, newSchema)
}

func (aw *activityWrapper) Eval(ctx activity.Context) (done bool, err error) {
	legacyCtx := wrapActContext(ctx, aw.legacyAct)
	return aw.legacyAct.Eval(legacyCtx)
}

func wrapActContext(ctx activity.Context, legacyAct legacyActivity.Activity) legacyActivity.Context {

	wrappedCtx := &activityCtxWrapper{ctx: ctx, legacyAct: legacyAct}

	if lCtx, ok := ctx.(activity.LegacyCtx); ok {
		wrappedCtx.lCtx = lCtx
	}

	return wrappedCtx
}

type activityCtxWrapper struct {
	legacyAct legacyActivity.Activity
	ctx       activity.Context
	lCtx      activity.LegacyCtx
}

func (w *activityCtxWrapper) ActivityHost() legacyActivity.Host {
	return &activityHostWrapper{host: w.ctx.ActivityHost()}
}

func (w *activityCtxWrapper) Name() string {
	return w.ctx.Name()
}

func (w *activityCtxWrapper) GetInput(name string) interface{} {

	val := w.ctx.GetInput(name)

	// if the input value is complex, we need to modify it
	if oldMdInput := w.legacyAct.Metadata().Input; oldMdInput != nil {
		if attr, ok := oldMdInput[name]; ok {
			if attr.Type() == legacyData.TypeComplexObject {
				md := ""

				//see if there is a corresponding input schema to construct complex object
				if sIO, ok := w.ctx.(schema.HasSchemaIO); ok {
					s := sIO.GetInputSchema(name)
					if s != nil {
						md = s.Value()
					}
				}
				if val == "" {
					//Set to empty object
					val = "{}"
				}
				return &legacyData.ComplexObject{Metadata: md, Value: val}
			}
		}
	}

	return val
}

func (w *activityCtxWrapper) GetOutput(name string) interface{} {

	// if the input value is complex, we need to modify it
	if oldMdOutput := w.legacyAct.Metadata().Output; oldMdOutput != nil {
		if attr, ok := oldMdOutput[name]; ok {
			if attr.Type() == legacyData.TypeComplexObject {
				md := ""

				//see if there is a corresponding input schema to construct complex object
				if sIO, ok := w.ctx.(schema.HasSchemaIO); ok {
					s := sIO.GetOutputSchema(name)
					if s != nil {
						md = s.Value()
					}
				}

				return &legacyData.ComplexObject{Metadata: md, Value: nil}
			}
		}
	}

	if w.lCtx != nil {
		return w.lCtx.GetOutput(name)
	}

	return nil
}

func (w *activityCtxWrapper) SetOutput(name string, value interface{}) {

	if oldMdOutput := w.legacyAct.Metadata().Output; oldMdOutput != nil {

		if attr, ok := oldMdOutput[name]; ok {
			if attr.Type() == legacyData.TypeComplexObject {

				if cVal, ok := value.(*legacyData.ComplexObject); ok {
					err := w.ctx.SetOutput(name, cVal.Value)
					if err != nil {
						log.RootLogger().Errorf("error setting output '%s': %v", attr.Name(), err)
					}
					return
				}
			}
		}
	}

	err := w.ctx.SetOutput(name, value)
	if err != nil {
		log.RootLogger().Errorf("error setting output '%s': %v", name, err)
	}
}

func (w *activityCtxWrapper) GetSetting(setting string) (value interface{}, exists bool) {
	return nil, false
}

func (*activityCtxWrapper) GetInitValue(key string) (value interface{}, exists bool) {
	return nil, false
}

func (w *activityCtxWrapper) TaskName() string {
	return w.ctx.Name()
}

func (w *activityCtxWrapper) FlowDetails() legacyActivity.FlowDetails {
	return &flowDetails{host: w.ctx.ActivityHost()}
}

type activityHostWrapper struct {
	host activity.Host
}

func (w *activityHostWrapper) ID() string {
	return w.host.ID()
}

func (w *activityHostWrapper) Name() string {
	return w.host.Name()
}

func (w *activityHostWrapper) IOMetadata() *legacyData.IOMetadata {

	md := w.host.IOMetadata()

	oldMd := &legacyData.IOMetadata{}
	oldMd.Input = make(map[string]*legacyData.Attribute, len(md.Input))
	oldMd.Output = make(map[string]*legacyData.Attribute, len(md.Output))

	for name, tv := range md.Input {
		oldMd.Input[name], _ = toLegacyAttribute(name, tv)
	}

	for name, tv := range md.Output {
		oldMd.Output[name], _ = toLegacyAttribute(name, tv)
	}

	return oldMd
}

func toLegacyAttribute(name string, tv data.TypedValue) (*legacyData.Attribute, error) {
	legacyType, _ := ToLegacyFromNewType(tv.Type())
	legacyValue := tv.Value()

	if tv.Type() == data.TypeObject {
		if s, ok := tv.(schema.HasSchema); ok {
			legacyType = legacyData.TypeComplexObject
			legacyValue = &legacyData.ComplexObject{Metadata: s.Schema().Value(), Value: tv.Value()}
		}
	}

	return legacyData.NewAttribute(name, legacyType, legacyValue)
}

func (w *activityHostWrapper) Reply(replyData map[string]*legacyData.Attribute, err error) {

	reply := make(map[string]interface{}, len(replyData))
	for name, attr := range replyData {
		if attr.Type() == legacyData.TypeComplexObject {
			if attr.Value() != nil {
				if compx, err := legacyData.CoerceToComplexObject(attr.Value()); err != nil && compx != nil {
					reply[name] = compx.Value
				} else {
					if err != nil {
						log.RootLogger().Errorf("Unable to coerce legacy complex attr '%s': %v", name, err)
					}
				}
			} else {
				reply[name] = nil
			}
		} else {
			reply[attr.Name()] = attr.Value()
		}
	}
	w.host.Reply(reply, err)
}

func (w *activityHostWrapper) Return(returnData map[string]*legacyData.Attribute, err error) {
	ret := make(map[string]interface{}, len(returnData))

	for name, attr := range returnData {
		if attr.Type() == legacyData.TypeComplexObject {
			if attr.Value() != nil {
				if compx, err := legacyData.CoerceToComplexObject(attr.Value()); err != nil && compx != nil {
					ret[name] = compx.Value
				} else {
					if err != nil {
						log.RootLogger().Errorf("Unable to coerce legacy complex attr '%s': %v", name, err)
					}
				}
			} else {
				ret[name] = nil
			}
		} else {
			ret[attr.Name()] = attr.Value()
		}
	}

	w.host.Return(ret, err)
}

func (w *activityHostWrapper) WorkingData() legacyData.Scope {
	return &scopeWrapper{s: w.host.Scope()}
}

func (w *activityHostWrapper) GetResolver() legacyData.Resolver {
	return &resolverWrapper{resolver: definition.GetDataResolver()}
}

type resolverWrapper struct {
	resolver resolve.CompositeResolver
}

func (w *resolverWrapper) Resolve(toResolve string, scope legacyData.Scope) (value interface{}, err error) {
	return w.resolver.Resolve(toResolve, &legacyScopeWrapper{scope})
}

type legacyScopeWrapper struct {
	s legacyData.Scope
}

func (w *legacyScopeWrapper) GetValue(name string) (value interface{}, exists bool) {

	if attr, exists := w.s.GetAttr(name); exists {
		return attr.Value(), true
	}

	return nil, false
}

func (w *legacyScopeWrapper) SetValue(name string, value interface{}) error {
	err := w.s.SetAttrValue(name, value)
	return err
}

type scopeWrapper struct {
	s data.Scope
}

func (w *scopeWrapper) GetAttr(name string) (attr *legacyData.Attribute, exists bool) {
	if val, exists := w.s.GetValue(name); exists {
		attr, err := legacyData.NewAttribute(name, legacyData.TypeAny, val)
		if err != nil {
			return nil, false
		}
		return attr, true
	}

	return nil, false
}

func (w *scopeWrapper) SetAttrValue(name string, value interface{}) error {
	err := w.s.SetValue(name, value)
	return err
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// FlowDetails

type flowDetails struct {
	host activity.Host
}

// ID implements activity.FlowDetails.ID
func (fd *flowDetails) ID() string {
	return fd.host.ID()
}

// Name implements activity.FlowDetails.Name
func (fd *flowDetails) Name() string {
	return fd.host.Name()
}

func (fd *flowDetails) ReplyHandler() legacyActivity.ReplyHandler {
	return nil
}
