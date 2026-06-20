-- lib/macros.lua
-- FFXI macro memory read/write and backup operations
--
-- TODO: Locate the macro memory base address via Windower's memory module.
-- Macro data lives in a fixed region of the FFXI process; offsets change
-- with client patches. Reference: Windower docs / Ashita macro plugin source.

local yaml = require('lib/yaml')

local macros = {}

-- Memory base — must be determined per-client-version
-- local MACRO_BASE = 0x????????

function macros.read()
    -- TODO: Walk FFXI memory at MACRO_BASE, build sparse table:
    --   { books = { [idx] = { name=..., sets = { [idx] = { ctrl={...}, alt={...} } } } } }
    -- Each macro slot: name (8 chars), 6 lines (255 chars each), ctrl/alt flag
    windower.add_to_chat(207, '[Macromog] macros.read() not yet implemented.')
    return nil
end

function macros.write(data)
    -- TODO: Serialize `data` back into FFXI process memory at MACRO_BASE.
    -- Must zero out cleared slots so stale entries don't survive.
    windower.add_to_chat(207, '[Macromog] macros.write() not yet implemented.')
end

function macros.backup()
    local player = windower.ffxi.get_player()
    if not player then
        windower.add_to_chat(207, '[Macromog] Cannot backup: no player data.')
        return false
    end

    local data = macros.read()
    if not data then
        windower.add_to_chat(207, '[Macromog] Cannot backup: failed to read macros.')
        return false
    end

    local timestamp = os.date('%Y%m%d_%H%M%S')
    local filename  = player.name .. '_macros_backup_' .. timestamp .. '.yml'
    local path      = windower.addon_path .. 'data/' .. filename

    local f = io.open(path, 'w')
    if not f then
        windower.add_to_chat(207, '[Macromog] Cannot write backup: ' .. path)
        return false
    end
    f:write(yaml.dump(data))
    f:close()
    windower.add_to_chat(207, '[Macromog] Backup saved: ' .. filename)
    return true
end

return macros
