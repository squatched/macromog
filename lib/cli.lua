-- Invoke the bundled macromog CLI binary.

local json = require('lib/json')
local log = require('lib/log')
local process = require('lib/process')

local cli = {}

function cli.binary_path()
    return windower.addon_path .. 'bin/macromog.exe'
end

function cli.run(args, opts)
    opts = opts or {}
    local bin = cli.binary_path()
    -- MACROMOG_DEBUG writes to stderr; we merge stderr into stdout (2>&1), so
    -- only enable it for explicit debug probes — never for --output json calls.
    log.debug('cli.run: ' .. bin .. ' ' .. table.concat(args, ' '))
    local pipe = process.popen(bin, args, opts)
    log.debug('cli.run backend: ' .. tostring(process.last_backend))
    if not pipe then
        return 1, '', 'failed to run CLI'
    end
    local out = pipe:read('*a') or ''
    local ok, how, code = pipe:close()
    if not ok then
        if type(code) == 'number' then
            log.debug('cli.run exit=' .. tostring(code))
            return code, out, nil
        end
        return 1, out, how or 'command failed'
    end
    log.debug('cli.run exit=0')
    return 0, out, nil
end

function cli.run_json(args, opts)
    local full = { '--output', 'json' }
    for _, a in ipairs(args) do
        full[#full + 1] = a
    end
    local code, out, err = cli.run(full, opts)
    if code ~= 0 then
        return nil, out ~= '' and out or err
    end
    local data, perr = json.decode(out)
    if not data then
        return nil, perr or 'invalid json'
    end
    return data
end

function cli.debug_all()
    return cli.run({ 'debug', 'all' }, { debug = true })
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
