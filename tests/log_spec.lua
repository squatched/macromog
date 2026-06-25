package.path = './?.lua;./?/init.lua;' .. package.path

_G.windower = {
    addon_path = '/tmp/macromog_log_test/',
    add_to_chat = function() end,
    file_exists = function(path)
        return path:find('.debug', 1, true) ~= nil
    end,
    debug = function(msg)
        windower._debug_msg = msg
    end,
}

package.loaded['lib/log'] = nil
local log = require('lib/log')

describe('lib/log', function()
    before_each(function()
        log.enabled = false
        log.session = false
    end)

    it('debug writes only when active', function()
        log.debug('hidden')
        log.begin_session()
        log.debug('visible')
        assert.is_not_nil(windower._debug_msg)
        assert.is_true(windower._debug_msg:find('visible', 1, true) ~= nil)
    end)

    it('refresh reads flag file', function()
        log.refresh()
        assert.is_true(log.enabled)
    end)
end)
