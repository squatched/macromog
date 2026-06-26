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
    import = function(path, name, backup_dir, no_backup)
        cli_calls.import = { path = path, name = name, backup_dir = backup_dir, no_backup = no_backup }
        return cli_calls.import_code or 0, cli_calls.import_out or ''
    end,
    validate = function(path)
        cli_calls.validate = { path = path }
        return cli_calls.validate_code or 0, cli_calls.validate_out or ''
    end,
    backup = function(name, dest_dir)
        cli_calls.backup = { name = name, dest_dir = dest_dir }
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
    setup.pending_import = nil
end

describe('macromog command routing', function()
    before_each(reset_state)

    it('shows help for empty command', function()
        events['addon command']('')
        assert.is_true(table.concat(chat_msgs, '\n'):find('Commands:', 1, true) ~= nil)
    end)

    it('shows help for help command', function()
        events['addon command']('help')
        assert.is_true(table.concat(chat_msgs, '\n'):find('export', 1, true) ~= nil)
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

    it('login event is registered', function()
        assert.is_function(events.login)
    end)

    it('login event resets zone flags', function()
        setup.zoned_since_load = true
        setup.noticed_zone = true
        events.login()
        assert.is_false(setup.zoned_since_load)
        assert.is_false(setup.noticed_zone)
    end)

    -- Regression: second zone packet in the same session (no logout in between)
    -- must not re-register or fire the ready message a second time.
    it('second zone packet without login is gated by zoned_since_load', function()
        setup.install_ready = true
        events['incoming chunk'](0x0A)
        assert.is_true(setup.learned.Squatched)
        local msg_count = #chat_msgs
        cli_calls.set_alias = nil

        events['incoming chunk'](0x0A)

        assert.is_nil(cli_calls.set_alias)
        assert.are.equal(msg_count, #chat_msgs)
    end)

    -- Regression: after login fires, the next zone packet must trigger character
    -- learning for the new character without clobbering the first character's alias.
    it('zone packet after login registers new character without clobbering first', function()
        setup.install_ready = true

        -- First character zones in.
        events['incoming chunk'](0x0A)
        assert.is_true(setup.learned.Squatched)

        -- Switch to a second character.
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

        -- Login now eagerly registers the new character.
        events.login()
        assert.is_true(setup.zoned_since_load)
        assert.is_true(setup.learned.Altchar)
        assert.are.equal('b2c3d4e5', cli_calls.set_alias.char_id)
        assert.are.equal('Altchar', cli_calls.set_alias.name)

        windower.ffxi.get_player = orig_get_player
    end)
end)

describe('export command', function()
    before_each(reset_state)

    it('generates timestamped filename when no filename given', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('export')
        assert.is_not_nil(cli_calls.export)
        assert.is_true(cli_calls.export.path:find('data/', 1, true) ~= nil)
        assert.is_true(last_chat():find('Exported', 1, true) ~= nil)
    end)

    it('surfaces export CLI failure to user', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.export_code = 1
        cli_calls.export_out = 'disk full'
        events['addon command']('export', 'out.yml')
        assert.is_true(last_chat():find('Export failed', 1, true) ~= nil)
        assert.is_true(last_chat():find('disk full', 1, true) ~= nil)
    end)

    it('passes filename in path and char name to cli', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('export', 'out.yml')
        assert.are.equal('Squatched', cli_calls.export.name)
        assert.is_true(cli_calls.export.path:find('out.yml', 1, true) ~= nil)
    end)

    it('does not call cli when not ready', function()
        events['addon command']('export', 'out.yml')
        assert.is_nil(cli_calls.export)
    end)
end)

describe('import command', function()
    local orig_io_open

    before_each(function()
        reset_state()
        orig_io_open = io.open
    end)

    after_each(function()
        io.open = orig_io_open
    end)

    it('does not call cli when not ready', function()
        io.open = function() return { close = function() end } end
        events['addon command']('import', 'macros.yml')
        assert.is_true(last_chat():find('not configured', 1, true) ~= nil)
        assert.is_nil(cli_calls.import)
    end)

    it('imports file successfully and sets pending import', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        io.open = function() return { close = function() end } end
        events['addon command']('import', 'macros.yml')
        assert.is_not_nil(cli_calls.import)
        assert.is_true(last_chat():find('Imported', 1, true) ~= nil)
        assert.is_true(last_chat():find('Zone', 1, true) ~= nil)
        assert.is_not_nil(setup.pending_import)
    end)

    it('zone-in applies pending import and clears it', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        io.open = function() return { close = function() end } end
        events['addon command']('import', 'macros.yml')
        assert.is_not_nil(setup.pending_import)
        cli_calls.import = nil

        events['incoming chunk'](0x0A)

        assert.is_not_nil(cli_calls.import)
        assert.is_true(cli_calls.import.no_backup)
        assert.is_nil(setup.pending_import)
        assert.is_true(last_chat():find('applied', 1, true) ~= nil)
    end)

    it('surfaces import CLI failure to user', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.import_code = 1
        cli_calls.import_out = 'invalid macro data'
        io.open = function() return { close = function() end } end
        events['addon command']('import', 'macros.yml')
        assert.is_true(last_chat():find('Import failed', 1, true) ~= nil)
        assert.is_true(last_chat():find('invalid macro data', 1, true) ~= nil)
    end)

    it('passes path with filename, char name, and backup_dir to cli', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        io.open = function() return { close = function() end } end
        events['addon command']('import', 'mybook.yml')
        assert.are.equal('Squatched', cli_calls.import.name)
        assert.is_true(cli_calls.import.path:find('mybook.yml', 1, true) ~= nil)
        assert.is_true((cli_calls.import.backup_dir or ''):find('data', 1, true) ~= nil)
    end)
end)

describe('backup command', function()
    before_each(reset_state)

    it('does not call cli when not ready', function()
        events['addon command']('backup')
        assert.is_nil(cli_calls.backup)
        assert.is_true(last_chat():find('not configured', 1, true) ~= nil)
    end)

    it('reports backup location from cli output', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.backup_out = 'backed up to C:\\ffxi\\USER\\a1b2c3d4\\backups\\a1b2c3d4_20240101_120000'
        events['addon command']('backup')
        assert.is_not_nil(cli_calls.backup)
        assert.is_true(last_chat():find('backed up to', 1, true) ~= nil)
    end)

    it('falls back to generic message when cli output is empty', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.backup_out = ''
        events['addon command']('backup')
        assert.is_true(last_chat():find('Backup complete', 1, true) ~= nil)
    end)

    it('passes addon data dir as backup destination', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('backup')
        assert.is_not_nil(cli_calls.backup)
        assert.is_true((cli_calls.backup.dest_dir or ''):find('data', 1, true) ~= nil)
    end)

    it('surfaces backup CLI failure to user', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.backup_code = 1
        cli_calls.backup_out = 'no character found'
        events['addon command']('backup')
        assert.is_true(last_chat():find('Backup failed', 1, true) ~= nil)
        assert.is_true(last_chat():find('no character found', 1, true) ~= nil)
    end)
end)

describe('validate command', function()
    local orig_io_open

    before_each(function()
        reset_state()
        orig_io_open = io.open
    end)

    after_each(function()
        io.open = orig_io_open
    end)

    it('does not call cli when not ready', function()
        io.open = function() return { close = function() end } end
        events['addon command']('validate', 'macros.yml')
        assert.is_true(last_chat():find('not configured', 1, true) ~= nil)
        assert.is_nil(cli_calls.validate)
    end)

    it('shows usage when no filename given', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('validate')
        assert.is_true(last_chat():find('Usage:', 1, true) ~= nil)
        assert.is_nil(cli_calls.validate)
    end)

    it('reports file not found for missing file', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        events['addon command']('validate', 'nofile.yml')
        assert.is_true(last_chat():find('File not found', 1, true) ~= nil)
        assert.is_nil(cli_calls.validate)
    end)

    it('reports validation passed', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        io.open = function() return { close = function() end } end
        events['addon command']('validate', 'macros.yml')
        assert.is_true(last_chat():find('Validation passed', 1, true) ~= nil)
    end)

    it('surfaces validation failure to user', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        cli_calls.validate_code = 1
        cli_calls.validate_out = 'version: must be 1, got 99'
        io.open = function() return { close = function() end } end
        events['addon command']('validate', 'bad.yml')
        assert.is_true(last_chat():find('Validation failed', 1, true) ~= nil)
        assert.is_true(last_chat():find('version', 1, true) ~= nil)
    end)

    it('passes path with filename to cli', function()
        setup.install_ready = true
        setup.zoned_since_load = true
        io.open = function() return { close = function() end } end
        events['addon command']('validate', 'check.yml')
        assert.is_true(cli_calls.validate.path:find('check.yml', 1, true) ~= nil)
    end)
end)
