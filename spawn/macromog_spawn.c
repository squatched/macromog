#define WIN32_LEAN_AND_MEAN
#include <windows.h>

#include "helpers.h"

/* ── Minimal Lua 5.1 API ───────────────────────────────────────────── */

typedef struct lua_State lua_State;
typedef double lua_Number;
typedef int (*lua_CFunction)(lua_State *L);

#define LUA_TTABLE 5

typedef struct {
    const char *name;
    lua_CFunction func;
} luaL_Reg;

extern void lua_settop(lua_State *L, int idx);
extern void lua_pushnil(lua_State *L);
extern void lua_pushnumber(lua_State *L, lua_Number n);
extern void lua_pushlstring(lua_State *L, const char *s, size_t len);
extern void lua_rawgeti(lua_State *L, int idx, int n);
extern size_t lua_objlen(lua_State *L, int idx);
extern const char *lua_tolstring(lua_State *L, int idx, size_t *len);
extern int lua_type(lua_State *L, int idx);
extern void luaL_checktype(lua_State *L, int narg, int t);
extern const char *luaL_checklstring(lua_State *L, int narg, size_t *len);
extern void luaL_register(lua_State *L, const char *lib, const luaL_Reg *l);

#define lua_pop(L, n) lua_settop(L, -(n) - 1)
#define lua_tostring(L, i) lua_tolstring(L, (i), NULL)
#define luaL_checkstring(L, n) luaL_checklstring(L, (n), NULL)

/* ── spawn(bin, args) -> output, exit_code ─────────────────────────── */

static int l_spawn(lua_State *L)
{
    const char *bin = luaL_checkstring(L, 1);
    luaL_checktype(L, 2, LUA_TTABLE);

    char cmdline[32768];
    size_t pos = 0;
    cmdline[0] = '\0';
    push_quoted(cmdline, &pos, sizeof(cmdline), bin);

    int i, n = (int)lua_objlen(L, 2);
    for (i = 1; i <= n; i++) {
        lua_rawgeti(L, 2, i);
        const char *arg = lua_tostring(L, -1);
        if (arg)
            push_quoted(cmdline, &pos, sizeof(cmdline), arg);
        lua_pop(L, 1);
    }

    SECURITY_ATTRIBUTES sa = {sizeof(SECURITY_ATTRIBUTES), NULL, TRUE};

    HANDLE hRead = INVALID_HANDLE_VALUE, hWrite = INVALID_HANDLE_VALUE;
    if (!CreatePipe(&hRead, &hWrite, &sa, 0)) {
        lua_pushnil(L);
        lua_pushnumber(L, (lua_Number)GetLastError());
        return 2;
    }
    SetHandleInformation(hRead, HANDLE_FLAG_INHERIT, 0);

    STARTUPINFOA si = {0};
    si.cb = sizeof(si);
    si.dwFlags = STARTF_USESTDHANDLES;
    si.hStdInput = NULL;
    si.hStdOutput = hWrite;
    si.hStdError = hWrite;

    PROCESS_INFORMATION pi = {0};

    BOOL ok =
        CreateProcessA(NULL, cmdline, NULL, NULL, TRUE, CREATE_NO_WINDOW, NULL, NULL, &si, &pi);
    CloseHandle(hWrite);

    if (!ok) {
        CloseHandle(hRead);
        lua_pushnil(L);
        lua_pushnumber(L, (lua_Number)GetLastError());
        return 2;
    }

    DWORD cap = 65536, used = 0;
    char *buf = (char *)HeapAlloc(GetProcessHeap(), 0, cap);
    if (buf) {
        char chunk[4096];
        DWORD nr;
        while (ReadFile(hRead, chunk, sizeof(chunk), &nr, NULL) && nr > 0) {
            if (used + nr >= cap) {
                cap = (used + nr + 1) * 2;
                char *tmp = (char *)HeapReAlloc(GetProcessHeap(), 0, buf, cap);
                if (!tmp)
                    break;
                buf = tmp;
            }
            mem_copy(buf + used, chunk, nr);
            used += nr;
        }
    }
    CloseHandle(hRead);

    WaitForSingleObject(pi.hProcess, INFINITE);
    DWORD exit_code = 1;
    GetExitCodeProcess(pi.hProcess, &exit_code);
    CloseHandle(pi.hProcess);
    CloseHandle(pi.hThread);

    if (buf) {
        lua_pushlstring(L, buf, used);
        HeapFree(GetProcessHeap(), 0, buf);
    } else {
        lua_pushlstring(L, "", 0);
    }
    lua_pushnumber(L, (lua_Number)exit_code);
    return 2;
}

/* ── file_mtime(path) -> hex_string | nil ──────────────────────────── */

static int l_file_mtime(lua_State *L)
{
    const char *path = luaL_checkstring(L, 1);

    WIN32_FILE_ATTRIBUTE_DATA data;
    if (!GetFileAttributesExA(path, GetFileExInfoStandard, &data)) {
        lua_pushnil(L);
        return 1;
    }

    char stamp[17];
    filetime_to_hex((uint32_t)data.ftLastWriteTime.dwHighDateTime,
                    (uint32_t)data.ftLastWriteTime.dwLowDateTime, stamp);
    lua_pushlstring(L, stamp, 16);
    return 1;
}

/* ── Module entry point ────────────────────────────────────────────── */

static const luaL_Reg funcs[] = {
    {"spawn", l_spawn},
    {"file_mtime", l_file_mtime},
    {NULL, NULL},
};

__declspec(dllexport) int luaopen_macromog_spawn(lua_State *L)
{
    luaL_register(L, "macromog_spawn", funcs);
    return 1;
}
