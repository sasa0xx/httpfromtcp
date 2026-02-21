package headers

import (
	"bytes"
	"errors"
	"strings"
)

var crfl = []byte("\r\n")
var ERROR_INVALID_FIELD_LINE = errors.New("Invalid field-line")
var ERROR_DUPLUCATED_FIELD_LINE = errors.New("Duplucated field-line")

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(name string) string {
	return h[strings.ToLower(name)]
}

func (h Headers) Set(field_name string, field_value string) {
	h[strings.ToLower(field_name)] = field_value
}

func (h Headers) Add(field_name string, field_value string) {
	key := strings.ToLower(field_name)
	if val, ok := h[key]; ok {
		h[key] = val + ", " + field_value
	} else {
		h[key] = field_value
	}
}

func (h Headers) Delete(field_name string) {
	key := strings.ToLower(field_name)
	delete(h, key)
}

func (h Headers) Parse(data []byte) (int, bool, error) { // this function should parse one header at a time
	idx := bytes.Index(data, crfl)
	if idx == -1 {
		return 0, false, nil
	}

	header := data[:idx]
	if len(header) == 0 {
		return len(crfl), true, nil
	}

	parts := bytes.SplitN(header, []byte(":"), 2)
	if len(parts) != 2 || bytes.ContainsAny(parts[0], " \t") { // field name can't have spaces
		return 0, false, ERROR_INVALID_FIELD_LINE
	}

	field_name := strings.ToLower(string(parts[0]))
	field_value := string(bytes.TrimSpace(parts[1]))

	h.Add(field_name, field_value)

	return idx + len(crfl), false, nil
}
