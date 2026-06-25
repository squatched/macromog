-- Spawn subprocesses without a flashing console on Windows.
-- Uses macromog_spawn.dll when available (no cmd.exe flash).
-- Falls back to io.popen otherwise.

local process = {
    last_backend = 'uninitialized',
}

local spawn_ok, spawn_mod = pcall(require, 'macromog_spawn')

local function quote_shell(arg)
    arg = tostring(arg or '')
    if arg:find(' ', 1, true) or arg:find('"', 1, true) then
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

function process.popen(bin, args, opts)
    if spawn_ok and not (opts and opts.debug) then
        local output, code = spawn_mod.spawn(bin, args)
        if output ~= nil then
            process.last_backend = 'dll'
            local captured = output
            local exit_code = tonumber(code) or 1
            return {
                read = function(_, _mode)
                    return captured
                end,
                close = function()
                    if exit_code == 0 then
                        return true, 'exit', 0
                    end
                    return false, 'exit', exit_code
                end,
            }
        end
        process.last_backend = 'dll-failed'
    else
        process.last_backend = spawn_ok and 'dll-debug-skip' or 'dll-unavailable'
    end

    return io.popen(shell_command(bin, args, opts), 'r')
end

-- Returns a hex string timestamp for the file's last write time, or nil.
-- Populated by macromog_spawn.dll; returns nil until the DLL is available.
function process.file_mtime(path)
    if spawn_ok and spawn_mod.file_mtime then
        return spawn_mod.file_mtime(path)
    end
    return nil
end

return process
