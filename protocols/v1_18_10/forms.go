package v1_18_10

import (
	"bytes"
	"encoding/json"
)

// targetFormData converts the current in-packet form JSON schema to the schema
// accepted by Minecraft 1.18. The ModalFormRequest wire layout itself did not
// change, so the compatibility boundary is the embedded JSON document.
func targetFormData(data []byte) []byte {
	cloned := append([]byte(nil), data...)
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
		// Current simple forms place buttons and display-only elements in one
		// elements array. Protocol 486 requires a buttons array containing old
		// button objects without the current-only type discriminator.
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
			// Header and divider did not exist as custom-form element types in
			// 1.18. Convert each to one read-only label so response array indexes
			// remain aligned with Dragonfly's current form submission handler.
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
