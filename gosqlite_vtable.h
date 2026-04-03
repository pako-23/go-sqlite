#ifndef GOSQLITE_VTABLE_H_INCLUDED
#define GOSQLITE_VTABLE_H_INCLUDED

#include "sqlite3.h"
#include <stdint.h>

extern int gosqlite_create_module(sqlite3 * db, const char *name,
				  uintptr_t module);
extern int gosqlite_create_eponymous_module(sqlite3 * db, const char *name,
					    uintptr_t module);

#endif
