-- Macromog - FFXI macro manager via YAML files
-- Windower 4 addon

_addon.name = 'Macromog'
_addon.author = 'Squatched'
_addon.version = '0.1.0'
_addon.commands = { 'macromog', 'mmog' }

local cli = require('lib/cli')
local setup = require('lib/setup')

local CHAT_COLOR = 207 -- moogle purple

local function log(msg)
    windower.add_to_chat(CHAT_COLOR, '[Macromog] ' .. tostring(msg))
end

local function usage()
    log('Commands: export | import <file> | validate <file> | backup | help')
end

local function require_ready()
    local msg = setup.ready_message()
    if msg then
        log(msg)
        return false
    end
    return true
end

local function char_name()
    local player = windower.ffxi.get_player()
    return player and player.name
end

local handlers = {}

function handlers.export(filename)
    if not require_ready() then
        return
    end
    local name = char_name()
    if not filename then
        local timestamp = os.date('%Y%m%d_%H%M%S')
        filename = name .. '_macros_' .. timestamp .. '.yml'
    end
    local path = windower.addon_path .. 'data/' .. filename
    local code, out = cli.export(path, name)
    if code ~= 0 then
        log('Export failed: ' .. (out or ''))
        return
    end
    log('Exported to ' .. filename .. ', kupo!')
end

function handlers.import(filename)
    if not require_ready() then
        return
    end
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
    f:close()

    local code, out = cli.import(path, char_name())
    if code ~= 0 then
        log('Import failed: ' .. (out or ''))
        return
    end
    log('Imported ' .. filename .. ' successfully, kupo!')
end

function handlers.validate(filename)
    if not require_ready() then
        return
    end
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
    f:close()

    local code, out = cli.validate(path)
    if code == 0 then
        log('Validation passed, kupo!')
    else
        log('Validation failed: ' .. (out or ''))
    end
end

function handlers.backup()
    if not require_ready() then
        return
    end
    local code, out = cli.backup(char_name())
    if code ~= 0 then
        log('Backup failed: ' .. (out or ''))
        return
    end
    log('Backup complete, kupo!')
end

windower.register_event('addon command', function(cmd, ...)
    local args = { ... }
    cmd = (cmd or ''):lower()

    if cmd == 'export' then
        handlers.export(args[1])
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
    setup.on_load()
    log('Kupomog at your service, kupo! Type //mmog help for commands.')
end)

windower.register_event('login', function()
    setup.on_login()
end)

windower.register_event('incoming chunk', function(id)
    if id == 0x0A and not setup.zoned_since_load then
        setup.on_zone()
    end
end)
