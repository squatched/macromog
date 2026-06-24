-- Invoke the bundled macromog CLI binary.

local json = require('lib/json')

local cli = {}

local function arch_suffix()
    if jit and jit.arch == 'x64' then
        return 'amd64'
    end
    return '386'
end

function cli.binary_path()
    return windower.addon_path .. 'bin/macromog-windows-' .. arch_suffix() .. '.exe'
end

local function quote(arg)
    arg = tostring(arg or '')
    if arg:find(' ', 1, true) or arg:find('"', 1, true) then
        return '"' .. arg:gsub('"', '\\"') .. '"'
    end
    return arg
end

function cli.run(args)
    local bin = cli.binary_path()
    local parts = { quote(bin) }
    for _, a in ipairs(args) do
        parts[#parts + 1] = quote(a)
    end
    local cmd = table.concat(parts, ' ') .. ' 2>&1'
    local pipe = io.popen(cmd, 'r')
    if not pipe then
        return 1, '', 'failed to run CLI'
    end
    local out = pipe:read('*a') or ''
    local ok, how, code = pipe:close()
    if not ok then
        if type(code) == 'number' then
            return code, out, nil
        end
        return 1, out, how or 'command failed'
    end
    return 0, out, nil
end

function cli.run_json(args)
    local full = {}
    for _, a in ipairs(args) do
        full[#full + 1] = a
    end
    full[#full + 1] = '--output'
    full[#full + 1] = 'json'
    local code, out, err = cli.run(full)
    if code ~= 0 then
        return nil, out ~= '' and out or err
    end
    local data, perr = json.decode(out)
    if not data then
        return nil, perr or 'invalid json'
    end
    return data
end

function cli.config_show()
    return cli.run_json({ 'config', 'show' })
end

function cli.config_add_install(name, path)
    return cli.run({ 'config', 'add-install', name, path })
end

function cli.config_set_alias(char_id, name, extra)
    local args = { 'config', 'set-alias', char_id, name }
    if extra then
        for _, a in ipairs(extra) do
            args[#args + 1] = a
        end
    end
    return cli.run(args)
end

function cli.list_all()
    return cli.run_json({ 'list' })
end

function cli.export(output, char_name)
    return cli.run({ 'export', '--char-name', char_name, '-o', output })
end

function cli.import(yaml_path, char_name)
    return cli.run({ 'import', yaml_path, '--char-name', char_name })
end

function cli.validate(yaml_path)
    return cli.run({ 'validate', yaml_path })
end

function cli.backup(char_name)
    return cli.run({ 'backup', '--char-name', char_name, '--in-place' })
end

return cli
