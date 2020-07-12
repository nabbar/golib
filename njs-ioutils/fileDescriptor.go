/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package njs_ioutils

/**
 * SystemFileDescriptor is returning current Limit & max system limit for file descriptor (open file or I/O resource) currently set in the system
 * This function return the current setting (current number of file descriptor and the max value) if the newValue given is zero
 * Otherwise if the newValue is more than the current system limit, try to change the current limit in the system for this process only
 *
 *  For Windows build, please follow this note :
 *	1) install package gcc-multilib gcc-mingw-w64 to build C source with GCC
 *		you will having this binaries
 *			- i686-w64-mingw32* for 32-bit Windows;
 *			- x86_64-w64-mingw32* for 64-bit Windows.
 *		locate you binaries gcc mingw path and note it:
 *			- win32 : updatedb && locate i686-w64-mingw32-gcc
 *			- win64 : updatedb && locate x86_64-w64-mingw32-gcc
 *	2) if you have an error in the build, or if the .o object file is not present in golib/njg-ioutils/maxstdio/, run this step
 *		- go to golib/njg-ioutils/maxstdio folder
 *		- gcc -c maxstdio.c
 *	3) for Win32 use this env var in prefix of your go build command (recommend to use -a flag) :
 *		CC=/usr/bin/i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -a -v ...
 *	4) for win64 use this env var in prefix of your go build command (recommend to use -a flag) :
 *		CC=/usr/bin/x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -a -v ...
 *
 *	Normally no problem will be result in the build
 *
 */
func SystemFileDescriptor(newValue int) (current int, max int, err error) {
	return systemFileDescriptor(newValue)
}
