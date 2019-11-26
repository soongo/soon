// Copyright 2019 Guoyao Wu. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package soon

import (
	"fmt"
	"strings"
)

// IsDebugging returns true if the framework is running in debug mode.
// Use SetMode(soon.DebugMode) to enable debug mode.
func IsDebugging() bool {
	return modeCode == debugCode
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		if !strings.HasSuffix(format, "\n") {
			format += "\n"
		}
		fmt.Fprintf(DefaultWriter, "[SOON-debug] "+format, values...)
	}
}
