-- .luacheckrc
-- Windower 4 globals — these are injected by the Windower environment

globals = {
    -- Windower API
    "windower",
    "_addon",
    -- Standard Lua globals Windower exposes
    "io",
    "os",
    "math",
    "string",
    "table",
    "type",
    "pairs",
    "ipairs",
    "tostring",
    "tonumber",
    "require",
    "pcall",
    "error",
    "assert",
    "print",
    "setmetatable",
    "getmetatable",
    "rawget",
    "rawset",
    "select",
    "unpack",
    "next",
}

-- Ignore line-length warnings (macro lines can be long in comments/TODOs)
max_line_length = 120

-- Allow unused loop variables named _ or _something
unused_args = false
