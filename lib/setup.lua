-- Auto-config orchestration and readiness state for the addon.

local cli = require('lib/cli')
local detect = require('lib/detect')
local log = require('lib/log')
local process = require('lib/process')

local setup = {
    install_ready = false,
    zoned_since_load = false,
    noticed_zone = false,
    learned = {},
}

local function logged_in()
    local info = windower.ffxi.get_info()
    return info and info.logged_in
end

local function player_name()
    local player = windower.ffxi.get_player()
    return player and player.name
end

local function any_install(cfg)
    if not cfg or not cfg.config or not cfg.config.installs then
        return false
    end
    for _, inst in pairs(cfg.config.installs) do
        if inst.path and inst.path ~= '' then
            return true
        end
    end
    return false
end

local function alias_exists(cfg, name)
    if not cfg or not cfg.config or not cfg.config.installs then
        return false
    end
    local lower = name:lower()
    for _, inst in pairs(cfg.config.installs) do
        if inst.characters then
            for _, ch in pairs(inst.characters) do
                if ch.name and ch.name:lower() == lower then
                    return true
                end
            end
        end
    end
    return false
end

local function pick_char_id(user_dir, characters)
    if #characters == 1 then
        return characters[1].id
    end
    local best_id, best_stamp
    for _, ch in ipairs(characters) do
        local stamp = process.file_mtime(user_dir .. '\\' .. ch.id .. '\\mcr.dat')
        if stamp and (not best_stamp or stamp > best_stamp) then
            best_stamp = stamp
            best_id = ch.id
        end
    end
    return best_id or characters[1].id
end

function setup.ensure_install()
    local cfg, err = cli.config_show()
    if not cfg then
        log.user('Config check failed: ' .. (err or 'unknown'))
        log.debug('config_show failed: ' .. tostring(err))
        return false
    end
    log.debug('config path: ' .. tostring(cfg.path))
    if any_install(cfg) then
        setup.install_ready = true
        log.debug('install already configured')
        return true
    end

    local root = detect.ffxi_root(cli)
    log.debug('detected ffxi root (wine-native): ' .. tostring(root))
    if not root then
        log.user('Could not detect FFXI install. Run macromog config add-install, kupo!')
        return false
    end

    local name = detect.suggest_install_name(root)
    log.debug('registering install name=' .. name .. ' detected path (wine-native)=' .. root)
    local code, out = cli.config_add_install(name, root)
    if code ~= 0 then
        log.user('Install registration failed: ' .. (out or ''))
        log.debug('add-install failed: ' .. tostring(out))
        return false
    end
    log.debug('add-install stdout: ' .. tostring(out))
    local saved, show_err = cli.config_show()
    if not saved then
        log.user('Install registration could not be verified, kupo!')
        log.debug('post add-install config_show failed: ' .. tostring(show_err))
        return false
    end
    local inst = saved.config and saved.config.installs and saved.config.installs[name]
    if not inst or not inst.path or inst.path == '' then
        log.user('Install registration could not be verified, kupo!')
        log.debug('post add-install verify failed for install ' .. tostring(name))
        return false
    end
    log.debug('stored install path: ' .. tostring(inst.path))
    setup.install_ready = true
    return true
end

function setup.ensure_character(name)
    if not name or setup.learned[name] then
        return true
    end
    local cfg = cli.config_show()
    if not cfg then
        return false
    end
    if alias_exists(cfg, name) then
        setup.learned[name] = true
        return true
    end

    local list_data = cli.list_all()
    if not list_data or not list_data.characters or #list_data.characters == 0 then
        log.user('No character folders found for alias setup, kupo!')
        return false
    end

    local candidates = {}
    for _, ch in ipairs(list_data.characters) do
        if not ch.name or ch.name == '' then
            candidates[#candidates + 1] = ch
        end
    end
    if #candidates == 0 then
        candidates = list_data.characters
    end

    local char_id = pick_char_id(list_data.user_dir, candidates)
    if not char_id then
        log.user('Could not determine character folder for ' .. name .. ', kupo!')
        return false
    end

    local code, out = cli.config_set_alias(char_id, name)
    if code ~= 0 then
        log.user('Alias setup failed: ' .. (out or ''))
        return false
    end
    log.user("Character '" .. name .. "' has been registered with this install, kupo!")
    setup.learned[name] = true
    return true
end

function setup.on_zone()
    setup.zoned_since_load = true
    if not setup.noticed_zone then
        setup.noticed_zone = true
        local name = player_name()
        if name then
            setup.ensure_character(name)
        end
        log.user('Ready! Type //mmog help for commands, kupo!')
    end
end

function setup.on_load()
    log.refresh()
    setup.ensure_install()
    if logged_in() then
        local name = player_name()
        if setup.install_ready and name and setup.ensure_character(name) then
            setup.zoned_since_load = true
            setup.noticed_zone = true
        elseif not setup.zoned_since_load then
            log.user('Zone once before using //mmog commands, kupo!')
        end
    end
end

function setup.on_login()
    setup.zoned_since_load = false
    setup.noticed_zone = false
end

function setup.ready()
    return setup.install_ready and setup.zoned_since_load and logged_in()
end

function setup.ready_message()
    if not setup.install_ready then
        return 'Macromog install is not configured yet, kupo!'
    end
    if not setup.zoned_since_load then
        return 'Zone once before using Macromog commands, kupo!'
    end
    if not logged_in() then
        return 'Log in before using Macromog commands, kupo!'
    end
    return nil
end

return setup
