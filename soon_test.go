// Copyright 2020 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"testing"
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
	if err != nil {
		t.Error(err)
	}
	if statusCode != 200 {
		t.Errorf(testErrorFormat, statusCode, 200)
	}
	if body != "hello" {
		t.Errorf(testErrorFormat, body, "hello")
	}
}
