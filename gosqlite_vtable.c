#include "gosqlite_vtable.h"
#include "sqlite3.h"
#include <stdlib.h>
#include <stdint.h>
#include <string.h>

struct gosqlite_vtab {
	sqlite3_vtab base;
	uintptr_t handle;
};

struct gosqlite_vtab_cursor {
    sqlite3_vtab_cursor base;
    uintptr_t handle;
};

extern int gosqliteConnect(sqlite3 *conn, void *handle, uintptr_t *out, char **errout);
extern int gosqliteCreate(sqlite3 *conn, void *handle, uintptr_t *out, char **errout);
extern int gosqliteDisconnect(uintptr_t handle, char **errout);
extern int gosqliteDestroy(uintptr_t handle, char **errout);
extern int gosqliteBestIndex(uintptr_t handle, void *out, char **errout);
extern int gosqliteOpen(void *handle, uintptr_t *out, char **errout);
extern int gosqliteClose(uintptr_t handle, char **errout);
extern int gosqliteFilter(uintptr_t handle, int indexId, const char *indexName,
                          int argc, sqlite3_value **argv, char **errout);
extern int gosqliteNext(uintptr_t handle, char **errout);
extern int gosqliteEOF(uintptr_t handle);
extern int gosqliteColumn(uintptr_t handle, sqlite3_context *ctx, int column);

static void gosqlite_reset_error(struct sqlite3_vtab *vtable)
{
    if (vtable->zErrMsg == NULL)
        return;

    sqlite3_free(vtable->zErrMsg);
    vtable->zErrMsg = NULL;
}

static int gosqlite_setup_vtable(uintptr_t handle, sqlite3_vtab **vtable)
{
	struct gosqlite_vtab *vtab =
	    sqlite3_malloc(sizeof(struct gosqlite_vtab));
	if (vtab == NULL)
		return SQLITE_NOMEM;

	memset(vtab, 0, sizeof(struct gosqlite_vtab));
	vtab->handle = handle;
	*vtable = &vtab->base;

	return SQLITE_OK;
}

static int gosqlite_connect(sqlite3 *db, void *aux, int argc,
			    const char *const *argv, sqlite3_vtab **vtable,
			    char **error)
{
	uintptr_t handle;
	int rv = gosqliteConnect(db, aux, &handle, error);
	if (rv != SQLITE_OK)
		return rv;

    return gosqlite_setup_vtable(handle, vtable);
}

static int gosqlite_create(sqlite3 *db, void *aux, int argc,
			   const char *const *argv, sqlite3_vtab **vtable,
			   char **error)
{
	uintptr_t handle;
	int rv = gosqliteCreate(db, aux, &handle, error);
	if (rv != SQLITE_OK)
		return rv;

    return gosqlite_setup_vtable(handle, vtable);
}

static int gosqlite_best_index(sqlite3_vtab *vtable, sqlite3_index_info *info)
{
    struct gosqlite_vtab *govtable = (struct gosqlite_vtab *)vtable;

    gosqlite_reset_error(vtable);
    return gosqliteBestIndex(govtable->handle, info, &vtable->zErrMsg);
}

static int gosqlite_disconnect(sqlite3_vtab *vtable)
{
    struct gosqlite_vtab *govtable = (struct gosqlite_vtab *)vtable;

    gosqlite_reset_error(vtable);
    int rv = gosqliteDisconnect(govtable->handle, &vtable->zErrMsg);
    if (rv != SQLITE_OK)
        return rv;

    sqlite3_free(govtable);

    return SQLITE_OK;
}

static int gosqlite_destroy(sqlite3_vtab *vtable)
{
	struct gosqlite_vtab *govtable = (struct gosqlite_vtab *)vtable;

    gosqlite_reset_error(vtable);
    int rv = gosqliteDestroy(govtable->handle, &vtable->zErrMsg);
    if (rv != SQLITE_OK)
        return rv;

    sqlite3_free(govtable);

    return SQLITE_OK;
}

static int gosqlite_open(sqlite3_vtab *vtable, sqlite3_vtab_cursor **out)
{
  struct gosqlite_vtab_cursor *cursor =
      (struct gosqlite_vtab_cursor *)sqlite3_malloc(
          sizeof(struct gosqlite_vtab_cursor));
  if (cursor == NULL)
    return SQLITE_NOMEM;

  cursor->base.pVtab = vtable;
  gosqlite_reset_error(vtable);
  int rv = gosqliteOpen(vtable, &cursor->handle, &vtable->zErrMsg);
  if (rv != SQLITE_OK) return rv;

  *out = &cursor->base;

  return SQLITE_OK;
}

