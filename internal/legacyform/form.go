// Package legacyform downgrades current Dragonfly form documents to the JSON
// schema used before menu elements and custom-form tooltips were introduced.
package legacyform

import (
	"bytes"
	"encoding/json"
)

// Downgrade returns an independent form document accepted by legacy clients.
// Malformed, unknown, modal, and already-compatible documents are cloned and
// returned unchanged.
func Downgrade(data []byte) []byte {
	cloned := bytes.Clone(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var form map[string]any
	if err := decoder.Decode(&form); err != nil {
		return cloned
	}

	switch form["type"] {
	case "form":
		elements, ok := form["elements"].([]any)
		if !ok {
			return cloned
		}
		buttons := make([]any, 0, len(elements))
		body, _ := form["content"].(string)
		for _, raw := range elements {
			element, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			switch element["type"] {
			case "button":
				delete(element, "type")
				delete(element, "tooltip")
				buttons = append(buttons, element)
			case "header", "label":
				if text, ok := element["text"].(string); ok && text != "" {
					if body != "" {
						body += "\n"
					}
					body += text
				}
			}
		}
		form["content"] = body
		form["buttons"] = buttons
		delete(form, "elements")
	case "custom_form":
		content, ok := form["content"].([]any)
		if !ok {
			return cloned
		}
		for index, raw := range content {
			element, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			delete(element, "tooltip")
			switch element["type"] {
			case "divider":
				content[index] = map[string]any{"type": "label", "text": ""}
			case "header":
				element["type"] = "label"
			}
		}
		form["content"] = content
	default:
		return cloned
	}

	encoded, err := json.Marshal(form)
	if err != nil {
		return cloned
	}
	return encoded
}
