package.path = './?.lua;./?/init.lua;' .. package.path

-- Minimal Windower environment mock required by lib/macros
_G.windower = {
    add_to_chat = function() end,
    ffxi = {
        get_player = function()
            return { name = 'TestChar' }
        end,
    },
    addon_path = '/tmp/macromog_test/',
}

local macros = require('lib/macros')

describe('macros.read', function()
    it('returns nil (not yet implemented)', function()
        assert.is_nil(macros.read())
    end)
end)

describe('macros.write', function()
    it('completes without error (not yet implemented)', function()
        assert.has_no.errors(function()
            macros.write({})
        end)
    end)
end)

describe('macros.backup', function()
    it('returns false when read() returns nil', function()
        -- macros.read() is a stub that returns nil, so backup cannot proceed
        assert.is_false(macros.backup())
    end)

    it('returns false when get_player returns nil', function()
        local orig = _G.windower.ffxi.get_player
        _G.windower.ffxi.get_player = function()
            return nil
        end
        assert.is_false(macros.backup())
        _G.windower.ffxi.get_player = orig
    end)
end)
