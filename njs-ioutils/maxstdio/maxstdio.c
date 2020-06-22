#include "maxstdio.h"
#include <stdio.h>

int _getmaxstdio();
int _setmaxstdio(int new_max);

int CGetMaxSTDIO() {
	return _getmaxstdio();
}

int CSetMaxSTDIO(int new_max) {
	return _setmaxstdio(new_max);
}
