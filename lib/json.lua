-- Minimal JSON helpers for macromog CLI --output json responses.

local json = {}

local function skip_ws(s, i)
    while i <= #s do
        local c = s:sub(i, i)
        if c ~= ' ' and c ~= '\t' and c ~= '\n' and c ~= '\r' then
            return i
        end
        i = i + 1
    end
    return i
end

local function parse_string(s, i)
    if s:sub(i, i) ~= '"' then
        return nil, i
    end
    i = i + 1
    local parts = {}
    while i <= #s do
        local c = s:sub(i, i)
        if c == '"' then
            return table.concat(parts), i + 1
        end
        if c == '\\' then
            local esc = s:sub(i + 1, i + 1)
            if esc == '"' or esc == '\\' or esc == '/' then
                parts[#parts + 1] = esc
                i = i + 2
            elseif esc == 'n' then
                parts[#parts + 1] = '\n'
                i = i + 2
            elseif esc == 't' then
                parts[#parts + 1] = '\t'
                i = i + 2
            else
                parts[#parts + 1] = esc
                i = i + 2
            end
        else
            parts[#parts + 1] = c
            i = i + 1
        end
    end
    return nil, i
end

local function parse_value(s, i)
    i = skip_ws(s, i)
    local c = s:sub(i, i)
    if c == '"' then
        return parse_string(s, i)
    end
    if c == '{' then
        return json.parse_object(s, i)
    end
    if c == '[' then
        return json.parse_array(s, i)
    end
    local num = s:match('^%-?%d+%.?%d*', i)
    if num then
        return tonumber(num), i + #num
    end
    if s:sub(i, i + 3) == 'true' then
        return true, i + 4
    end
    if s:sub(i, i + 4) == 'false' then
        return false, i + 5
    end
    if s:sub(i, i + 3) == 'null' then
        return nil, i + 4
    end
    return nil, i
end

function json.parse_object(s, i)
    i = skip_ws(s, i)
    if s:sub(i, i) ~= '{' then
        return nil, i
    end
    i = i + 1
    local obj = {}
    i = skip_ws(s, i)
    if s:sub(i, i) == '}' then
        return obj, i + 1
    end
    while i <= #s do
        i = skip_ws(s, i)
        local key
        key, i = parse_string(s, i)
        if not key then
            return nil, i
        end
        i = skip_ws(s, i)
        if s:sub(i, i) ~= ':' then
            return nil, i
        end
        i = i + 1
        local val
        val, i = parse_value(s, i)
        obj[key] = val
        i = skip_ws(s, i)
        local sep = s:sub(i, i)
        if sep == '}' then
            return obj, i + 1
        end
        if sep ~= ',' then
            return nil, i
        end
        i = i + 1
    end
    return nil, i
end

function json.parse_array(s, i)
    i = skip_ws(s, i)
    if s:sub(i, i) ~= '[' then
        return nil, i
    end
    i = i + 1
    local arr = {}
    i = skip_ws(s, i)
    if s:sub(i, i) == ']' then
        return arr, i + 1
    end
    while i <= #s do
        local val
        val, i = parse_value(s, i)
        arr[#arr + 1] = val
        i = skip_ws(s, i)
        local sep = s:sub(i, i)
        if sep == ']' then
            return arr, i + 1
        end
        if sep ~= ',' then
            return nil, i
        end
        i = i + 1
    end
    return nil, i
end

function json.decode(s)
    if not s or s == '' then
        return nil, 'empty json'
    end
    local val = parse_value(s, 1)
    if not val then
        return nil, 'parse error'
    end
    return val
end

return json
