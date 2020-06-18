#include "maxstdio.h"
#include <stdio.h>

int CGetMaxSTDIO() {
	return _getmaxstdio();
}

int CSetMaxSTDIO(int new_max) {
	return _setmaxstdio(new_max);
}
