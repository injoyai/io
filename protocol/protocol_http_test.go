package protocol

import (
	"testing"
)

func TestNewHTTPResponse(t *testing.T) {
	t.Log(string(NewHTTPResponseBytes(200, []byte("{}"))))
}
