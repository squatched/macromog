package.path = './?.lua;./?/init.lua;' .. package.path

local json = require('lib/json')

describe('json.decode happy paths', function()
    local cases = {
        {
            name = 'simple object',
            raw = '{"path":"/tmp","version":1}',
            check = function(data)
                assert.are.same('/tmp', data.path)
                assert.are.equal(1, data.version)
            end,
        },
        {
            name = 'nested config show shape',
            raw = '{"config":{"installs":{"steam":{"path":"/ffxi"}}}}',
            check = function(data)
                assert.are.same('/ffxi', data.config.installs.steam.path)
            end,
        },
        {
            name = 'arrays',
            raw = '{"characters":[{"id":"abc"}]}',
            check = function(data)
                assert.are.equal('abc', data.characters[1].id)
            end,
        },
        {
            name = 'booleans and null',
            raw = '{"ok":true,"off":false,"missing":null}',
            check = function(data)
                assert.is_true(data.ok)
                assert.is_false(data.off)
                assert.is_nil(data.missing)
            end,
        },
        {
            name = 'escaped strings',
            raw = '{"msg":"line\\none\\t tab"}',
            check = function(data)
                assert.are.equal('line\none\t tab', data.msg)
            end,
        },
        {
            name = 'empty object and array',
            raw = '{"empty":{},"list":[]}',
            check = function(data)
                assert.are.same({}, data.empty)
                assert.are.same({}, data.list)
            end,
        },
        {
            name = 'negative number',
            raw = '{"n": -42}',
            check = function(data)
                assert.are.equal(-42, data.n)
            end,
        },
    }

    for _, tc in ipairs(cases) do
        it('parses ' .. tc.name, function()
            local data = json.decode(tc.raw)
            tc.check(data)
        end)
    end
end)

describe('json.decode failures', function()
    local cases = {
        { name = 'invalid object', raw = '{bad', want_err = true },
        { name = 'empty input', raw = '', want_err = 'empty json' },
        { name = 'missing colon', raw = '{"a" 1}', want_err = true },
        { name = 'unclosed string', raw = '{"a":"oops}', want_err = true },
        { name = 'trailing garbage', raw = '{"a":1}xyz', check = function(data)
            assert.are.equal(1, data.a)
        end },
        { name = 'unknown escape', raw = '{"a":"\\x"}', check = function(data)
            assert.are.equal('x', data.a)
        end },
    }

    for _, tc in ipairs(cases) do
        it('handles ' .. tc.name, function()
            local data, err = json.decode(tc.raw)
            if tc.check then
                tc.check(data)
            else
                assert.is_nil(data)
                if tc.want_err == true then
                    assert.is_string(err)
                else
                    assert.are.equal(tc.want_err, err)
                end
            end
        end)
    end
end)
