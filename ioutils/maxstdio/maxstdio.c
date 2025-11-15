#include "maxstdio.h"
#include <stdio.h>

// CGetMaxSTDIO wraps the Windows CRT _getmaxstdio() function.
// Returns the current maximum number of simultaneously open files.
int CGetMaxSTDIO() {
	return _getmaxstdio();
}

// CSetMaxSTDIO wraps the Windows CRT _setmaxstdio() function.
// Sets the maximum number of simultaneously open files.
// Returns the previous maximum value.
int CSetMaxSTDIO(int new_max) {
	return _setmaxstdio(new_max);
}
