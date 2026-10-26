// Package axml writes Android binary XML directly, based on AOSP ResourceTypes.h.
package axml

import (
	"fmt"
	"koru.dev/koru/internal/androidbin"
	"sort"
)

const android = "http://schemas.android.com/apk/res/android"
const none uint32 = 0xffffffff

type Manifest struct {
	Package, Name, Version, Activity string
	VersionCode, MinSDK, TargetSDK   uint32
	Permissions                      []string
	Debuggable                       bool
	Icon                             uint32
	// Orientation is "", "portrait", "landscape" or "sensor"; "" emits no
	// android:screenOrientation.
	Orientation string
	// AllowBackup sets android:allowBackup on the application.
	AllowBackup bool
	// Keyboard is "" or "resize" (adjustResize, the default) or "pan" (adjustPan).
	Keyboard string
	// DeepLinks add one VIEW/DEFAULT/BROWSABLE intent filter per link. One
	// filter per link avoids Android's cross product of schemes, hosts and
	// paths declared in the same filter. The URL is not yet delivered to
	// JavaScript: the link opens the app on its initial page.
	DeepLinks []DeepLink
}

// DeepLink is one URL pattern handled by the Activity.
type DeepLink struct{ Scheme, Host, PathPrefix string }

// Values of android:screenOrientation, from attrs_manifest.xml
// (<attr name="screenOrientation">): landscape=0, portrait=1, sensor=4.
var orientations = map[string]uint32{"landscape": 0, "portrait": 1, "sensor": 4}

// Values of android:windowSoftInputMode, from attrs_manifest.xml:
// adjustResize=0x10, adjustPan=0x20.
var keyboards = map[string]uint32{"": 0x10, "resize": 0x10, "pan": 0x20}

type attribute struct {
	name, text string
	kind       byte
	value      uint32
	android    bool
}
type node struct {
	name     string
	attrs    []attribute
	children []node
}

// Attribute resource IDs come from AOSP frameworks/base/core/res/res/values/
// public-final.xml (https://android.googlesource.com/platform/frameworks/base/):
// screenOrientation 0x0101001e, scheme 0x01010027, host 0x01010028,
// pathPrefix 0x0101002b and autoVerify 0x010104ee (API 23), besides the
// attributes below that were already in use. Only attributes present in the
// document are written to the resource map.
var ids = map[string]uint32{"screenOrientation": 0x0101001e, "scheme": 0x01010027, "host": 0x01010028, "pathPrefix": 0x0101002b, "autoVerify": 0x010104ee, "theme": 0x01010000, "label": 0x01010001, "icon": 0x01010002, "name": 0x01010003, "debuggable": 0x0101000f, "exported": 0x01010010, "minSdkVersion": 0x0101020c, "versionCode": 0x0101021b, "versionName": 0x0101021c, "windowSoftInputMode": 0x0101022b, "targetSdkVersion": 0x01010270, "allowBackup": 0x01010280, "hardwareAccelerated": 0x010102d3, "usesCleartextTraffic": 0x010104ec, "enableOnBackInvokedCallback": 0x0101066c}

