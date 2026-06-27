-- Macromog - FFXI macro manager via YAML files
-- Windower 4 addon

_addon.name = 'Macromog'
_addon.author = 'Squatched'
_addon.version = '1.0.1' -- x-release-please-version
_addon.commands = { 'macromog', 'mmog' }

local cli = require('lib/cli')
local detect = require('lib/detect')
local log = require('lib/log')
local setup = require('lib/setup')

local function usage()
    log.user('Commands: export | import <file> | validate <file> | backup | debug | diag | help')
    log.user('File paths for import/validate are relative to the addon data/ folder, kupo!')
end

local function require_ready()
    local msg = setup.ready_message()
    if msg then
        log.user(msg)
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
        log.user('Export failed: ' .. (out or ''))
        return
    end
    log.user('Exported to ' .. filename .. ', kupo!')
end

function handlers.import(filename)
    if not require_ready() then
        return
    end
    if not filename then
        log.user('Usage: //mmog import <filename>  (relative to addon data/ folder)')
        return
    end

    local path = windower.addon_path .. 'data/' .. filename
    local f = io.open(path, 'r')
    if not f then
        log.user('File not found: ' .. filename .. '  (paths are relative to the addon data/ folder)')
        return
    end
    f:close()

    local code, out = cli.validate(path)
    if code ~= 0 then
        log.user('Import validation failed: ' .. (out or ''))
        return
    end

    local name = char_name()
    local data_dir = windower.addon_path .. 'data'
    setup.pending_import = { path = path, name = name, backup_dir = data_dir }
    log.user('Staged ' .. filename .. '. Zone once to apply in-game, kupo!')
end

function handlers.validate(filename)
    if not require_ready() then
        return
    end
    if not filename then
        log.user('Usage: //mmog validate <filename>  (relative to addon data/ folder)')
        return
    end

    local path = windower.addon_path .. 'data/' .. filename
    local f = io.open(path, 'r')
    if not f then
        log.user('File not found: ' .. filename .. '  (paths are relative to the addon data/ folder)')
        return
    end
    f:close()

    local code, out = cli.validate(path)
    if code == 0 then
        log.user('Validation passed, kupo!')
    else
        log.user('Validation failed: ' .. (out or ''))
    end
end

function handlers.backup()
    if not require_ready() then
        return
    end
    local code, out = cli.backup(char_name(), windower.addon_path .. 'data')
    if code ~= 0 then
        log.user('Backup failed: ' .. (out or ''))
        return
    end
    local msg = (out or ''):match('^(.-)%s*$')
    log.user(msg ~= '' and msg or 'Backup complete, kupo!')
end

function handlers.debug(mode)
    mode = (mode or ''):lower()
    if mode == 'on' then
        log.enable_persist()
        log.user('Debug logging enabled (data/.debug). Log: data/debug.log')
        return
    end
    if mode == 'off' then
        log.disable_persist()
        log.end_session()
        log.user('Debug logging disabled, kupo!')
        return
    end
    log.user('Usage: //mmog debug on | off')
end

function handlers.diag()
    log.begin_session()
    log.reset_log()
    detect.run_diag(log, cli)
    log.end_session()
    log.user('Wrote diag log to data/debug.log (and Windower console if debug is on)')
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
    elseif cmd == 'debug' then
        handlers.debug(args[1])
    elseif cmd == 'diag' then
        handlers.diag()
    elseif cmd == 'help' or cmd == '' then
        usage()
    else
        log.user('Unknown command: ' .. cmd)
        usage()
    end
end)

windower.register_event('load', function()
    log.refresh()
    setup.on_load()
    log.user('Kupomog at your service, kupo! Type //mmog help for commands.')
end)

windower.register_event('login', function(name)
    setup.on_login(name)
end)

windower.register_event('incoming chunk', function(id)
    if id == 0x0A then
        if not setup.zoned_since_load then
            setup.on_zone()
        end
        if setup.pending_import then
            local pi = setup.pending_import
            setup.pending_import = nil
            local code, out = cli.import(pi.path, pi.name, pi.backup_dir)
            if code == 0 then
                local backup_path = (out or ''):match('backed up to ([^\n\r]+)')
                if backup_path then
                    local backup_name = backup_path:match('[^\\/]+$') or backup_path
                    log.user('Macros successfully applied, kupo! (Pre-import backup at ' .. backup_name .. ')')
                else
                    log.user('Macros successfully applied, kupo!')
                end
            else
                log.user('Zone-in macro apply failed: ' .. (out or '') .. '. Re-run //mmog import, kupo!')
            end
        end
    end
end)
