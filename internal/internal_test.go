package internal

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter_Healthcheck(t *testing.T) {
	r := NewRouter("https://0.0.0.0")

	server := httptest.NewServer(r)
	defer server.Close()

	get, err := http.Get(server.URL + "/healthcheck")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, get.StatusCode)
}
