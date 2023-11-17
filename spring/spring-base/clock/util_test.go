/*
 * Copyright 2012-2019 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package clock_test

import (
	"testing"
	"time"

	"github.com/jiangguilong2000/go-spring/spring-base/assert"
	"github.com/jiangguilong2000/go-spring/spring-base/clock"
)

func TestUnixMilli(t *testing.T) {
	assert.Equal(t, clock.UnixMilli(time.Unix(5, 0)), int64(5000))
}
