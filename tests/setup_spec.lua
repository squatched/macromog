package.path = './?.lua;./?/init.lua;' .. package.path

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
        cli_calls.config_show_calls = (cli_calls.config_show_calls or 0) + 1
        if cli_calls.config_show_sequence then
            return cli_calls.config_show_sequence[cli_calls.config_show_calls]
        end
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
local process_mod = require('lib/process')
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
        cli_calls.config_show_sequence = {
            { config = {} },
            { config = { installs = { steam = { path = '/ffxi' } } } },
        }
        assert.is_true(setup.ensure_install())
        assert.is_true(setup.install_ready)
        assert.are.equal('steam', cli_calls.add_install.name)
    end)

    it('sends wine-native path to cli and accepts host-stored verify', function()
        cli_calls.ffxi_root = 'C:/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI'
        cli_calls.config_show_sequence = {
            { config = {} },
            {
                config = {
                    installs = {
                        steam = {
                            path = '/home/adventurer/Games/ffxi/drive_c/Program Files (x86)/FINAL FANTASY XI',
                        },
                    },
                },
            },
        }
        assert.is_true(setup.ensure_install())
        assert.are.equal('steam', cli_calls.add_install.name)
        assert.are.equal(
            'C:/Program Files (x86)/PlayOnline/SquareEnix/FINAL FANTASY XI',
            cli_calls.add_install.path
        )
        assert.is_nil(cli_calls.add_install.path:find('^/home/'))
    end)

    it('ignores installs without a stored path', function()
        cli_calls.config_show_sequence = {
            { config = { installs = { lutris = { path = '' } } } },
            { config = { installs = { steam = { path = '/ffxi' } } } },
        }
        assert.is_true(setup.ensure_install())
        assert.is_true(setup.install_ready)
        assert.are.equal('steam', cli_calls.add_install.name)
    end)

    it('fails ensure_install when post-add config_show cannot verify', function()
        cli_calls.config_show_sequence = {
            { config = {} },
            { config = {} },
        }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_install())
        assert.is_true(msgs[#msgs]:find('could not be verified', 1, true) ~= nil)
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

    it('on_load warns when player name unavailable at load', function()
        windower.ffxi.get_info = function()
            return { logged_in = true }
        end
        local orig_get_player = windower.ffxi.get_player
        windower.ffxi.get_player = function()
            return nil
        end
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        setup.on_load()
        windower.ffxi.get_player = orig_get_player
        assert.is_true(msgs[#msgs]:find('Zone once', 1, true) ~= nil)
    end)

    it('on_load skips zone requirement when character is already registered', function()
        windower.ffxi.get_info = function()
            return { logged_in = true }
        end
        cli_calls.config_show = {
            config = {
                installs = {
                    steam = {
                        path = '/ffxi',
                        characters = { a1b2c3d4 = { name = 'Squatched' } },
                    },
                },
            },
        }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        setup.on_load()
        assert.is_true(setup.zoned_since_load)
        for _, msg in ipairs(msgs) do
            assert.is_false(msg:find('Zone once', 1, true) ~= nil)
        end
    end)

    it('on_load eagerly registers unregistered character and skips zone requirement', function()
        windower.ffxi.get_info = function()
            return { logged_in = true }
        end
        cli_calls.config_show = { config = { installs = { steam = { path = '/ffxi' } } } }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        setup.on_load()
        assert.is_true(setup.zoned_since_load)
        assert.are.equal('a1b2c3d4', cli_calls.set_alias.char_id)
        for _, msg in ipairs(msgs) do
            assert.is_false(msg:find('Zone once', 1, true) ~= nil)
        end
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
        cli_calls.config_show_sequence = {
            { config = {} },
        }
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

    it('falls back to first character when mtime is unavailable', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'aaa' },
                { id = 'bbb' },
            },
        }
        local orig_mtime = process_mod.file_mtime
        process_mod.file_mtime = function()
            return nil
        end
        assert.is_true(setup.ensure_character('Squatched'))
        assert.are.equal('aaa', cli_calls.set_alias.char_id)
        process_mod.file_mtime = orig_mtime
    end)

    it('picks newest mtime among multiple characters', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'older' },
                { id = 'newer' },
            },
        }
        local orig_mtime = process_mod.file_mtime
        process_mod.file_mtime = function(path)
            if path:find('older', 1, true) then
                return '20200101120000'
            end
            return '20250202120000'
        end
        setup.ensure_character('Squatched')
        process_mod.file_mtime = orig_mtime
        assert.are.equal('newer', cli_calls.set_alias.char_id)
    end)

    it('on_login resets zone state so next character is learned', function()
        setup.zoned_since_load = true
        setup.noticed_zone = true
        setup.on_login()
        assert.is_false(setup.zoned_since_load)
        assert.is_false(setup.noticed_zone)
    end)

    it('skips already-aliased characters when picking folder', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'char1', name = 'Char1' },
                { id = 'char2' },
            },
        }
        local orig_mtime = process_mod.file_mtime
        process_mod.file_mtime = function()
            return nil
        end
        assert.is_true(setup.ensure_character('Char2'))
        assert.are.equal('char2', cli_calls.set_alias.char_id)
        process_mod.file_mtime = orig_mtime
    end)

    it('learns new character alias after login resets state', function()
        setup.on_zone()
        assert.is_true(setup.learned.Squatched)

        local orig_get_player = windower.ffxi.get_player
        windower.ffxi.get_player = function()
            return { name = 'Altchar' }
        end
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
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'a1b2c3d4', name = 'Squatched' },
                { id = 'b2c3d4e5' },
            },
        }
        local orig_mtime = process_mod.file_mtime
        process_mod.file_mtime = function()
            return nil
        end

        setup.on_login()
        setup.on_zone()

        windower.ffxi.get_player = orig_get_player
        process_mod.file_mtime = orig_mtime

        assert.is_true(setup.learned.Altchar)
        assert.are.equal('b2c3d4e5', cli_calls.set_alias.char_id)
        assert.are.equal('Altchar', cli_calls.set_alias.name)
    end)

    it('reports alias setup failure', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.set_alias_code = 1
        cli_calls.set_alias_out = 'bad id'
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_character('Squatched'))
        assert.is_true(msgs[#msgs]:find('Alias setup failed', 1, true) ~= nil)
    end)

    it('announces character when newly registered', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_true(setup.ensure_character('Squatched'))
        local found = false
        for _, msg in ipairs(msgs) do
            if msg:find('Squatched', 1, true) and msg:find('registered', 1, true) then
                found = true
            end
        end
        assert.is_true(found)
    end)

    it('does not announce when character was already registered', function()
        cli_calls.config_show = {
            config = {
                installs = {
                    steam = {
                        path = '/ffxi',
                        characters = { a1b2c3d4 = { name = 'Squatched' } },
                    },
                },
            },
        }
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_true(setup.ensure_character('Squatched'))
        assert.are.equal(0, #msgs)
    end)

    it('ensure_character returns false when config_show fails', function()
        cli_calls.config_show = nil
        assert.is_false(setup.ensure_character('Squatched'))
        assert.is_nil(cli_calls.set_alias)
    end)

    it('ensure_character returns false when list_all returns nil', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.list_all = nil
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        assert.is_false(setup.ensure_character('Squatched'))
        assert.is_true(msgs[#msgs]:find('No character folders', 1, true) ~= nil)
    end)

    it('falls back to full character list when all are already aliased', function()
        cli_calls.config_show = { config = { installs = { steam = {} } } }
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'char1', name = 'Char1' },
                { id = 'char2', name = 'Char2' },
            },
        }
        local orig_mtime = process_mod.file_mtime
        process_mod.file_mtime = function()
            return nil
        end
        assert.is_true(setup.ensure_character('NewChar'))
        assert.are.equal('char1', cli_calls.set_alias.char_id)
        process_mod.file_mtime = orig_mtime
    end)
end)

