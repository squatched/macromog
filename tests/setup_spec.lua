package.path = './?.lua;./?/init.lua;' .. package.path

_G.jit = { arch = 'x64' }
_G.windower = {
    addon_path = '/tmp/macromog_test/',
    add_to_chat = function() end,
    windower_path = 'C:/Windower4/',
    file_exists = function()
        return false
    end,
    dir_exists = function()
        return false
    end,
    ffxi = {
        get_info = function()
            return { logged_in = true }
        end,
        get_player = function()
            return { name = 'Squatched' }
        end,
    },
}

local cli_calls = {}

package.loaded['lib/cli'] = {
    config_show = function()
        return cli_calls.config_show
    end,
    config_add_install = function(name, path)
        cli_calls.add_install = { name = name, path = path }
        return cli_calls.add_install_code or 0, cli_calls.add_install_out or ''
    end,
    config_set_alias = function(char_id, name)
        cli_calls.set_alias = { char_id = char_id, name = name }
        return cli_calls.set_alias_code or 0, cli_calls.set_alias_out or ''
    end,
    list_all = function()
        return cli_calls.list_all
    end,
}

package.loaded['lib/detect'] = {
    ffxi_root = function()
        return cli_calls.ffxi_root
    end,
    suggest_install_name = function()
        return 'steam'
    end,
}

package.loaded['lib/log'] = nil
package.loaded['lib/setup'] = nil
local setup = require('lib/setup')

describe('setup readiness', function()
    before_each(function()
        setup.install_ready = false
        setup.zoned_since_load = false
        setup.noticed_zone = false
        setup.learned = {}
        cli_calls = {
            config_show = { config = { installs = { steam = { path = '/ffxi' } } } },
            list_all = {
                user_dir = 'C:/ffxi/USER',
                characters = { { id = 'a1b2c3d4' } },
            },
            ffxi_root = 'C:/ffxi',
        }
    end)

    it('blocks commands before zone', function()
        setup.install_ready = true
        assert.is_not_nil(setup.ready_message())
        assert.is_false(setup.ready())
    end)

    it('allows commands after zone', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        assert.is_nil(setup.ready_message())
        assert.is_true(setup.ready())
    end)

    it('registers install when config is empty', function()
        cli_calls.config_show = { config = {} }
        assert.is_true(setup.ensure_install())
        assert.is_true(setup.install_ready)
        assert.are.equal('steam', cli_calls.add_install.name)
    end)

    it('ignores installs without a stored path', function()
        cli_calls.config_show = { config = { installs = { lutris = { path = '' } } } }
        assert.is_true(setup.ensure_install())
        assert.is_true(setup.install_ready)
        assert.are.equal('steam', cli_calls.add_install.name)
    end)

    it('learns character alias after zone', function()
        setup.on_zone()
        assert.is_true(setup.learned.Squatched)
        assert.are.equal('a1b2c3d4', cli_calls.set_alias.char_id)
    end)

    it('reports install not configured', function()
        assert.are.equal('Macromog install is not configured yet, kupo!', setup.ready_message())
    end)

    it('reports login required', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        windower.ffxi.get_info = function()
            return { logged_in = false }
        end
        assert.are.equal('Log in before using Macromog commands, kupo!', setup.ready_message())
    end)

    it('skips registration when install has a path', function()
        cli_calls.config_show = { config = { installs = { steam = { path = 'C:/ffxi' } } } }
        assert.is_true(setup.ensure_install())
        assert.is_true(setup.install_ready)
        assert.is_nil(cli_calls.add_install)
    end)

    it('short-circuits learned characters', function()
        setup.learned.Squatched = true
        assert.is_true(setup.ensure_character('Squatched'))
        assert.is_nil(cli_calls.set_alias)
    end)

    it('on_load warns when logged in without zone', function()
        windower.ffxi.get_info = function()
            return { logged_in = true }
        end
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        setup.on_load()
        assert.is_true(msgs[#msgs]:find('Zone once', 1, true) ~= nil)
    end)

    it('fails ensure_install when config_show errors', function()
        cli_calls.config_show = nil
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_install())
        assert.is_true(msgs[#msgs]:find('Config check failed', 1, true) ~= nil)
    end)

    it('fails ensure_install when add-install fails', function()
        cli_calls.config_show = { config = {} }
        cli_calls.ffxi_root = 'C:/ffxi'
        cli_calls.add_install_code = 1
        cli_calls.add_install_out = 'duplicate install'
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_install())
        assert.is_true(msgs[#msgs]:find('registration failed', 1, true) ~= nil)
    end)

    it('fails ensure_install when ffxi root is unknown', function()
        cli_calls.config_show = { config = {} }
        cli_calls.ffxi_root = nil
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_install())
        assert.is_true(msgs[#msgs]:find('Could not detect FFXI', 1, true) ~= nil)
    end)

    it('fails ensure_character when list_all has no characters', function()
        cli_calls.list_all = { user_dir = 'C:/ffxi/USER', characters = {} }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_character('Squatched'))
        assert.is_true(msgs[#msgs]:find('No character folders', 1, true) ~= nil)
    end)

    it('skips ensure_character when alias already exists', function()
        cli_calls.config_show = {
            config = {
                installs = {
                    steam = {
                        path = 'C:/ffxi',
                        characters = { a1b2c3d4 = { name = 'Squatched' } },
                    },
                },
            },
        }
        assert.is_true(setup.ensure_character('Squatched'))
        assert.is_true(setup.learned.Squatched)
        assert.is_nil(cli_calls.set_alias)
    end)
end)