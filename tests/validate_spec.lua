package.path = './?.lua;./?/init.lua;' .. package.path

local validate = require('lib/validate')

describe('validate.macros', function()
    describe('top-level structure', function()
        it('rejects nil', function()
            local ok, err = validate.macros(nil)
            assert.is_false(ok)
            assert.matches('table', err)
        end)

        it('rejects a string', function()
            local ok, err = validate.macros('hello')
            assert.is_false(ok)
            assert.matches('table', err)
        end)

        it('rejects a number', function()
            local ok, err = validate.macros(42)
            assert.is_false(ok)
            assert.matches('table', err)
        end)

        it('rejects missing books key', function()
            local ok, err = validate.macros({})
            assert.is_false(ok)
            assert.matches('books', err)
        end)

        it('rejects non-table books', function()
            local ok, err = validate.macros({ books = 'nope' })
            assert.is_false(ok)
            assert.matches('books', err)
        end)

        it('accepts empty books table', function()
            local ok, err = validate.macros({ books = {} })
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts version 1', function()
            local ok, err = validate.macros({ version = 1, books = {} })
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects unsupported version', function()
            local ok, err = validate.macros({ version = 2, books = {} })
            assert.is_false(ok)
            assert.matches('version', err)
        end)
    end)

    describe('book constraints', function()
        local function book_data(idx, overrides)
            overrides = overrides or {}
            local book = { name = overrides.name, sets = overrides.sets or {} }
            if overrides.drop_sets then
                book.sets = nil
            end
            return { books = { [idx] = book } }
        end

        it('accepts a valid book at index 1', function()
            local ok, err = validate.macros(book_data(1))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts a valid book at max index 40', function()
            local ok, err = validate.macros(book_data(40))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects book index 0', function()
            local ok, err = validate.macros(book_data(0))
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects negative book index', function()
            local ok, err = validate.macros(book_data(-1))
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects book index 41 (out of range)', function()
            local ok, err = validate.macros(book_data(41))
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects non-numeric book index', function()
            local ok, err = validate.macros({ books = { bad = { sets = {} } } })
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('accepts book name exactly 15 chars', function()
            local ok, err = validate.macros(book_data(1, { name = 'ABCDE12345ABCDE' }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects book name longer than 15 chars', function()
            local ok, err = validate.macros(book_data(1, { name = 'ABCDE12345ABCDEF' }))
            assert.is_false(ok)
            assert.matches('name too long', err)
        end)

        it('rejects book name with spaces', function()
            local ok, err = validate.macros(book_data(1, { name = 'My Book' }))
            assert.is_false(ok)
            assert.matches('alphanumeric', err)
        end)

        it('rejects book name with symbols', function()
            local ok, err = validate.macros(book_data(1, { name = 'rdm75-nin' }))
            assert.is_false(ok)
            assert.matches('alphanumeric', err)
        end)

        it('accepts book name with mixed case alphanumeric', function()
            local ok, err = validate.macros(book_data(1, { name = 'WHM75NIN' }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts book with nil name', function()
            local ok, err = validate.macros({ books = { [1] = { sets = {} } } })
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects book missing sets key', function()
            local ok, err = validate.macros(book_data(1, { drop_sets = true }))
            assert.is_false(ok)
            assert.matches('sets', err)
        end)

        it('rejects a 41st book — index 41 is out of range', function()
            local data = { books = {} }
            for i = 1, 41 do
                data.books[i] = { sets = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('accepts exactly 40 books', function()
            local data = { books = {} }
            for i = 1, 40 do
                data.books[i] = { sets = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)

    describe('set constraints', function()
        local function set_data(set_idx, set_content)
            return {
                books = {
                    [1] = {
                        sets = { [set_idx] = set_content or { ctrl = {}, alt = {} } },
                    },
                },
            }
        end

        it('accepts valid ctrl and alt keys', function()
            local ok, err = validate.macros(set_data(1))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts set at max index 10', function()
            local ok, err = validate.macros(set_data(10))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects set index 0', function()
            local ok, err = validate.macros(set_data(0))
            assert.is_false(ok)
            assert.matches('invalid set index', err)
        end)

        it('rejects set index 11 (out of range)', function()
            local ok, err = validate.macros(set_data(11))
            assert.is_false(ok)
            assert.matches('invalid set index', err)
        end)

        it('accepts set with only ctrl', function()
            local ok, err = validate.macros(set_data(1, { ctrl = {} }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts set with only alt', function()
            local ok, err = validate.macros(set_data(1, { alt = {} }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects unknown modifier key', function()
            local ok, err = validate.macros(set_data(1, { shift = {} }))
            assert.is_false(ok)
            assert.matches('unknown modifier', err)
        end)

        it('rejects too many sets (11 per book)', function()
            local data = { books = { [1] = { sets = {} } } }
            for i = 1, 11 do
                data.books[1].sets[i] = { ctrl = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_false(ok)
            assert.matches('invalid set index', err)
        end)

        it('accepts exactly 10 sets', function()
            local data = { books = { [1] = { sets = {} } } }
            for i = 1, 10 do
                data.books[1].sets[i] = { ctrl = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)

    describe('macro constraints', function()
        local function macro_data(macros_list, modifier)
            modifier = modifier or 'ctrl'
            local slot = {}
            for i, m in ipairs(macros_list) do
                slot[i] = m
            end
            return { books = { [1] = { sets = { [1] = { [modifier] = slot } } } } }
        end

        it('accepts a valid macro', function()
            local ok, err = validate.macros(macro_data({ { name = 'Attack', contents = { '/ja "Provoke" <t>' } } }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts macro via alt modifier', function()
            local ok, err = validate.macros(macro_data({ { name = 'Heal', contents = {} } }, 'alt'))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects macro name longer than 8 chars', function()
            local ok, err = validate.macros(macro_data({ { name = 'TooLongNm', contents = {} } }))
            assert.is_false(ok)
            assert.matches('name too long', err)
        end)

        it('accepts macro name exactly 8 chars', function()
            local ok, err = validate.macros(macro_data({ { name = 'Exactly8', contents = {} } }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts macro with nil name', function()
            local ok, err = validate.macros(macro_data({ { contents = { '/say hi' } } }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects too many macros per modifier (11)', function()
            local list = {}
            for i = 1, 11 do
                list[i] = { name = 'M' .. i, contents = {} }
            end
            local ok, err = validate.macros(macro_data(list))
            assert.is_false(ok)
            assert.matches('too many macros', err)
        end)

        it('accepts exactly 10 macros per modifier', function()
            local list = {}
            for i = 1, 10 do
                list[i] = { name = 'Macro' .. i, contents = {} }
            end
            local ok, err = validate.macros(macro_data(list))
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)

    describe('line constraints', function()
        local function line_data(lines)
            return {
                books = {
                    [1] = {
                        sets = {
                            [1] = {
                                ctrl = { [1] = { name = 'Test', contents = lines } },
                            },
                        },
                    },
                },
            }
        end

        it('accepts up to 6 lines', function()
            local ok, err =
                validate.macros(line_data({ '/ja "Provoke" <t>', '/wait 1', '/p Done!', '/echo 4', '/echo 5', '/echo 6' }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects more than 6 lines', function()
            local ok, err = validate.macros(line_data({ '1', '2', '3', '4', '5', '6', '7' }))
            assert.is_false(ok)
            assert.matches('too many lines', err)
        end)

        it('rejects a non-string line', function()
            local ok, err = validate.macros(line_data({ 42 }))
            assert.is_false(ok)
            assert.matches('not a string', err)
        end)

        it('rejects a line exceeding 60 chars', function()
            local ok, err = validate.macros(line_data({ string.rep('a', 61) }))
            assert.is_false(ok)
            assert.matches('too long', err)
        end)

        it('accepts a line exactly 60 chars', function()
            local ok, err = validate.macros(line_data({ string.rep('a', 60) }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts macro with nil contents', function()
            local data = { books = { [1] = { sets = { [1] = { ctrl = { [1] = { name = 'T' } } } } } } }
            local ok, err = validate.macros(data)
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)

    describe('full valid structure', function()
        it('accepts a fully populated valid macro table', function()
            local data = {
                version = 1,
                character = 'squatched',
                books = {
                    [1] = {
                        name = 'WHM75NIN',
                        sets = {
                            [1] = {
                                ctrl = {
                                    [1] = { name = 'Cure', contents = { '/ma "Cure IV" <me>', '/wait 1' } },
                                    [2] = { name = 'Esuna', contents = { '/ma Esuna <me>' } },
                                },
                                alt = {
                                    [1] = { name = 'Protect', contents = { '/ma "Protect III" <me>' } },
                                },
                            },
                        },
                    },
                    [6] = {
                        name = 'RDM75NIN',
                        sets = {
                            [1] = { ctrl = {}, alt = {} },
                        },
                    },
                },
            }
            local ok, err = validate.macros(data)
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)
end)
