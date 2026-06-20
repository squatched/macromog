-- lib/validate.lua
-- Validates macro data tables against FFXI constraints.

local validate = {}

local LIMITS = {
    books    = 40,
    sets     = 10,   -- per book
    macros   = 20,   -- per set per modifier
    lines    = 6,    -- per macro
    name_len = 8,    -- chars, macro and book names
    line_len = 255,  -- chars per line (verify against live client)
}

-- Returns true, nil on success; false, error_string on failure.
function validate.macros(data)
    if type(data) ~= 'table' then
        return false, 'data must be a table'
    end
    if type(data.books) ~= 'table' then
        return false, 'missing top-level "books" key'
    end

    local book_count = 0
    for book_idx, book in pairs(data.books) do
        book_count = book_count + 1
        if book_count > LIMITS.books then
            return false, ('too many books: %d (max %d)'):format(book_count, LIMITS.books)
        end
        if type(book_idx) ~= 'number' or book_idx < 0 or book_idx > LIMITS.books - 1 then
            return false, ('invalid book index: %s'):format(tostring(book_idx))
        end
        if book.name ~= nil and #book.name > LIMITS.name_len then
            return false, ('book %d name too long (%d > %d)'):format(
                book_idx, #book.name, LIMITS.name_len)
        end
        if type(book.sets) ~= 'table' then
            return false, ('book %d missing "sets"'):format(book_idx)
        end

        local set_count = 0
        for set_idx, set in pairs(book.sets) do
            set_count = set_count + 1
            if set_count > LIMITS.sets then
                return false, ('book %d: too many sets (max %d)'):format(book_idx, LIMITS.sets)
            end

            for mod_key, slot_macros in pairs(set) do
                if mod_key ~= 'ctrl' and mod_key ~= 'alt' then
                    return false, ('book %d set %d: unknown modifier "%s"'):format(
                        book_idx, set_idx, tostring(mod_key))
                end

                local macro_count = 0
                for macro_idx, macro in pairs(slot_macros) do
                    macro_count = macro_count + 1
                    if macro_count > LIMITS.macros then
                        return false, ('book %d set %d %s: too many macros (max %d)'):format(
                            book_idx, set_idx, mod_key, LIMITS.macros)
                    end
                    if macro.name ~= nil and #macro.name > LIMITS.name_len then
                        return false, ('book %d set %d %s macro %d: name too long'):format(
                            book_idx, set_idx, mod_key, macro_idx)
                    end
                    if type(macro.contents) == 'table' then
                        if #macro.contents > LIMITS.lines then
                            return false, ('book %d set %d %s macro %d: too many lines (max %d)'):format(
                                book_idx, set_idx, mod_key, macro_idx, LIMITS.lines)
                        end
                        for line_idx, line in ipairs(macro.contents) do
                            if type(line) ~= 'string' then
                                return false, ('book %d set %d %s macro %d line %d: not a string'):format(
                                    book_idx, set_idx, mod_key, macro_idx, line_idx)
                            end
                            if #line > LIMITS.line_len then
                                return false, ('book %d set %d %s macro %d line %d: too long'):format(
                                    book_idx, set_idx, mod_key, macro_idx, line_idx)
                            end
                        end
                    end
                end
            end
        end
    end

    return true, nil
end

return validate
