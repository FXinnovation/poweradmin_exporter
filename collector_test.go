package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	assert.Equal(t, &Collector{}, c)
}

func TestCollector_Collect(t *testing.T) {
	c:= NewCollector()


}
