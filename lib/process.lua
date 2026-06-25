-- Spawn subprocesses without a flashing console on Windows.

local process = {}

local function is_windows()
    return type(jit) == 'table' and jit.os == 'Windows'
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

local win_spawn

if is_windows() then
    local ok, ffi = pcall(require, 'ffi')
    if ok then
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
            BOOL CreatePipe(HANDLE* hReadPipe, HANDLE* hWritePipe,
                SECURITY_ATTRIBUTES* lpPipeAttributes, DWORD nSize);
            BOOL SetHandleInformation(HANDLE hObject, DWORD dwMask, DWORD dwFlags);
            BOOL CreateProcessA(LPCSTR lpApplicationName, LPSTR lpCommandLine,
                void* lpProcessAttributes, void* lpThreadAttributes, BOOL bInheritHandles,
                DWORD dwCreationFlags, void* lpEnvironment, LPCSTR lpCurrentDirectory,
                STARTUPINFOA* lpStartupInfo, PROCESS_INFORMATION* lpProcessInformation);
            BOOL ReadFile(HANDLE hFile, void* lpBuffer, DWORD nNumberOfBytesToRead,
                DWORD* lpNumberOfBytesRead, void* lpOverlapped);
            DWORD WaitForSingleObject(HANDLE hHandle, DWORD dwMilliseconds);
            BOOL GetExitCodeProcess(HANDLE hProcess, DWORD* lpExitCode);
            BOOL CloseHandle(HANDLE hObject);
            HANDLE GetStdHandle(DWORD nStdHandle);
        ]])

        local kernel32 = ffi.load('kernel32')

        local STARTF_USESTDHANDLES = 0x00000100
        local STARTF_USESHOWWINDOW = 0x00000001
        local SW_HIDE = 0
        local HANDLE_FLAG_INHERIT = 0x00000001
        local INFINITE = 0xFFFFFFFF
        local CREATE_NO_WINDOW = 0x08000000
        local STD_INPUT_HANDLE = -10

        local function build_cmdline(bin, args)
            local parts = { quote_win(bin) }
            for _, a in ipairs(args) do
                parts[#parts + 1] = quote_win(a)
            end
            return table.concat(parts, ' ')
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

            local created = kernel32.CreateProcessA(bin, cmdline_buf, nil, nil, 1, CREATE_NO_WINDOW, nil, nil, si, pi)

            kernel32.CloseHandle(write_pipe[0])

            if created == 0 then
                kernel32.CloseHandle(read_pipe[0])
                return nil
            end

            kernel32.CloseHandle(pi.hThread)

            local chunks = {}
            local buf = ffi.new('char[4096]')
            local bytes_read = ffi.new('DWORD[1]')
            while true do
                if kernel32.ReadFile(read_pipe[0], buf, 4096, bytes_read, nil) == 0 then
                    break
                end
                if bytes_read[0] == 0 then
                    break
                end
                chunks[#chunks + 1] = ffi.string(buf, bytes_read[0])
            end

            kernel32.CloseHandle(read_pipe[0])
            kernel32.WaitForSingleObject(pi.hProcess, INFINITE)

            local exit_code = ffi.new('DWORD[1]')
            kernel32.GetExitCodeProcess(pi.hProcess, exit_code)
            kernel32.CloseHandle(pi.hProcess)

            local output = table.concat(chunks)
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
    end
end

function process.popen(bin, args, opts)
    if win_spawn then
        local pipe = win_spawn(bin, args, opts)
        if pipe then
            return pipe
        end
    end

    return io.popen(shell_command(bin, args, opts), 'r')
end

return process
