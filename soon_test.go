// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApp_Run(t *testing.T) {
	app, c := New(), make(chan struct{})
	go func() {
		c <- struct{}{}
		app.Run(":54321")
	}()
	<-c
	app.GET("/", func(c *Context) {
		c.Send("hello")
	})
	statusCode, _, body, err := request("GET", "http://localhost:54321", nil)
	assert := assert.New(t)
	assert.Nil(err)
	assert.Equal(200, statusCode)
	assert.Equal("hello", body)
}
