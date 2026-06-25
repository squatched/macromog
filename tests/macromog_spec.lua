package.path = './?.lua;./?/init.lua;' .. package.path

_G._addon = {
    name = 'Macromog',
    author = 'Squatched',
    version = '0.1.0',
    commands = { 'macromog', 'mmog' },
}

local events = {}
local chat_msgs = {}
local cli_calls = {}
local diag_called = false

_G.windower = {
    addon_path = '/tmp/macromog_entry/',
    add_to_chat = function(_, msg)
        chat_msgs[#chat_msgs + 1] = msg
    end,
    register_event = function(name, fn)
        events[name] = fn
    end,
    file_exists = function()
        return false
    end,
    debug = function() end,
    windower_path = 'C:/Windower4/',
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
        return 0, ''
    end,
    list_all = function()
        return cli_calls.list_all
    end,
    export = function(path, name)
        cli_calls.export = { path = path, name = name }
        return cli_calls.export_code or 0, cli_calls.export_out or ''
    end,
    import = function(path, name)
        cli_calls.import = { path = path, name = name }
        return cli_calls.import_code or 0, cli_calls.import_out or ''
    end,
    validate = function(path)
        cli_calls.validate = { path = path }
        return cli_calls.validate_code or 0, cli_calls.validate_out or ''
    end,
    backup = function(name)
        cli_calls.backup = { name = name }
        return cli_calls.backup_code or 0, cli_calls.backup_out or ''
    end,
    debug_all = function()
        return 0, ''
    end,
}

package.loaded['lib/detect'] = {
    ffxi_root = function()
        return cli_calls.ffxi_root
    end,
    suggest_install_name = function()
        return 'steam'
    end,
    run_diag = function()
        diag_called = true
    end,
}

package.loaded['lib/log'] = nil
package.loaded['lib/setup'] = nil
package.loaded['macromog'] = nil

dofile('macromog.lua')

local setup = require('lib/setup')

local function last_chat()
    return chat_msgs[#chat_msgs] or ''
end

local function reset_state()
    chat_msgs = {}
    diag_called = false
    cli_calls = {
        config_show = { config = { installs = { steam = { path = 'C:/ffxi' } } } },
        list_all = {
            user_dir = 'C:/ffxi/USER',
            characters = { { id = 'a1b2c3d4' } },
        },
        ffxi_root = 'C:/ffxi',
        export_code = 0,
        import_code = 0,
        validate_code = 0,
        backup_code = 0,
    }
    setup.install_ready = false
    setup.zoned_since_load = false
    setup.noticed_zone = false
    setup.learned = {}
end

describe('macromog command routing', function()
    before_each(reset_state)

    it('shows help for empty command', function()
        events['addon command']('')
        assert.is_true(last_chat():find('Commands:', 1, true) ~= nil)
    end)

    it('shows help for help command', function()
        events['addon command']('help')
        assert.is_true(last_chat():find('export', 1, true) ~= nil)
    end)

    it('rejects unknown commands', function()
        events['addon command']('nope')
        local joined = table.concat(chat_msgs, '\n')
        assert.is_true(joined:find('Unknown command', 1, true) ~= nil)
        assert.is_true(joined:find('Commands:', 1, true) ~= nil)
    end)

    it('blocks export before readiness', function()
        events['addon command']('export', 'out.yml')
        assert.is_true(last_chat():find('not configured', 1, true) ~= nil)
        assert.is_nil(cli_calls.export)
    end)

    it('runs export when ready', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('export', 'out.yml')
        assert.is_not_nil(cli_calls.export)
        assert.is_true(last_chat():find('Exported', 1, true) ~= nil)
    end)

    it('reports missing import file', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('import', 'missing.yml')
        assert.is_true(last_chat():find('File not found', 1, true) ~= nil)
    end)

    it('reports import usage without filename', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('import')
        assert.is_true(last_chat():find('Usage:', 1, true) ~= nil)
    end)

    it('handles debug on and off', function()
        events['addon command']('debug', 'on')
        assert.is_true(last_chat():find('enabled', 1, true) ~= nil)
        events['addon command']('debug', 'off')
        assert.is_true(last_chat():find('disabled', 1, true) ~= nil)
    end)

    it('shows debug usage for invalid mode', function()
        events['addon command']('debug', 'maybe')
        assert.is_true(last_chat():find('debug on', 1, true) ~= nil)
    end)

    it('runs diag and reports log path', function()
        events['addon command']('diag')
        assert.is_true(diag_called)
        assert.is_true(last_chat():find('debug.log', 1, true) ~= nil)
    end)
end)

describe('macromog lifecycle events', function()
    before_each(reset_state)

    it('load event configures install and greets user', function()
        cli_calls.config_show_sequence = {
            { config = {} },
            { config = { installs = { steam = { path = 'C:/ffxi' } } } },
        }
        events.load()
        assert.is_true(setup.install_ready)
        assert.is_true(last_chat():find('Kupomog', 1, true) ~= nil)
    end)

    it('zone packet marks ready after first zone', function()
        setup.install_ready = true
        events['incoming chunk'](0x0A)
        assert.is_true(setup.zoned_since_load)
        assert.is_true(last_chat():find('Ready!', 1, true) ~= nil)
    end)

    it('ignores non-zone packets', function()
        events['incoming chunk'](0x01)
        assert.is_false(setup.zoned_since_load)
    end)
end)
