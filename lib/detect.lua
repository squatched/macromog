-- Detect the active FFXI install root for auto-config.

local detect = {}

local function normalize_slashes(path)
    return (path:gsub('\\', '/'))
end

local function dirname(path)
    path = normalize_slashes(path)
    return path:match('^(.*)/[^/]+$') or path
end

local function has_user_dir(root)
    if not root or root == '' then
        return false
    end
    local user = normalize_slashes(root) .. '/USER'
    return windower.dir_exists(user) or windower.dir_exists(user:gsub('/', '\\'))
end

local function extract_path_from_xml(content)
    local path = content:match('([A-Za-z]:[\\/].-FINAL FANTASY XI[^<]*)')
    if path then
        path = path:gsub('^%s+', ''):gsub('%s+$', '')
        return normalize_slashes(path)
    end
    return nil
end

local function read_file(path)
    local f = io.open(path, 'r')
    if not f then
        return nil
    end
    local content = f:read('*a')
    f:close()
    return content
end

function detect.profile_candidates()
    local out = {}
    local windower_path = windower.windower_path or ''
    if windower_path ~= '' then
        out[#out + 1] = windower_path .. 'settings.xml'
        out[#out + 1] = windower_path .. 'profiles/settings.xml'
    end
    local appdata = os.getenv('APPDATA')
    if appdata and appdata ~= '' then
        out[#out + 1] = appdata .. '\\Windower\\settings.xml'
    end
    return out
end

function detect.from_windower_profile()
    for _, path in ipairs(detect.profile_candidates()) do
        local content = read_file(path)
        if content then
            local game = extract_path_from_xml(content)
            if game then
                local root = game
                if game:lower():find('pol.exe', 1, true) then
                    root = dirname(game)
                end
                if has_user_dir(root) then
                    return root
                end
            end
        end
    end
    return nil
end

function detect.from_list_json(list_data)
    if not list_data or not list_data.user_dir then
        return nil
    end
    local user = normalize_slashes(list_data.user_dir)
    return dirname(user)
end

function detect.ffxi_root(cli)
    local root = detect.from_windower_profile()
    if root then
        return root
    end
    local list_data = cli.list_all()
    if list_data then
        return detect.from_list_json(list_data)
    end
    return nil
end

function detect.run_diag(log, cli_mod)
    log.debug('windower_path: ' .. tostring(windower.windower_path or ''))
    log.debug('addon_path: ' .. tostring(windower.addon_path or ''))
    log.debug('APPDATA: ' .. tostring(os.getenv('APPDATA') or ''))
    log.debug('USERPROFILE: ' .. tostring(os.getenv('USERPROFILE') or ''))
    log.debug('WINEPREFIX: ' .. tostring(os.getenv('WINEPREFIX') or ''))
    log.debug('MACROMOG_CONFIG: ' .. tostring(os.getenv('MACROMOG_CONFIG') or ''))

    for _, path in ipairs(detect.profile_candidates()) do
        local content = read_file(path)
        if content then
            local game = extract_path_from_xml(content)
            log.debug('profile ' .. path .. ' -> ' .. tostring(game))
            if game then
                log.debug('  has_user_dir=' .. tostring(has_user_dir(game)))
            end
        else
            log.debug('profile missing: ' .. path)
        end
    end

    local profile_root = detect.from_windower_profile()
    log.debug('from_windower_profile: ' .. tostring(profile_root))

    local list_data = cli_mod.list_all()
    if list_data then
        log.debug('list_all user_dir: ' .. tostring(list_data.user_dir))
        log.debug('list_all characters: ' .. tostring(#(list_data.characters or {})))
    else
        log.debug('list_all: failed')
    end

    log.debug('ffxi_root: ' .. tostring(detect.ffxi_root(cli_mod)))

    local code, out = cli_mod.debug_all()
    log.debug('cli debug all exit=' .. tostring(code))
    if out and out ~= '' then
        for line in out:gmatch('[^\r\n]+') do
            log.debug('cli: ' .. line)
        end
    end
end

function detect.suggest_install_name(path)
    local lower = path:lower()
    if lower:find('steam', 1, true) then
        return 'steam'
    end
    if lower:find('lutris', 1, true) or lower:find('final-fantasy', 1, true) then
        return 'lutris'
    end
    if lower:find('wine', 1, true) or lower:find('%.wine', 1, true) then
        return 'wine'
    end
    if lower:find('playonline', 1, true) then
        return 'playonline'
    end
    return 'default'
end

return detect
