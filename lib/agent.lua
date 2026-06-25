-- Talk to a long-running macromog agent over TCP (avoids io.popen/cmd.exe).

local json = require('lib/json')
local log = require('lib/log')

local agent = {
    last_backend = 'direct',
}

local socket_ok, socket = pcall(require, 'socket')

local function normalize_slash(path)
    return tostring(path or ''):gsub('\\', '/')
end

local function wine_prefix()
    return os.getenv('WINEPREFIX') or os.getenv('WINE_PREFIX')
end

function agent.wine_on_linux()
    if wine_prefix() then
        return true
    end
    local addon = normalize_slash(windower.addon_path)
    return addon:match('^[Zz]:/') ~= nil
end

function agent.host_path(path)
    path = normalize_slash(path)
    if path:match('^[Zz]:/') then
        return path:sub(3)
    end
    if path:match('^[Cc]:/') then
        local prefix = wine_prefix()
        if prefix then
            return prefix .. '/drive_c/' .. path:sub(4)
        end
    end
    return path
end

local function port_path()
    return windower.addon_path .. 'data/agent.port'
end

local function read_portfile()
    local f = io.open(port_path(), 'r')
    if not f then
        return nil
    end
    local line = f:read('*l')
    f:close()
    if not line or line == '' then
        return nil
    end
    return line
end

local function loopback_host(host)
    if not host then
        return false
    end
    host = host:lower()
    return host == '127.0.0.1' or host == 'localhost' or host == '::1'
end

local function parse_addr(addr)
    local host, port = addr:match('^%[(.+)%]:(%d+)$')
    if host then
        return host, tonumber(port)
    end
    host, port = addr:match('^([^:]+):(%d+)$')
    return host, tonumber(port)
end

local function probe_addr(addr)
    if not socket_ok then
        return false
    end
    local host, port = parse_addr(addr)
    if not host or not port or not loopback_host(host) then
        return false
    end
    local sock = socket.tcp()
    sock:settimeout(0.3)
    local ok = sock:connect(host, port)
    sock:close()
    return ok == 1
end

local function wait_for_agent()
    for _ = 1, 40 do
        local addr = read_portfile()
        if addr and probe_addr(addr) then
            return addr
        end
        if socket_ok then
            socket.sleep(0.05)
        end
    end
    return nil
end

local function encode_string(value)
    return '"' .. tostring(value):gsub('"', '\\"') .. '"'
end

local function encode_request(args, output)
    local quoted = {}
    for _, a in ipairs(args) do
        quoted[#quoted + 1] = encode_string(a)
    end
    return '{"args":[' .. table.concat(quoted, ',') .. '],"output":' .. encode_string(output) .. '}'
end

local function wants_json(args)
    for i, a in ipairs(args) do
        if a == '--output' and args[i + 1] == 'json' then
            return true
        end
        if a == '--output=json' then
            return true
        end
    end
    return false
end

local function quote_shell(arg)
    arg = tostring(arg or '')
    if arg:find(' ', 1, true) or arg:find('"', 1, true) then
        return '"' .. arg:gsub('"', '\\"') .. '"'
    end
    return arg
end

local function start_agent(bin)
    local portfile = port_path()
    local host_portfile = agent.host_path(portfile)
    local cmd

    if agent.wine_on_linux() then
        local host_bin = agent.host_path(windower.addon_path .. 'bin/macromog-host')
        cmd = 'start /unix ' .. quote_shell(host_bin) .. ' agent --portfile ' .. quote_shell(host_portfile)
        agent.last_backend = 'agent-unix'
    else
        cmd = quote_shell(bin) .. ' agent --portfile ' .. quote_shell(portfile)
        agent.last_backend = 'agent-win'
    end

    log.debug('agent.start: ' .. cmd)
    os.execute(cmd)
    return wait_for_agent()
end

local function ensure_addr(bin)
    local addr = read_portfile()
    if addr and probe_addr(addr) then
        return addr
    end
    return start_agent(bin)
end

function agent.available()
    return socket_ok
end

function agent.run(bin, args, opts)
    if not socket_ok then
        return nil
    end

    local addr = ensure_addr(bin)
    if not addr then
        log.debug('agent.run: no reachable agent')
        return nil
    end

    local host, port = parse_addr(addr)
    if not loopback_host(host) then
        log.debug('agent.run: refusing non-loopback address ' .. tostring(addr))
        return nil
    end
    local sock = socket.tcp()
    sock:settimeout(5)
    if sock:connect(host, port) ~= 1 then
        sock:close()
        return nil
    end

    local output = wants_json(args) and 'json' or 'text'
    local payload = encode_request(args, output)
    sock:send(payload .. '\n')

    local body = sock:receive('*a') or ''
    sock:close()

    local resp = json.decode(body)
    if not resp then
        return nil
    end
    if resp.error and resp.error ~= '' then
        return nil, resp.error
    end

    local out = resp.stdout or ''
    if resp.stderr and resp.stderr ~= '' then
        if out ~= '' then
            out = out .. resp.stderr
        else
            out = resp.stderr
        end
    end

    return tonumber(resp.code) or 1, out, nil
end

function agent.shutdown()
    if not socket_ok then
        return
    end

    local addr = read_portfile()
    if not addr then
        return
    end

    local host, port = parse_addr(addr)
    if not loopback_host(host) then
        log.debug('agent.shutdown: refusing non-loopback address ' .. tostring(addr))
        return
    end

    log.debug('agent.shutdown: ' .. addr)
    local sock = socket.tcp()
    sock:settimeout(1)
    if sock:connect(host, port) == 1 then
        sock:send('{"shutdown":true}\n')
        sock:receive('*a')
        sock:close()
    end

    local path = port_path()
    os.remove(path)
    local host_path = agent.host_path(path)
    if host_path ~= path then
        os.remove(host_path)
    end
end

return agent
