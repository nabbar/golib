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

package head

var _defaultConfig = []byte(`{
   "Content-Security-Policy":"default-src 'self' 'unsafe-inline' 'unsafe-eval'; img-src 'self' data: image/svg+xml*; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'self'; child-src 'none'; frame-src 'none'; worker-src 'none'; frame-ancestors 'none'; form-action 'none'; upgrade-insecure-requests 1; block-all-mixed-content; disown-opener; require-sri-for script style; sandbox allow-same-origin allow-scripts; reflected-xss block; referrer no-referrer",
   "Feature-Policy":"geolocation 'self'; midi 'self'; notifications 'self'; push 'self'; sync-xhr 'self'; microphone 'self'; camera 'self'; magnetometer 'self'; gyroscope 'self'; speaker 'self'; vibrate 'self'; fullscreen 'self'; payment 'self';",
   "Strict-Transport-Security":"max-age=1; preload; includeSubDomains",
   "X-Frame-Options":"DENY",
   "X-Xss-Protection":"1; mode=block",
   "X-Content-Type-Options":"nosniff",
   "Referrer-Policy":"no-referrer"
}`)

func (c *componentHead) DefaultConfig() []byte {
	return _defaultConfig
}

func SetDefaultConfig(cfg []byte) {
	_defaultConfig = cfg
}