static int gosqlite_close(sqlite3_vtab_cursor *cursor)
{
    struct gosqlite_vtab_cursor *gocursor = (struct gosqlite_vtab_cursor *)cursor;

    gosqlite_reset_error(cursor->pVtab);
    int rv = gosqliteClose(gocursor->handle, &cursor->pVtab->zErrMsg);
    if (rv != SQLITE_OK) return rv;
    sqlite3_free(cursor);

    return SQLITE_OK;
}

static int gosqlite_filter(sqlite3_vtab_cursor *cursor, int indexId,
			   const char *indexName, int argc,
			   sqlite3_value **argv)
{
    struct gosqlite_vtab_cursor *gocursor = (struct gosqlite_vtab_cursor *)cursor;

    gosqlite_reset_error(cursor->pVtab);
    return gosqliteFilter(gocursor->handle, indexId, indexName, argc, argv,
                            &cursor->pVtab->zErrMsg);
}

static int gosqlite_next(sqlite3_vtab_cursor *cursor)
{
    struct gosqlite_vtab_cursor *gocursor = (struct gosqlite_vtab_cursor *)cursor;

    gosqlite_reset_error(cursor->pVtab);
    return gosqliteNext(gocursor->handle, &cursor->pVtab->zErrMsg);
}

static int gosqlite_eof(sqlite3_vtab_cursor *cursor)
{
    struct gosqlite_vtab_cursor *gocursor = (struct gosqlite_vtab_cursor *)cursor;
    return gosqliteEOF(gocursor->handle);
}

static int gosqlite_column(sqlite3_vtab_cursor *cursor, sqlite3_context *ctx,
			   int column)
{
    struct gosqlite_vtab_cursor *gocursor = (struct gosqlite_vtab_cursor *)cursor;

    gosqlite_reset_error(cursor->pVtab);
    return gosqliteColumn(gocursor->handle, ctx, column);
}

static int gosqlite_rowid(sqlite3_vtab_cursor *cursor, sqlite_int64 *rowid)
{
	return SQLITE_OK;
}

static const sqlite3_module gomodule = {
	.iVersion = 0,
	.xCreate = gosqlite_connect,
	.xConnect = gosqlite_connect,
	.xBestIndex = gosqlite_best_index,
	.xDisconnect = gosqlite_disconnect,
	.xDestroy = gosqlite_destroy,
	.xOpen = gosqlite_open,
	.xClose = gosqlite_close,
	.xFilter = gosqlite_filter,
	.xNext = gosqlite_next,
	.xEof = gosqlite_eof,
	.xColumn = gosqlite_column,
	.xRowid = gosqlite_rowid,
	.xUpdate = NULL,
	.xBegin = NULL,
	.xSync = NULL,
	.xCommit = NULL,
	.xRollback = NULL,
	.xFindFunction = NULL,
	.xRename = NULL,
	.xSavepoint = NULL,
	.xRelease = NULL,
	.xRollbackTo = NULL,
	.xShadowName = NULL,
	.xIntegrity = NULL,
};

static const sqlite3_module gomodule_eponymous = {
	.iVersion = 0,
	.xCreate = NULL,
	.xConnect = gosqlite_connect,
	.xBestIndex = gosqlite_best_index,
	.xDisconnect = gosqlite_disconnect,
	.xDestroy = NULL,
	.xOpen = gosqlite_open,
	.xClose = gosqlite_close,
	.xFilter = gosqlite_filter,
	.xNext = gosqlite_next,
	.xEof = gosqlite_eof,
	.xColumn = gosqlite_column,
	.xRowid = gosqlite_rowid,
	.xUpdate = NULL,
	.xBegin = NULL,
	.xSync = NULL,
	.xCommit = NULL,
	.xRollback = NULL,
	.xFindFunction = NULL,
	.xRename = NULL,
	.xSavepoint = NULL,
	.xRelease = NULL,
	.xRollbackTo = NULL,
	.xShadowName = NULL,
	.xIntegrity = NULL,
};

int gosqlite_create_module(sqlite3 *db, const char *name, uintptr_t module)
{
	return sqlite3_create_module(db, name, &gomodule, (void *)module);
}

int gosqlite_create_eponymous_module(sqlite3 *db, const char *name,
				     uintptr_t module)
{
	return sqlite3_create_module(db, name, &gomodule_eponymous,
				     (void *)module);
}

void gosqlite_free(void *ptr)
{
    free(ptr);
}
