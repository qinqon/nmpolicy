package expender

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

type JSONWrapper struct {
	Map         map[string]JSONWrapper
	Slice       []JSONWrapper
	Scalar      json.RawMessage
	Placeholder interface{}
}

func (w *JSONWrapper) UnmarshalJSON(b []byte) error {
	if b[0] == '{' {
		fmt.Println("struct")
		w.Map = map[string]JSONWrapper{}
		if err := json.Unmarshal(b, &w.Map); err != nil {
			return err
		}
	} else if b[0] == '[' {
		fmt.Println("list")
		w.Slice = []JSONWrapper{}
		if err := json.Unmarshal(b, &w.Slice); err != nil {
			return err
		}
	} else if strings.HasPrefix(string(b), "\"{{") {
		w.Placeholder = map[string]interface{}{
			"name": "eth1",
			"ipv4": map[string]interface{}{
				"address": "192.16.1.10",
			},
		}
	} else {
		fmt.Println("scalar")
		if err := json.Unmarshal(b, &w.Scalar); err != nil {
			return err
		}
	}
	return nil
}

func (w JSONWrapper) MarshalJSON() ([]byte, error) {
	if w.Map != nil {
		return json.Marshal(w.Map)
	}
	if w.Slice != nil {
		return json.Marshal(w.Slice)
	}
	if w.Scalar != nil {
		return json.Marshal(w.Scalar)
	}
	if w.Placeholder != nil {
		return json.Marshal(w.Placeholder)
	}
	return nil, nil
}

func TestExpander(t *testing.T) {
	template := `
interfaces:
- name: br1
  description: Linux bridge with base interface as a port
  type: linux-bridge
  state: up
  ipv4: "{{ capture.base-iface.interfaces[0].ipv4 }}"
  bridge:
    options:
      stp:
        enabled: false
    port:
    - name: "{{ capture.base-iface.interfaces[0].name }}"
routes:
  config: "{{ capture.bridge-routes-takeover.running }}"
`

	templateStruct := &JSONWrapper{}
	err := yaml.Unmarshal([]byte(template), &templateStruct)
	assert.NoError(t, err)

	raw, err := yaml.Marshal(templateStruct)
	assert.NoError(t, err)
	fmt.Println(string(raw))
}
