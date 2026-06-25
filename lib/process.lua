-- Spawn subprocesses without a flashing console on Windows.

local process = {
    last_backend = 'shell',
}

local win_spawn -- false = unavailable, function = ready

local function is_windows()
    if type(jit) == 'table' and jit.os == 'Windows' then
        return true
    end
    return package.config:sub(1, 1) == '\\'
end

local function quote_shell(arg)
    arg = tostring(arg or '')
    if arg:find(' ', 1, true) or arg:find('"', 1, true) then
        return '"' .. arg:gsub('"', '\\"') .. '"'
    end
    return arg
end

local function quote_win(arg)
    arg = tostring(arg or '')
    if arg == '' then
        return '""'
    end
    if arg:find('[%s"]', 1, false) then
        return '"' .. arg:gsub('"', '\\"') .. '"'
    end
    return arg
end

local function shell_command(bin, args, opts)
    local parts = { quote_shell(bin) }
    for _, a in ipairs(args) do
        parts[#parts + 1] = quote_shell(a)
    end
    local prefix = ''
    if opts and opts.debug then
        prefix = 'set MACROMOG_DEBUG=1&& '
    end
    return prefix .. table.concat(parts, ' ') .. ' 2>&1'
end

local function build_cmdline(bin, args)
    local parts = { quote_win(bin) }
    for _, a in ipairs(args) do
        parts[#parts + 1] = quote_win(a)
    end
    return table.concat(parts, ' ')
end

local function init_win_spawn()
    if win_spawn ~= nil then
        return win_spawn ~= false
    end

    if not is_windows() then
        win_spawn = false
        return false
    end

    if type(jit) == 'table' and type(jit.on) == 'function' then
        jit.on()
    end

    local ok, ffi = pcall(require, 'ffi')
    if not ok then
        win_spawn = false
        return false
    end

    ffi.cdef([[
        typedef unsigned long DWORD;
        typedef int BOOL;
        typedef void* HANDLE;
        typedef const char* LPCSTR;
        typedef char* LPSTR;
        typedef struct {
            DWORD cb;
            LPSTR lpReserved;
            LPSTR lpDesktop;
            LPSTR lpTitle;
            DWORD dwX;
            DWORD dwY;
            DWORD dwXSize;
            DWORD dwYSize;
            DWORD dwXCountChars;
            DWORD dwYCountChars;
            DWORD dwFillAttribute;
            DWORD dwFlags;
            unsigned short wShowWindow;
            unsigned short cbReserved2;
            unsigned char* lpReserved2;
            HANDLE hStdInput;
            HANDLE hStdOutput;
            HANDLE hStdError;
        } STARTUPINFOA;
        typedef struct {
            HANDLE hProcess;
            HANDLE hThread;
            DWORD dwProcessId;
            DWORD dwThreadId;
        } PROCESS_INFORMATION;
        typedef struct {
            DWORD nLength;
            void* lpSecurityDescriptor;
            BOOL bInheritHandle;
        } SECURITY_ATTRIBUTES;
        BOOL __stdcall CreatePipe(HANDLE* hReadPipe, HANDLE* hWritePipe,
            SECURITY_ATTRIBUTES* lpPipeAttributes, DWORD nSize);
        BOOL __stdcall SetHandleInformation(HANDLE hObject, DWORD dwMask, DWORD dwFlags);
        BOOL __stdcall CreateProcessA(LPCSTR lpApplicationName, LPSTR lpCommandLine,
            void* lpProcessAttributes, void* lpThreadAttributes, BOOL bInheritHandles,
            DWORD dwCreationFlags, void* lpEnvironment, LPCSTR lpCurrentDirectory,
            STARTUPINFOA* lpStartupInfo, PROCESS_INFORMATION* lpProcessInformation);
        BOOL __stdcall ReadFile(HANDLE hFile, void* lpBuffer, DWORD nNumberOfBytesToRead,
            DWORD* lpNumberOfBytesRead, void* lpOverlapped);
        BOOL __stdcall PeekNamedPipe(HANDLE hNamedPipe, void* lpBuffer, DWORD nBufferSize,
            DWORD* lpBytesRead, DWORD* lpTotalBytesAvail, DWORD* lpBytesLeftThisMessage);
        DWORD __stdcall WaitForSingleObject(HANDLE hHandle, DWORD dwMilliseconds);
        BOOL __stdcall GetExitCodeProcess(HANDLE hProcess, DWORD* lpExitCode);
        BOOL __stdcall CloseHandle(HANDLE hObject);
        HANDLE __stdcall GetStdHandle(DWORD nStdHandle);
        DWORD __stdcall GetLastError(void);
        typedef struct {
            unsigned long dwLowDateTime;
            unsigned long dwHighDateTime;
        } FILETIME;
        typedef struct {
            unsigned long dwFileAttributes;
            FILETIME ftCreationTime;
            FILETIME ftLastAccessTime;
            FILETIME ftLastWriteTime;
            unsigned long nFileSizeHigh;
            unsigned long nFileSizeLow;
        } WIN32_FILE_ATTRIBUTE_DATA;
        BOOL __stdcall GetFileAttributesExA(LPCSTR lpFileName, int fInfoLevelId,
            WIN32_FILE_ATTRIBUTE_DATA* lpFileInformation);
    ]])

    local kernel32 = ffi.load('kernel32')

    local STARTF_USESTDHANDLES = 0x00000100
    local STARTF_USESHOWWINDOW = 0x00000001
    local SW_HIDE = 0
    local HANDLE_FLAG_INHERIT = 0x00000001
    local INFINITE = 0xFFFFFFFF
    local CREATE_NO_WINDOW = 0x08000000
    local STD_INPUT_HANDLE = -10

    local function drain_output(read_handle, proc_handle)
        local chunks = {}
        local buf = ffi.new('char[8192]')
        local bytes_read = ffi.new('DWORD[1]')
        local avail = ffi.new('DWORD[1]')

        while true do
            avail[0] = 0
            kernel32.PeekNamedPipe(read_handle, nil, 0, nil, avail, nil)

            if avail[0] > 0 then
                local to_read = math.min(avail[0], 8192)
                if kernel32.ReadFile(read_handle, buf, to_read, bytes_read, nil) ~= 0 and bytes_read[0] > 0 then
                    chunks[#chunks + 1] = ffi.string(buf, bytes_read[0])
                end
            end

            if kernel32.WaitForSingleObject(proc_handle, 0) == 0 then
                while kernel32.ReadFile(read_handle, buf, 8192, bytes_read, nil) ~= 0 and bytes_read[0] > 0 do
                    chunks[#chunks + 1] = ffi.string(buf, bytes_read[0])
                end
                break
            end

            if avail[0] == 0 then
                kernel32.WaitForSingleObject(proc_handle, 10)
            end
        end

        return table.concat(chunks)
    end

    win_spawn = function(bin, args, opts)
        if opts and opts.debug then
            return nil
        end

        local cmdline = build_cmdline(bin, args)
        local cmdline_buf = ffi.new('char[?]', #cmdline + 1)
        ffi.copy(cmdline_buf, cmdline)

        local sa = ffi.new('SECURITY_ATTRIBUTES')
        sa.nLength = ffi.sizeof('SECURITY_ATTRIBUTES')
        sa.bInheritHandle = 1

        local read_pipe = ffi.new('HANDLE[1]')
        local write_pipe = ffi.new('HANDLE[1]')
        if kernel32.CreatePipe(read_pipe, write_pipe, sa, 0) == 0 then
            return nil
        end

        if kernel32.SetHandleInformation(read_pipe[0], HANDLE_FLAG_INHERIT, 0) == 0 then
            kernel32.CloseHandle(read_pipe[0])
            kernel32.CloseHandle(write_pipe[0])
            return nil
        end

        local si = ffi.new('STARTUPINFOA')
        local pi = ffi.new('PROCESS_INFORMATION')
        ffi.fill(si, ffi.sizeof('STARTUPINFOA'), 0)
        si.cb = ffi.sizeof('STARTUPINFOA')
        si.dwFlags = STARTF_USESTDHANDLES + STARTF_USESHOWWINDOW
        si.wShowWindow = SW_HIDE
        si.hStdInput = kernel32.GetStdHandle(STD_INPUT_HANDLE)
        si.hStdOutput = write_pipe[0]
        si.hStdError = write_pipe[0]

        local created = kernel32.CreateProcessA(nil, cmdline_buf, nil, nil, 1, CREATE_NO_WINDOW, nil, nil, si, pi)

        kernel32.CloseHandle(write_pipe[0])

        if created == 0 then
            kernel32.CloseHandle(read_pipe[0])
            return nil
        end

        kernel32.CloseHandle(pi.hThread)

        local output = drain_output(read_pipe[0], pi.hProcess)
        kernel32.CloseHandle(read_pipe[0])
        kernel32.WaitForSingleObject(pi.hProcess, INFINITE)

        local exit_code = ffi.new('DWORD[1]')
        kernel32.GetExitCodeProcess(pi.hProcess, exit_code)
        kernel32.CloseHandle(pi.hProcess)

        local code = tonumber(exit_code[0]) or 1

        return {
            read = function(_, mode)
                if mode == '*a' then
                    return output
                end
                return output
            end,
            close = function()
                if code == 0 then
                    return true, 'exit', 0
                end
                return false, 'exit', code
            end,
        }
    end

    return true
end

function process.popen(bin, args, opts)
    if init_win_spawn() and win_spawn then
        local pipe = win_spawn(bin, args, opts)
        if pipe then
            process.last_backend = 'ffi'
            return pipe
        end
    end

    process.last_backend = 'shell'
    return io.popen(shell_command(bin, args, opts), 'r')
end

function process.file_mtime(path)
    if init_win_spawn() and win_spawn then
        local ffi = require('ffi')
        local kernel32 = ffi.load('kernel32')
        local data = ffi.new('WIN32_FILE_ATTRIBUTE_DATA')
        local normalized = tostring(path):gsub('/', '\\')
        if kernel32.GetFileAttributesExA(normalized, 0, data) ~= 0 then
            local high = tonumber(data.ftLastWriteTime.dwHighDateTime) or 0
            local low = tonumber(data.ftLastWriteTime.dwLowDateTime) or 0
            return string.format('%08x%08x', high, low)
        end
    end

    local handle = io.popen('cmd /c for %I in ("' .. tostring(path):gsub('"', '') .. '") do @echo %~tI')
    if not handle then
        return nil
    end
    local line = handle:read('*l')
    handle:close()
    return line
end

return process
