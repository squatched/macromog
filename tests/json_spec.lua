package.path = './?.lua;./?/init.lua;' .. package.path

local json = require('lib/json')

describe('json.decode', function()
    it('parses a simple object', function()
        local data = json.decode('{"path":"/tmp","version":1}')
        assert.are.same('/tmp', data.path)
        assert.are.equal(1, data.version)
    end)

    it('parses nested config show shape', function()
        local raw = '{"config":{"installs":{"steam":{"path":"/ffxi"}}}}'
        local data = json.decode(raw)
        assert.are.same('/ffxi', data.config.installs.steam.path)
    end)

    it('parses arrays', function()
        local data = json.decode('{"characters":[{"id":"abc"}]}')
        assert.are.equal('abc', data.characters[1].id)
    end)

    it('returns nil on invalid input', function()
        local data, err = json.decode('{bad')
        assert.is_nil(data)
        assert.is_string(err)
    end)

    it('parses booleans and null', function()
        local data = json.decode('{"ok":true,"off":false,"missing":null}')
        assert.is_true(data.ok)
        assert.is_false(data.off)
        assert.is_nil(data.missing)
    end)

    it('parses escaped strings', function()
        local data = json.decode('{"msg":"line\\none\\t tab"}')
        assert.are.equal('line\none\t tab', data.msg)
    end)

    it('returns nil on empty input', function()
        local data, err = json.decode('')
        assert.is_nil(data)
        assert.are.equal('empty json', err)
    end)
end)