func str(n, v string) attribute { return attribute{name: n, text: v, kind: 3, android: true} }
func integer(n string, v uint32) attribute {
	return attribute{name: n, kind: 0x10, value: v, android: true}
}
func boolean(n string, v bool) attribute {
	a := attribute{name: n, kind: 0x12, android: true}
	if v {
		a.value = 0xffffffff
	}
	return a
}
func Build(m Manifest) ([]byte, error) {
	mode, ok := keyboards[m.Keyboard]
	if !ok {
		return nil, fmt.Errorf("unsupported keyboard mode %q", m.Keyboard)
	}
	activityAttrs := []attribute{str("name", m.Activity), boolean("exported", true), integer("windowSoftInputMode", mode), boolean("enableOnBackInvokedCallback", false)}
	if m.Orientation != "" {
		v, ok := orientations[m.Orientation]
		if !ok {
			return nil, fmt.Errorf("unsupported orientation %q", m.Orientation)
		}
		activityAttrs = append(activityAttrs, integer("screenOrientation", v))
	}
	root := node{name: "manifest", attrs: []attribute{{name: "package", text: m.Package, kind: 3}, integer("versionCode", m.VersionCode), str("versionName", m.Version)}}
	root.children = append(root.children, node{name: "uses-sdk", attrs: []attribute{integer("minSdkVersion", m.MinSDK), integer("targetSdkVersion", m.TargetSDK)}})
	permissions := append([]string(nil), m.Permissions...)
	sort.Strings(permissions)
	for _, p := range permissions {
		root.children = append(root.children, node{name: "uses-permission", attrs: []attribute{str("name", p)}})
	}
	application := node{name: "application", attrs: []attribute{str("label", m.Name), boolean("debuggable", m.Debuggable), boolean("allowBackup", m.AllowBackup), boolean("hardwareAccelerated", true), boolean("usesCleartextTraffic", false), {name: "theme", kind: 1, value: 0x01030241, android: true}}}
	if m.Icon != 0 {
		application.attrs = append(application.attrs, attribute{name: "icon", kind: 1, value: m.Icon, android: true})
	}
	activity := node{name: "activity", attrs: activityAttrs, children: []node{{name: "intent-filter", children: []node{{name: "action", attrs: []attribute{str("name", "android.intent.action.MAIN")}}, {name: "category", attrs: []attribute{str("name", "android.intent.category.LAUNCHER")}}}}}}
	for _, l := range m.DeepLinks {
		if l.Scheme == "" {
			return nil, fmt.Errorf("deep link without scheme")
		}
		data := []attribute{str("scheme", l.Scheme)}
		if l.Host != "" {
			data = append(data, str("host", l.Host))
		}
		if l.PathPrefix != "" {
			data = append(data, str("pathPrefix", l.PathPrefix))
		}
		filter := node{name: "intent-filter", children: []node{
			{name: "action", attrs: []attribute{str("name", "android.intent.action.VIEW")}},
			{name: "category", attrs: []attribute{str("name", "android.intent.category.DEFAULT")}},
			{name: "category", attrs: []attribute{str("name", "android.intent.category.BROWSABLE")}},
			{name: "data", attrs: data},
		}}
		if l.Scheme == "https" {
			filter.attrs = append(filter.attrs, boolean("autoVerify", true))
		}
		activity.children = append(activity.children, filter)
	}
	application.children = append(application.children, activity)
	root.children = append(root.children, application)
	used := map[string]bool{}
	var scan func(node)
	scan = func(n node) {
		for _, a := range n.attrs {
			if a.android {
				used[a.name] = true
			}
		}
		for _, c := range n.children {
			scan(c)
		}
	}
	scan(root)
	names := make([]string, 0, len(used))
	for n := range used {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return ids[names[i]] < ids[names[j]] })
	strings := []string{}
	index := map[string]uint32{}
	add := func(s string) uint32 {
		if i, ok := index[s]; ok {
			return i
		}
		i := uint32(len(strings))
		strings = append(strings, s)
		index[s] = i
		return i
	}
	for _, n := range names {
		add(n)
	}
	add("android")
	add(android)
	var collect func(node)
	collect = func(n node) {
		add(n.name)
		for _, a := range n.attrs {
			add(a.name)
			if a.kind == 3 {
				add(a.text)
			}
		}
		for _, c := range n.children {
			collect(c)
		}
	}
	collect(root)
	pool, e := androidbin.Pool(strings)
	if e != nil {
		return nil, e
	}
	payload := append([]byte(nil), pool...)
	resources := androidbin.Header(0x180, 8, uint32(8+4*len(names)))
	for _, n := range names {
		resources = androidbin.U32(resources, ids[n])
	}
	payload = append(payload, resources...)
	header := func(t uint16, size uint32) []byte {
		b := androidbin.Header(t, 16, size)
		b = androidbin.U32(b, 1)
		return androidbin.U32(b, none)
	}
	ns := func(t uint16) []byte {
		b := header(t, 24)
		b = androidbin.U32(b, index["android"])
		return androidbin.U32(b, index[android])
	}
	payload = append(payload, ns(0x100)...)
	var emit func(node)
	emit = func(n node) {
		attrs := append([]attribute(nil), n.attrs...)
		sort.SliceStable(attrs, func(i, j int) bool {
			if attrs[i].android != attrs[j].android {
				return !attrs[i].android
			}
			return ids[attrs[i].name] < ids[attrs[j].name]
		})
		b := header(0x102, uint32(36+20*len(attrs)))
		b = androidbin.U32(b, none)
		b = androidbin.U32(b, index[n.name])
		b = androidbin.U16(b, 20)
		b = androidbin.U16(b, 20)
		b = androidbin.U16(b, uint16(len(attrs)))
		b = androidbin.U16(b, 0)
		b = androidbin.U16(b, 0)
		b = androidbin.U16(b, 0)
		for _, a := range attrs {
			namespace := none
			if a.android {
				namespace = index[android]
			}
			b = androidbin.U32(b, namespace)
			b = androidbin.U32(b, index[a.name])
			raw, value := none, a.value
			if a.kind == 3 {
				raw = index[a.text]
				value = raw
			}
			b = androidbin.U32(b, raw)
			b = androidbin.U16(b, 8)
			b = append(b, 0, a.kind)
			b = androidbin.U32(b, value)
		}
		payload = append(payload, b...)
		for _, c := range n.children {
			emit(c)
		}
		b = header(0x103, 24)
		b = androidbin.U32(b, none)
		b = androidbin.U32(b, index[n.name])
		payload = append(payload, b...)
	}
	emit(root)
	payload = append(payload, ns(0x101)...)
	return append(androidbin.Header(3, 8, uint32(8+len(payload))), payload...), nil
}
