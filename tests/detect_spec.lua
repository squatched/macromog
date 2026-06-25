package.path = './?.lua;./?/init.lua;' .. package.path

_G.windower = {
    windower_path = 'C:/Windower4/',
    dir_exists = function(path)
        return path:find('FINAL FANTASY XI/USER', 1, true) ~= nil
            or path:find('FINAL FANTASY XI\\USER', 1, true) ~= nil
            or path:find('GoodRoot/USER', 1, true) ~= nil
    end,
}

local detect = require('lib/detect')

describe('detect.from_list_json', function()
    it('derives ffxi root from user_dir', function()
        local root = detect.from_list_json({
            user_dir = 'C:/Games/FFXI/USER',
        })
        assert.are.equal('C:/Games/FFXI', root)
    end)

    it('returns nil without user_dir', function()
        assert.is_nil(detect.from_list_json({}))
        assert.is_nil(detect.from_list_json(nil))
    end)
end)

describe('detect.from_windower_profile', function()
    local orig_open = io.open

    after_each(function()
        io.open = orig_open
    end)

    local function mock_settings(content)
        io.open = function(path)
            if path:find('settings.xml', 1, true) then
                return {
                    read = function()
                        return content
                    end,
                    close = function() end,
                }
            end
            return nil
        end
    end

    it('reads game path from settings.xml', function()
        mock_settings('<entry>C:\\Games\\FINAL FANTASY XI\\pol.exe</entry>')
        local root = detect.from_windower_profile()
        assert.are.equal('C:/Games/FINAL FANTASY XI', root)
    end)

    it('returns nil when no profile files exist', function()
        io.open = function()
            return nil
        end
        assert.is_nil(detect.from_windower_profile())
    end)

    local negative_cases = {
        {
            name = 'xml without ffxi path',
            content = '<entry>C:\\Games\\OtherGame\\game.exe</entry>',
        },
        {
            name = 'empty xml',
            content = '',
        },
        {
            name = 'malformed xml',
            content = 'not xml at all',
        },
    }

    for _, case in ipairs(negative_cases) do
        it('returns nil for ' .. case.name, function()
            mock_settings(case.content)
            assert.is_nil(detect.from_windower_profile())
        end)
    end

    it('returns nil when ffxi path has no USER directory', function()
        local orig_dir_exists = windower.dir_exists
        windower.dir_exists = function()
            return false
        end
        mock_settings('<entry>C:\\Games\\FINAL FANTASY XI\\pol.exe</entry>')
        assert.is_nil(detect.from_windower_profile())
        windower.dir_exists = orig_dir_exists
    end)

    it('accepts direct install root without pol.exe', function()
        mock_settings('<entry>C:\\Games\\GoodRoot\\FINAL FANTASY XI</entry>')
        local root = detect.from_windower_profile()
        assert.are.equal('C:/Games/GoodRoot/FINAL FANTASY XI', root)
    end)
end)

describe('detect.profile_candidates', function()
    local orig_getenv = os.getenv

    after_each(function()
        os.getenv = orig_getenv
    end)

    it('includes windower and appdata paths', function()
        os.getenv = function(name)
            if name == 'APPDATA' then
                return 'C:/Users/Adventurer/AppData/Roaming'
            end
            return orig_getenv(name)
        end
        local candidates = detect.profile_candidates()
        assert.is_true(candidates[1]:find('C:/Windower4/settings.xml', 1, true) ~= nil)
        local last = candidates[#candidates]:lower()
        assert.is_true(last:find('windower', 1, true) ~= nil)
        assert.is_true(last:find('settings.xml', 1, true) ~= nil)
    end)
end)

describe('detect.suggest_install_name', function()
    local cases = {
        { path = '/home/.steam/steamapps/FFXI', want = 'steam' },
        { path = '/games/lutris/ffxi', want = 'lutris' },
        {
            path = '/home/squatched/Games/final-fantasy-xi-online/drive_c/FFXI',
            want = 'lutris',
        },
        { path = '/home/user/.wine/drive_c/ffxi', want = 'wine' },
        { path = 'C:/PlayOnline/SquareEnix', want = 'playonline' },
        { path = 'C:/Games/FINAL FANTASY XI', want = 'default' },
    }

    for _, case in ipairs(cases) do
        it('detects ' .. case.want .. ' from ' .. case.path, function()
            assert.are.equal(case.want, detect.suggest_install_name(case.path))
        end)
    end
end)

describe('detect.ffxi_root', function()
    it('falls back to list json', function()
        local cli = {
            list_all = function()
                return { user_dir = 'D:/ffxi/USER' }
            end,
        }
        assert.are.equal('D:/ffxi', detect.ffxi_root(cli))
    end)

    it('prefers windower profile over list json', function()
        local orig_open = io.open
        io.open = function(path)
            if path:find('settings.xml', 1, true) then
                return {
                    read = function()
                        return '<entry>C:\\Games\\GoodRoot\\FINAL FANTASY XI</entry>'
                    end,
                    close = function() end,
                }
            end
            return nil
        end
        local cli = {
            list_all = function()
                return { user_dir = 'D:/ffxi/USER' }
            end,
        }
        assert.are.equal('C:/Games/GoodRoot/FINAL FANTASY XI', detect.ffxi_root(cli))
        io.open = orig_open
    end)

    it('returns nil when profile and list both fail', function()
        local orig_open = io.open
        io.open = function()
            return nil
        end
        local cli = {
            list_all = function()
                return nil
            end,
        }
        assert.is_nil(detect.ffxi_root(cli))
        io.open = orig_open
    end)
end)

describe('detect.run_diag', function()
    local orig_open = io.open
    local orig_getenv = os.getenv
    local lines

    local mock_log = {
        debug = function(msg)
            lines[#lines + 1] = msg
        end,
    }

    before_each(function()
        lines = {}
    end)

    after_each(function()
        io.open = orig_open
        os.getenv = orig_getenv
    end)

    it('logs profile misses and list_all success', function()
        io.open = function()
            return nil
        end
        os.getenv = function(name)
            if name == 'WINEPREFIX' then
                return '/home/adventurer/Games/ffxi'
            end
            return orig_getenv(name)
        end
        local cli_mod = {
            list_all = function()
                return {
                    user_dir = 'C:/ffxi/USER',
                    characters = { { id = 'abc' } },
                }
            end,
            debug_all = function()
                return 0, 'paths:\nlinux_home: /home/adventurer\n'
            end,
        }
        detect.run_diag(mock_log, cli_mod)
        assert.is_true(#lines > 0)
        assert.is_true(table.concat(lines, '\n'):find('profile missing', 1, true) ~= nil)
        assert.is_true(table.concat(lines, '\n'):find('list_all user_dir', 1, true) ~= nil)
        assert.is_true(table.concat(lines, '\n'):find('cli debug all exit=0', 1, true) ~= nil)
    end)

    it('logs list_all failure', function()
        io.open = function()
            return nil
        end
        local cli_mod = {
            list_all = function()
                return nil
            end,
            debug_all = function()
                return 1, 'debug failed'
            end,
        }
        detect.run_diag(mock_log, cli_mod)
        local joined = table.concat(lines, '\n')
        assert.is_true(joined:find('list_all: failed', 1, true) ~= nil)
        assert.is_true(joined:find('cli debug all exit=1', 1, true) ~= nil)
    end)
end)