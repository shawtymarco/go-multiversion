package v1_18_10

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestMenuFormConversion(t *testing.T) {
	input := []byte(`{"type":"form","title":"team","content":"body","elements":[{"type":"header","text":"section"},{"type":"button","text":"one","image":{"type":"path","data":"textures/ui/one"}},{"type":"divider","text":""},{"type":"button","text":"two"}]}`)
	request := &packet.ModalFormRequest{FormID: 7, FormData: append([]byte(nil), input...)}
	p := Protocol{runtime: &runtimeData{}}
	converted := p.convertGameplayFromLatest(request, nil)
	if len(converted) != 1 {
		t.Fatalf("conversion count: got %d, want 1", len(converted))
	}
	got := converted[0].(*packet.ModalFormRequest)
	if got == request || got.FormID != request.FormID {
		t.Fatalf("request was not cloned correctly: %#v", got)
	}
	if !bytes.Equal(request.FormData, input) {
		t.Fatalf("conversion mutated input: %s", request.FormData)
	}
	var form map[string]any
	if err := json.Unmarshal(got.FormData, &form); err != nil {
		t.Fatal(err)
	}
	if _, exists := form["elements"]; exists {
		t.Fatalf("target form still contains elements: %s", got.FormData)
	}
	if form["content"] != "body\nsection" {
		t.Fatalf("target form content: got %#v", form["content"])
	}
	buttons, ok := form["buttons"].([]any)
	if !ok || len(buttons) != 2 {
		t.Fatalf("target form buttons: %#v", form["buttons"])
	}
	for index, raw := range buttons {
		button := raw.(map[string]any)
		if _, exists := button["type"]; exists {
			t.Fatalf("target button %d contains current-only type: %#v", index, button)
		}
	}
	if buttons[0].(map[string]any)["text"] != "one" || buttons[1].(map[string]any)["text"] != "two" {
		t.Fatalf("target button order changed: %#v", buttons)
	}
}

func TestCustomFormConversion(t *testing.T) {
	input := []byte(`{"type":"custom_form","title":"settings","content":[{"type":"header","text":"section"},{"type":"divider","text":""},{"type":"label","text":"help"},{"type":"input","text":"name","default":"","placeholder":"name","tooltip":"tip"}]}`)
	got := targetFormData(input)
	var form map[string]any
	if err := json.Unmarshal(got, &form); err != nil {
		t.Fatal(err)
	}
	content := form["content"].([]any)
	if len(content) != 4 {
		t.Fatalf("target custom form element count: got %d, want 4", len(content))
	}
	for index := 0; index < 3; index++ {
		if elementType := content[index].(map[string]any)["type"]; elementType != "label" {
			t.Fatalf("target custom element %d type: got %#v, want label", index, elementType)
		}
	}
	if input := content[3].(map[string]any); input["type"] != "input" {
		t.Fatalf("target input changed type: %#v", input)
	} else if _, exists := input["tooltip"]; exists {
		t.Fatalf("target input contains current-only tooltip: %#v", input)
	}
}

func TestUnchangedFormDataIsCloned(t *testing.T) {
	for _, input := range [][]byte{
		[]byte(`{"type":"modal","title":"confirm","content":"body","button1":"yes","button2":"no"}`),
		[]byte(`not-json`),
	} {
		got := targetFormData(input)
		if !bytes.Equal(got, input) {
			t.Fatalf("unchanged form differs: got %s, want %s", got, input)
		}
		if len(input) != 0 && &got[0] == &input[0] {
			t.Fatal("unchanged form data was not cloned")
		}
	}
}
