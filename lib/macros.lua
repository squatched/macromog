-- lib/macros.lua
-- FFXI macro DAT file read/write and backup operations.
-- The plugin delegates heavy lifting (DAT parsing, YAML conversion) to the
-- CLI binary. These stubs represent the in-plugin surface; full implementation
-- requires invoking the CLI or bundling its DAT parsing logic.

local yaml = require('lib/yaml')

local macros = {}

function macros.read()
    -- TODO: Delegate to CLI: `macromog export --char <id> --output <tmp>`
    -- then read the resulting YAML. Alternatively, parse mcr*.dat files
    -- directly (see POLUtils / xi-tinkerer for DAT format reference).
    -- Returns sparse table: { books = { [idx] = { name=..., sets={...} } } }
    windower.add_to_chat(207, '[Macromog] macros.read() not yet implemented.')
    return nil
end

function macros.write(data)
    -- TODO: Delegate to CLI: `macromog import <file> --char <id>`.
    -- Write `data` to a temp YAML file, then invoke the CLI to apply it.
    -- Must zero out cleared slots so stale entries do not survive.
    windower.add_to_chat(207, '[Macromog] macros.write() not yet implemented.')
end

function macros.backup()
    local player = windower.ffxi.get_player()
    if not player then
        windower.add_to_chat(207, '[Macromog] Cannot backup: no player data.')
        return false
    end

    -- TODO: Delegate to CLI: `macromog backup --char <id>`.
    -- The CLI copies mcr*.dat files directly to a timestamped directory.
    -- Current fallback: round-trip through YAML (removed when CLI is available).
    local data = macros.read()
    if not data then
        windower.add_to_chat(207, '[Macromog] Cannot backup: failed to read macros.')
        return false
    end

    local timestamp = os.date('%Y%m%d_%H%M%S')
    local filename = player.name .. '_macros_backup_' .. timestamp .. '.yml'
    local path = windower.addon_path .. 'data/' .. filename

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
