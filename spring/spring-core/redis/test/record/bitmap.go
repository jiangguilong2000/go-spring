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

package record

import (
	"testing"

	"github.com/go-spring/spring-core/redis"
	"github.com/go-spring/spring-core/redis/test/cases"
)

func BitCount(t *testing.T, d redis.Driver) {
	RunCase(t, d, cases.BitCount)
}

func BitOpAnd(t *testing.T, d redis.Driver) {
	RunCase(t, d, cases.BitOpAnd)
}

func BitPos(t *testing.T, d redis.Driver) {
	RunCase(t, d, cases.BitPos)
}

func GetBit(t *testing.T, d redis.Driver) {
	RunCase(t, d, cases.GetBit)
}

func SetBit(t *testing.T, d redis.Driver) {
	RunCase(t, d, cases.SetBit)
}
