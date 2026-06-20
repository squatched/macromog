-- Macromog - FFXI macro manager via YAML files
-- Windower 4 addon

_addon.name = 'Macromog'
_addon.author = 'Caleb McCombs'
_addon.version = '0.1.0'
_addon.commands = { 'macromog', 'mmog' }

local macros = require('lib/macros')
local yaml = require('lib/yaml')
local validate = require('lib/validate')

local CHAT_COLOR = 207 -- moogle purple

local function log(msg)
    windower.add_to_chat(CHAT_COLOR, '[Macromog] ' .. tostring(msg))
end

local function usage()
    log('Commands: export | import <file> | validate <file> | backup | help')
end

local handlers = {}

function handlers.export()
    local char = windower.ffxi.get_player().name
    local filename = char .. '_macros.yml'
    local path = windower.addon_path .. 'data/' .. filename

    local data = macros.read()
    if not data then
        log('Failed to read macros from memory, kupo!')
        return
    end

    local ok, err = validate.macros(data)
    if not ok then
        log('Validation error: ' .. err)
        return
    end

    local f = io.open(path, 'w')
    if not f then
        log('Could not open for writing: ' .. path)
        return
    end
    f:write(yaml.dump(data))
    f:close()
    log('Exported to ' .. filename .. ', kupo!')
end

function handlers.import(filename)
    if not filename then
        log('Usage: //mmog import <filename>')
        return
    end

    local path = windower.addon_path .. 'data/' .. filename
    local f = io.open(path, 'r')
    if not f then
        log('File not found: ' .. filename)
        return
    end
    local content = f:read('*a')
    f:close()

    local data, err = yaml.parse(content)
    if not data then
        log('YAML parse error: ' .. (err or 'unknown'))
        return
    end

    local ok, verr = validate.macros(data)
    if not ok then
        log('Validation failed: ' .. verr)
        return
    end

    if not macros.backup() then
        log('Backup failed — aborting import for safety, kupo!')
        return
    end

    macros.write(data)
    log('Imported ' .. filename .. ' successfully, kupo!')
end

function handlers.validate(filename)
    if not filename then
        log('Usage: //mmog validate <filename>')
        return
    end

    local path = windower.addon_path .. 'data/' .. filename
    local f = io.open(path, 'r')
    if not f then
        log('File not found: ' .. filename)
        return
    end
    local content = f:read('*a')
    f:close()

    local data, err = yaml.parse(content)
    if not data then
        log('YAML parse error: ' .. (err or 'unknown'))
        return
    end

    local ok, verr = validate.macros(data)
    if ok then
        log('Validation passed, kupo!')
    else
        log('Validation failed: ' .. verr)
    end
end

function handlers.backup()
    if macros.backup() then
        log('Backup complete, kupo!')
    end
end

windower.register_event('addon command', function(cmd, ...)
    local args = { ... }
    cmd = (cmd or ''):lower()

    if cmd == 'export' then
        handlers.export()
    elseif cmd == 'import' then
        handlers.import(args[1])
    elseif cmd == 'validate' then
        handlers.validate(args[1])
    elseif cmd == 'backup' then
        handlers.backup()
    elseif cmd == 'help' or cmd == '' then
        usage()
    else
        log('Unknown command: ' .. cmd)
        usage()
    end
end)

windower.register_event('load', function()
    log('Kupomog at your service, kupo! Type //mmog help for commands.')
end)
