package.path = './?.lua;./?/init.lua;' .. package.path

_G.windower = {
    windower_path = 'C:/Windower4/',
    dir_exists = function(path)
        return path:find('FINAL FANTASY XI/USER', 1, true) ~= nil
            or path:find('FINAL FANTASY XI\\USER', 1, true) ~= nil
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

    it('reads game path from settings.xml', function()
        io.open = function(path)
            if path:find('settings.xml', 1, true) then
                return {
                    read = function()
                        return '<entry>C:\\Games\\FINAL FANTASY XI\\pol.exe</entry>'
                    end,
                    close = function() end,
                }
            end
            return nil
        end
        local root = detect.from_windower_profile()
        assert.are.equal('C:/Games/FINAL FANTASY XI', root)
    end)
end)

describe('detect.suggest_install_name', function()
    it('detects steam installs', function()
        assert.are.equal('steam', detect.suggest_install_name('/home/.steam/steamapps/FFXI'))
    end)

    it('detects lutris installs', function()
        assert.are.equal('lutris', detect.suggest_install_name('/games/lutris/ffxi'))
    end)

    it('detects lutris game folder names', function()
        assert.are.equal(
            'lutris',
            detect.suggest_install_name('/home/squatched/Games/final-fantasy-xi-online/drive_c/FFXI')
        )
    end)

    it('detects wine installs', function()
        assert.are.equal('wine', detect.suggest_install_name('/home/user/.wine/drive_c/ffxi'))
    end)

    it('detects playonline installs', function()
        assert.are.equal('playonline', detect.suggest_install_name('C:/PlayOnline/SquareEnix'))
    end)

    it('falls back to default', function()
        assert.are.equal('default', detect.suggest_install_name('C:/Games/FINAL FANTASY XI'))
    end)
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
end)