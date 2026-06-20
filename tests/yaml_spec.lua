package.path = './?.lua;./?/init.lua;' .. package.path

local yaml = require('lib/yaml')

describe('yaml.dump', function()
    it('returns a string', function()
        assert.is_string(yaml.dump({}))
    end)

    it('returns a string for nil input', function()
        assert.is_string(yaml.dump(nil))
    end)
end)

describe('yaml.parse', function()
    it('returns nil as first value (stub)', function()
        local result, _ = yaml.parse('')
        assert.is_nil(result)
    end)

    it('returns an error string as second value (stub)', function()
        local _, err = yaml.parse('books:')
        assert.is_string(err)
    end)
end)
