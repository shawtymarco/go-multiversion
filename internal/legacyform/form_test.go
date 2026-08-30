package legacyform

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMenu(t *testing.T) {
	input := []byte(`{"type":"form","title":"team","content":"body","elements":[{"type":"header","text":"section"},{"type":"button","text":"one","image":{"type":"path","data":"textures/ui/one"}},{"type":"divider","text":""},{"type":"label","text":"help"},{"type":"button","text":"two","tooltip":"tip"}]}`)
	got := Downgrade(input)
	if bytes.Equal(got, input) || !bytes.Equal(input, []byte(`{"type":"form","title":"team","content":"body","elements":[{"type":"header","text":"section"},{"type":"button","text":"one","image":{"type":"path","data":"textures/ui/one"}},{"type":"divider","text":""},{"type":"label","text":"help"},{"type":"button","text":"two","tooltip":"tip"}]}`)) {
		t.Fatal("menu conversion mutated or failed to convert the input")
	}
	var form map[string]any
	if err := json.Unmarshal(got, &form); err != nil {
		t.Fatal(err)
	}
	if _, exists := form["elements"]; exists {
		t.Fatalf("legacy menu still contains elements: %s", got)
	}
	if form["content"] != "body\nsection\nhelp" {
		t.Fatalf("legacy menu content: %#v", form["content"])
	}
	buttons := form["buttons"].([]any)
	if len(buttons) != 2 {
		t.Fatalf("button count: got %d, want 2", len(buttons))
	}
	for index, raw := range buttons {
		button := raw.(map[string]any)
		if _, exists := button["type"]; exists {
			t.Fatalf("button %d retains type: %#v", index, button)
		}
		if _, exists := button["tooltip"]; exists {
			t.Fatalf("button %d retains tooltip: %#v", index, button)
		}
	}
	if buttons[0].(map[string]any)["text"] != "one" || buttons[1].(map[string]any)["text"] != "two" {
		t.Fatalf("button indexes changed: %#v", buttons)
	}
	if _, ok := buttons[0].(map[string]any)["image"].(map[string]any); !ok {
		t.Fatalf("button image was not preserved: %#v", buttons[0])
	}
}

func TestCustomForm(t *testing.T) {
	input := []byte(`{"type":"custom_form","title":"settings","content":[{"type":"header","text":"section"},{"type":"divider","text":""},{"type":"label","text":"help"},{"type":"input","text":"name","default":"","placeholder":"name","tooltip":"tip"}]}`)
	got := Downgrade(input)
	var form map[string]any
	if err := json.Unmarshal(got, &form); err != nil {
		t.Fatal(err)
	}
	content := form["content"].([]any)
	if len(content) != 4 {
		t.Fatalf("element count: got %d, want 4", len(content))
	}
	for index := 0; index < 3; index++ {
		if elementType := content[index].(map[string]any)["type"]; elementType != "label" {
			t.Fatalf("element %d type: got %#v, want label", index, elementType)
		}
	}
	if input := content[3].(map[string]any); input["type"] != "input" {
		t.Fatalf("input changed: %#v", input)
	} else if _, exists := input["tooltip"]; exists {
		t.Fatalf("input retains tooltip: %#v", input)
	}
}

func TestUnchangedDataIsCloned(t *testing.T) {
	for _, input := range [][]byte{
		[]byte(`{"type":"modal","title":"confirm","content":"body","button1":"yes","button2":"no"}`),
		[]byte(`{"type":"form","title":"legacy","content":"body","buttons":[]}`),
		[]byte(`not-json`),
	} {
		got := Downgrade(input)
		if !bytes.Equal(got, input) {
			t.Fatalf("unchanged data differs: got %s, want %s", got, input)
		}
		if len(input) != 0 && &got[0] == &input[0] {
			t.Fatal("unchanged form data was not cloned")
		}
	}
}
