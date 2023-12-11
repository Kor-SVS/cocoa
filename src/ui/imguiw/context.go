package imguiw

import imgui "github.com/AllenDang/cimgui-go"

var Context *ImguiWContext

type ImguiWContext struct {
	imDPI     *ImguiDPI
	imBackend imgui.Backend[imgui.GLFWWindowFlags]
	FontAtlas *FontAtlas
	context   *imgui.Context
	idCounter int
}

func (ic *ImguiWContext) Backend() imgui.Backend[imgui.GLFWWindowFlags] {
	return ic.imBackend
}

func (ic *ImguiWContext) ID() int {
	ic.idCounter++
	return ic.idCounter
}

func (ic *ImguiWContext) IO() *imgui.IO {
	return imgui.CurrentIO()
}
