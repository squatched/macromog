-- Opt-in debug logging for the addon (file + Windower console, never FFXI chat).

local M = {
    enabled = false,
    session = false,
}

local CHAT_COLOR = 207
local FLAG_NAME = '.debug'
local LOG_NAME = 'debug.log'

local function flag_path()
    return windower.addon_path .. 'data/' .. FLAG_NAME
end

function M.log_path()
    return windower.addon_path .. 'data/' .. LOG_NAME
end

function M.refresh()
    M.enabled = windower.file_exists(flag_path())
end

function M.active()
    return M.enabled or M.session
end

function M.begin_session()
    M.session = true
end

function M.end_session()
    M.session = false
    M.refresh()
end

function M.enable_persist()
    local path = flag_path()
    local f = io.open(path, 'w')
    if f then
        f:close()
        M.enabled = true
    end
end

function M.disable_persist()
    os.remove(flag_path())
    M.enabled = false
end

function M.user(msg)
    windower.add_to_chat(CHAT_COLOR, '[Macromog] ' .. tostring(msg))
end

function M.debug(msg)
    if not M.active() then
        return
    end
    local stamp = os.date('!%Y-%m-%d %H:%M:%S')
    local line = stamp .. ' ' .. tostring(msg)
    local f = io.open(M.log_path(), 'a')
    if f then
        f:write(line .. '\n')
        f:close()
    end
    if windower.debug then
        windower.debug('[Macromog] ' .. line)
    end
end

function M.reset_log()
    local f = io.open(M.log_path(), 'w')
    if f then
        f:write(os.date('!%Y-%m-%d %H:%M:%S') .. ' === Macromog debug log ===\n')
        f:close()
    end
end

return M