describe('ensure_character candidate filtering', function()
    before_each(function()
        setup.install_ready = false
        setup.zoned_since_load = false
        setup.noticed_zone = false
        setup.learned = {}
        cli_calls = {
            config_show = { config = { installs = { steam = {} } } },
            list_all = nil,
            ffxi_root = 'C:/ffxi',
        }
    end)

    -- Table-driven: which character ID is picked given a mix of aliased and
    -- unaliased entries, with mtime unavailable (picks first candidate).
    local filter_cases = {
        {
            desc = 'single unaliased character is selected',
            chars = { { id = 'only' } },
            expected = 'only',
        },
        {
            desc = 'aliased character is excluded; unaliased is selected',
            chars = { { id = 'taken', name = 'Other' }, { id = 'free' } },
            expected = 'free',
        },
        {
            desc = 'empty-string name is treated as unaliased',
            chars = { { id = 'blank', name = '' }, { id = 'real', name = 'Real' } },
            expected = 'blank',
        },
        {
            desc = 'all aliased: falls back to full list and picks first',
            chars = { { id = 'aa', name = 'Alice' }, { id = 'bb', name = 'Bob' } },
            expected = 'aa',
        },
        {
            desc = 'single unaliased among several aliased is selected',
            chars = {
                { id = 'aa', name = 'Alice' },
                { id = 'bb' },
                { id = 'cc', name = 'Carol' },
            },
            expected = 'bb',
        },
        {
            desc = 'first of multiple unaliased is selected when mtime unavailable',
            chars = { { id = 'xx' }, { id = 'yy' } },
            expected = 'xx',
        },
    }

    for _, case in ipairs(filter_cases) do
        local c = case
        it(c.desc, function()
            cli_calls.list_all = { user_dir = 'C:/ffxi/USER', characters = c.chars }
            local orig = process_mod.file_mtime
            process_mod.file_mtime = function()
                return nil
            end
            assert.is_true(setup.ensure_character('NewChar'))
            assert.are.equal(c.expected, cli_calls.set_alias.char_id)
            process_mod.file_mtime = orig
        end)
    end

    it('mtime picks newest unaliased over older unaliased, ignoring aliased', function()
        cli_calls.list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = {
                { id = 'old_free' },
                { id = 'taken', name = 'Other' },
                { id = 'new_free' },
            },
        }
        local orig = process_mod.file_mtime
        process_mod.file_mtime = function(path)
            if path:find('new_free', 1, true) then
                return '20250101120000'
            end
            if path:find('old_free', 1, true) then
                return '20200101120000'
            end
            return nil
        end
        assert.is_true(setup.ensure_character('NewChar'))
        assert.are.equal('new_free', cli_calls.set_alias.char_id)
        process_mod.file_mtime = orig
    end)
end)

describe('setup.on_zone edge cases', function()
    before_each(function()
        setup.install_ready = false
        setup.zoned_since_load = false
        setup.noticed_zone = false
        setup.learned = {}
        cli_calls = {
            config_show = { config = { installs = { steam = {} } } },
            list_all = {
                user_dir = 'C:/ffxi/USER',
                characters = { { id = 'a1b2c3d4' } },
            },
            ffxi_root = 'C:/ffxi',
        }
    end)

    it('only announces ready once', function()
        local msgs = {}
        windower.add_to_chat = function(_, msg)
            msgs[#msgs + 1] = msg
        end
        setup.on_zone()
        setup.on_zone()
        local ready_count = 0
        for _, msg in ipairs(msgs) do
            if msg:find('Ready!', 1, true) then
                ready_count = ready_count + 1
            end
        end
        assert.are.equal(1, ready_count)
    end)

    it('skips alias when player name is unavailable', function()
        windower.ffxi.get_player = function()
            return nil
        end
        setup.on_zone()
        assert.is_nil(cli_calls.set_alias)
    end)
end)
