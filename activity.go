package legacybridge

import (
	oldactivity "github.com/TIBCOSoftware/flogo-lib/core/activity"
	olddata "github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data"
	"github.com/project-flogo/core/data/metadata"
	"github.com/project-flogo/core/data/resolve"
	"github.com/project-flogo/core/support"
)

func RegisterLegacyActivity(act oldactivity.Activity) {
	wa := wrapActivity(act)
	activity.LegacyRegister(act.Metadata().ID, wa)
}

func GetActivity(act oldactivity.Activity) activity.Activity {
	return wrapActivity(act)
}

func wrapActivity(legacyAct oldactivity.Activity) activity.Activity {
	ref := support.GetRef(legacyAct)
	aw := &activityWrapper{legacyAct: legacyAct, ref: ref}
	return aw
}

type activityWrapper struct {
	legacyAct oldactivity.Activity
	ref       string
}

func (aw *activityWrapper) Ref() string {
	return aw.ref
}

func (aw *activityWrapper) Metadata() *activity.Metadata {

	oldMd := aw.legacyAct.Metadata()

	if oldMd == nil {
		return nil
	}

	newMd := &activity.Metadata{IOMetadata: &metadata.IOMetadata{}}

	newMd.Settings = make(map[string]data.TypedValue, len(oldMd.Settings))
	newMd.Input = make(map[string]data.TypedValue, len(oldMd.Input))
	newMd.Output = make(map[string]data.TypedValue, len(oldMd.Output))

	for name, attr := range oldMd.Settings {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Settings[name] = data.NewTypedValue(newType, attr.Value())
	}

	for name, attr := range oldMd.Input {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Input[name] = data.NewTypedValue(newType, attr.Value())
	}

	for name, attr := range oldMd.Output {
		newType, _ := ToNewTypeFromLegacy(attr.Type())
		newMd.Output[name] = data.NewTypedValue(newType, attr.Value())
	}

	return newMd
}

func (aw *activityWrapper) Eval(ctx activity.Context) (done bool, err error) {
	legacyCtx := wrapActContext(ctx)
	return aw.legacyAct.Eval(legacyCtx)
}

func wrapActContext(ctx activity.Context) oldactivity.Context {

	wac := &activityCtxWrapper{ctx: ctx}

	return wac
}

type activityCtxWrapper struct {
	ctx activity.Context
}

func (w *activityCtxWrapper) ActivityHost() oldactivity.Host {
	return &activityHostWrapper{host: w.ctx.ActivityHost()}
}

func (w *activityCtxWrapper) Name() string {
	return w.ctx.Name()
}

func (w *activityCtxWrapper) GetInput(name string) interface{} {
	return w.ctx.GetInput(name)
}

func (w *activityCtxWrapper) GetOutput(name string) interface{} {
	return nil //used for schema
}

func (w *activityCtxWrapper) SetOutput(name string, value interface{}) {
	w.ctx.SetOutput(name, value)
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

func (w *activityCtxWrapper) FlowDetails() oldactivity.FlowDetails {
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

func (w *activityHostWrapper) IOMetadata() *olddata.IOMetadata {

	md := w.host.IOMetadata()

	oldMd := &olddata.IOMetadata{}
	oldMd.Input = make(map[string]*olddata.Attribute, len(md.Input))
	oldMd.Output = make(map[string]*olddata.Attribute, len(md.Output))

	for name, tv := range md.Input {
		legacyType, _ := ToLegacyFromNewType(tv.Type())
		oldMd.Input[name], _ = olddata.NewAttribute(name, legacyType, tv.Value())
	}

	for name, tv := range md.Output {
		legacyType, _ := ToLegacyFromNewType(tv.Type())
		oldMd.Output[name], _ = olddata.NewAttribute(name, legacyType, tv.Value())
	}

	return oldMd
}

func (w *activityHostWrapper) Reply(replyData map[string]*olddata.Attribute, err error) {

	reply := make(map[string]interface{}, len(replyData))
	for _, attr := range replyData {
		reply[attr.Name()] = attr.Value()
	}

	w.host.Reply(reply, err)
}

func (w *activityHostWrapper) Return(returnData map[string]*olddata.Attribute, err error) {
	ret := make(map[string]interface{}, len(returnData))
	for _, attr := range returnData {
		ret[attr.Name()] = attr.Value()
	}

	w.host.Reply(ret, err)
}

func (w *activityHostWrapper) WorkingData() olddata.Scope {
	return &scopeWrapper{s: w.host.WorkingData()}
}

func (w *activityHostWrapper) GetResolver() olddata.Resolver {
	return &resolverWrapper{resolver: w.host.GetResolver()}
}

type resolverWrapper struct {
	resolver resolve.CompositeResolver
}

func (w *resolverWrapper) Resolve(toResolve string, scope olddata.Scope) (value interface{}, err error) {
	return w.resolver.Resolve(toResolve, &legacyScopeWrapper{})
}

type legacyScopeWrapper struct {
	s olddata.Scope
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

func (w *scopeWrapper) GetAttr(name string) (attr *olddata.Attribute, exists bool) {
	if val, exists := w.s.GetValue(name); exists {
		attr, err := olddata.NewAttribute(name, olddata.TypeAny, val)
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

func (fd *flowDetails) ReplyHandler() oldactivity.ReplyHandler {
	return nil
}
