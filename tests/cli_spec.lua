package.path = './?.lua;./?/init.lua;' .. package.path

_G.jit = { arch = 'x86' }
_G.windower = {
    addon_path = '/tmp/macromog_test/',
}

local cli = require('lib/cli')

describe('cli.binary_path', function()
    it('selects 386 on x86', function()
        assert.are.equal('/tmp/macromog_test/bin/macromog-windows-386.exe', cli.binary_path())
    end)

    it('selects amd64 on x64', function()
        _G.jit.arch = 'x64'
        assert.are.equal('/tmp/macromog_test/bin/macromog-windows-amd64.exe', cli.binary_path())
    end)
end)

describe('cli.run', function()
    local orig_popen = io.popen

    after_each(function()
        io.popen = orig_popen
    end)

    it('captures subprocess output', function()
        io.popen = function()
            return {
                read = function()
                    return 'ok'
                end,
                close = function()
                    return true, 'exit', 0
                end,
            }
        end
        local code, out = cli.run({ 'config', 'path' })
        assert.are.equal(0, code)
        assert.are.equal('ok', out)
    end)

    it('quotes arguments with spaces', function()
        io.popen = function(cmd)
            assert.is_true(cmd:find('\"spaced arg\"', 1, true) ~= nil)
            return {
                read = function()
                    return ''
                end,
                close = function()
                    return true, 'exit', 0
                end,
            }
        end
        cli.run({ 'config', 'add-install', 'steam', 'spaced arg' })
    end)

    it('returns error when popen fails', function()
        io.popen = function()
            return nil
        end
        local code, out, err = cli.run({ 'list' })
        assert.are.equal(1, code)
        assert.are.equal('failed to run CLI', err)
    end)

    it('returns numeric exit code when command fails', function()
        io.popen = function()
            return {
                read = function()
                    return 'stderr'
                end,
                close = function()
                    return false, 'exit', 2
                end,
            }
        end
        local code, out = cli.run({ 'list' })
        assert.are.equal(2, code)
        assert.are.equal('stderr', out)
    end)

    it('quotes arguments containing double quotes', function()
        io.popen = function(cmd)
            assert.is_true(cmd:find('has\\"quote.yml', 1, true) ~= nil)
            return {
                read = function()
                    return ''
                end,
                close = function()
                    return true, 'exit', 0
                end,
            }
        end
        cli.run({ 'import', 'has"quote.yml' })
    end)
end)

describe('cli wrappers', function()
    local orig_popen = io.popen
    local last_cmd

    after_each(function()
        io.popen = orig_popen
        last_cmd = nil
    end)

    local function mock_popen(output, close_ok, close_code)
        io.popen = function(cmd)
            last_cmd = cmd
            return {
                read = function()
                    return output
                end,
                close = function()
                    if close_ok == false then
                        return false, 'exit', close_code or 1
                    end
                    return true, 'exit', 0
                end,
            }
        end
    end

    it('delegates config_show to run_json', function()
        mock_popen('{"config":{}}')
        local data = cli.config_show()
        assert.are.same({}, data.config)
        assert.is_true(last_cmd:find('config show', 1, true) ~= nil)
    end)

    it('does not enable MACROMOG_DEBUG for json when addon debug is on', function()
        package.loaded['lib/log'] = {
            active = function()
                return true
            end,
            debug = function() end,
        }
        package.loaded['lib/cli'] = nil
        local debug_cli = require('lib/cli')
        mock_popen('{"config":{}}')
        local data = debug_cli.config_show()
        assert.are.same({}, data.config)
        assert.is_false(last_cmd:find('MACROMOG_DEBUG', 1, true) ~= nil)
        package.loaded['lib/log'] = nil
        package.loaded['lib/cli'] = nil
        require('lib/cli')
    end)

    it('enables MACROMOG_DEBUG for debug_all', function()
        package.loaded['lib/log'] = {
            active = function()
                return false
            end,
            debug = function() end,
        }
        package.loaded['lib/cli'] = nil
        local debug_cli = require('lib/cli')
        mock_popen('paths:\nenvironment:\n')
        debug_cli.debug_all()
        assert.is_true(last_cmd:find('MACROMOG_DEBUG', 1, true) ~= nil)
        package.loaded['lib/log'] = nil
        package.loaded['lib/cli'] = nil
        require('lib/cli')
    end)

    it('delegates list_all to run_json', function()
        mock_popen('{"user_dir":"C:/ffxi/USER"}')
        local data = cli.list_all()
        assert.are.equal('C:/ffxi/USER', data.user_dir)
    end)

    it('delegates config_add_install to run', function()
        mock_popen('registered')
        local code, out = cli.config_add_install('steam', 'C:/ffxi')
        assert.are.equal(0, code)
        assert.are.equal('registered', out)
        assert.is_true(last_cmd:find('add-install steam', 1, true) ~= nil)
    end)

    it('delegates config_set_alias with extra args', function()
        mock_popen('')
        cli.config_set_alias('abc', 'Name', { '--install', 'steam' })
        assert.is_true(last_cmd:find('set-alias abc Name --install steam', 1, true) ~= nil)
    end)

    it('delegates export, import, validate, and backup', function()
        mock_popen('')
        cli.export('out.yml', 'Squatched')
        assert.is_true(last_cmd:find('export', 1, true) ~= nil)
        cli.import('in.yml', 'Squatched')
        assert.is_true(last_cmd:find('import in.yml', 1, true) ~= nil)
        cli.validate('in.yml')
        assert.is_true(last_cmd:find('validate in.yml', 1, true) ~= nil)
        cli.backup('Squatched')
        assert.is_true(last_cmd:find('backup', 1, true) ~= nil)
    end)
end)

describe('cli.run_json', function()
    local orig_popen = io.popen

    after_each(function()
        io.popen = orig_popen
    end)

    it('parses json output', function()
        io.popen = function()
            return {
                read = function()
                    return '{"user_dir":"/tmp/USER"}'
                end,
                close = function()
                    return true, 'exit', 0
                end,
            }
        end
        local data = cli.run_json({ 'list' })
        assert.are.equal('/tmp/USER', data.user_dir)
    end)

    it('returns cli error output on failure', function()
        io.popen = function()
            return {
                read = function()
                    return 'config missing'
                end,
                close = function()
                    return false, 'exit', 1
                end,
            }
        end
        local data, err = cli.run_json({ 'config', 'show' })
        assert.is_nil(data)
        assert.are.equal('config missing', err)
    end)

    it('returns parse error for invalid json', function()
        io.popen = function()
            return {
                read = function()
                    return 'not json'
                end,
                close = function()
                    return true, 'exit', 0
                end,
            }
        end
        local data, err = cli.run_json({ 'list' })
        assert.is_nil(data)
        assert.is_string(err)
    end)
end)