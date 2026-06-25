-- Auto-config orchestration and readiness state for the addon.

local cli = require('lib/cli')
local detect = require('lib/detect')

local setup = {
    install_ready = false,
    zoned_since_load = false,
    noticed_zone = false,
    learned = {},
}

local CHAT_COLOR = 207

local function log(msg)
    windower.add_to_chat(CHAT_COLOR, '[Macromog] ' .. tostring(msg))
end

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

local function file_mtime(path)
    local handle = io.popen('cmd /c for %I in ("' .. path:gsub('"', '') .. '") do @echo %~tI')
    if not handle then
        return nil
    end
    local line = handle:read('*l')
    handle:close()
    return line
end

local function pick_char_id(user_dir, characters)
    if #characters == 1 then
        return characters[1].id
    end
    local best_id, best_stamp
    for _, ch in ipairs(characters) do
        local stamp = file_mtime(user_dir .. '\\' .. ch.id .. '\\mcr.dat')
        if stamp and (not best_stamp or stamp > best_stamp) then
            best_stamp = stamp
            best_id = ch.id
        end
    end
    return best_id
end

function setup.ensure_install()
    local cfg, err = cli.config_show()
    if not cfg then
        log('Config check failed: ' .. (err or 'unknown'))
        return false
    end
    if any_install(cfg) then
        setup.install_ready = true
        return true
    end

    local root = detect.ffxi_root(cli)
    if not root then
        log('Could not detect FFXI install. Run macromog config add-install, kupo!')
        return false
    end

    local name = detect.suggest_install_name(root)
    local code, out = cli.config_add_install(name, root)
    if code ~= 0 then
        log('Install registration failed: ' .. (out or ''))
        return false
    end
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
        log('No character folders found for alias setup, kupo!')
        return false
    end

    local char_id = pick_char_id(list_data.user_dir, list_data.characters)
    if not char_id then
        log('Could not determine character folder for ' .. name .. ', kupo!')
        return false
    end

    local code, out = cli.config_set_alias(char_id, name)
    if code ~= 0 then
        log('Alias setup failed: ' .. (out or ''))
        return false
    end
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
        log('Ready! Type //mmog help for commands, kupo!')
    end
end

function setup.on_load()
    setup.ensure_install()
    if logged_in() and not setup.zoned_since_load then
        log('Zone once before using //mmog commands, kupo!')
    end
end

function setup.on_login()
    -- Character learning happens on the first zone after login.
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
