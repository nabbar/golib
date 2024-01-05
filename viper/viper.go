/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

package viper

import "time"

func (v *viper) GetBool(key string) bool {
	return v.Viper().GetBool(key)
}

func (v *viper) GetString(key string) string {
	return v.Viper().GetString(key)
}

func (v *viper) GetInt(key string) int {
	return v.Viper().GetInt(key)
}

func (v *viper) GetInt32(key string) int32 {
	return v.Viper().GetInt32(key)
}

func (v *viper) GetInt64(key string) int64 {
	return v.Viper().GetInt64(key)
}

func (v *viper) GetUint(key string) uint {
	return v.Viper().GetUint(key)
}

func (v *viper) GetUint16(key string) uint16 {
	return v.Viper().GetUint16(key)
}

func (v *viper) GetUint32(key string) uint32 {
	return v.Viper().GetUint32(key)
}

func (v *viper) GetUint64(key string) uint64 {
	return v.Viper().GetUint64(key)
}

func (v *viper) GetFloat64(key string) float64 {
	return v.Viper().GetFloat64(key)
}

func (v *viper) GetTime(key string) time.Time {
	return v.Viper().GetTime(key)
}

func (v *viper) GetDuration(key string) time.Duration {
	return v.Viper().GetDuration(key)
}

func (v *viper) GetIntSlice(key string) []int {
	return v.Viper().GetIntSlice(key)
}

func (v *viper) GetStringSlice(key string) []string {
	return v.Viper().GetStringSlice(key)
}

func (v *viper) GetStringMap(key string) map[string]any {
	return v.Viper().GetStringMap(key)
}

func (v *viper) GetStringMapString(key string) map[string]string {
	return v.Viper().GetStringMapString(key)
}

func (v *viper) GetStringMapStringSlice(key string) map[string][]string {
	return v.Viper().GetStringMapStringSlice(key)
}
