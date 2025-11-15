#ifndef _MAXSTDIO_H
#define _MAXSTDIO_H

// CGetMaxSTDIO returns the current maximum number of simultaneously open files.
int CGetMaxSTDIO();

// CSetMaxSTDIO sets the maximum number of simultaneously open files.
// Returns the previous maximum value.
int CSetMaxSTDIO(int new_max);

#endif