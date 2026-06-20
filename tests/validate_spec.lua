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

        it('accepts a valid book at index 0', function()
            local ok, err = book_data(0)
            ok, err = validate.macros(book_data(0))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts a valid book at max index 39', function()
            local ok, err = validate.macros(book_data(39))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects negative book index', function()
            local ok, err = validate.macros(book_data(-1))
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects book index 40 (out of range)', function()
            local ok, err = validate.macros(book_data(40))
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects non-numeric book index', function()
            local ok, err = validate.macros({ books = { bad = { sets = {} } } })
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('rejects book name longer than 8 chars', function()
            local ok, err = validate.macros(book_data(0, { name = 'TooLongNm' }))
            assert.is_false(ok)
            assert.matches('name too long', err)
        end)

        it('accepts book name exactly 8 chars', function()
            local ok, err = validate.macros(book_data(0, { name = 'Exactly8' }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts book with nil name', function()
            local ok, err = validate.macros({ books = { [0] = { sets = {} } } })
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects book missing sets key', function()
            local ok, err = validate.macros(book_data(0, { drop_sets = true }))
            assert.is_false(ok)
            assert.matches('sets', err)
        end)

        it('rejects a 41st book — index 40 is out of range', function()
            -- LIMITS.books = 40, valid indices are 0-39; there is no valid index
            -- for a 41st book, so the index check always fires before the count check.
            local data = { books = {} }
            for i = 0, 40 do
                data.books[i] = { sets = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_false(ok)
            assert.matches('invalid book index', err)
        end)

        it('accepts exactly 40 books', function()
            local data = { books = {} }
            for i = 0, 39 do
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
                    [0] = {
                        sets = { [set_idx] = set_content or { ctrl = {}, alt = {} } },
                    },
                },
            }
        end

        it('accepts valid ctrl and alt keys', function()
            local ok, err = validate.macros(set_data(0))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts set with only ctrl', function()
            local ok, err = validate.macros(set_data(0, { ctrl = {} }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts set with only alt', function()
            local ok, err = validate.macros(set_data(0, { alt = {} }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('rejects unknown modifier key', function()
            local ok, err = validate.macros(set_data(0, { shift = {} }))
            assert.is_false(ok)
            assert.matches('unknown modifier', err)
        end)

        it('rejects too many sets (11 per book)', function()
            local data = { books = { [0] = { sets = {} } } }
            for i = 0, 10 do
                data.books[0].sets[i] = { ctrl = {} }
            end
            local ok, err = validate.macros(data)
            assert.is_false(ok)
            assert.matches('too many sets', err)
        end)

        it('accepts exactly 10 sets', function()
            local data = { books = { [0] = { sets = {} } } }
            for i = 0, 9 do
                data.books[0].sets[i] = { ctrl = {} }
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
            return { books = { [0] = { sets = { [0] = { [modifier] = slot } } } } }
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

        it('rejects too many macros (21)', function()
            local list = {}
            for i = 1, 21 do
                list[i] = { name = 'M' .. i, contents = {} }
            end
            local ok, err = validate.macros(macro_data(list))
            assert.is_false(ok)
            assert.matches('too many macros', err)
        end)

        it('accepts exactly 20 macros', function()
            local list = {}
            for i = 1, 20 do
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
                    [0] = {
                        sets = {
                            [0] = {
                                ctrl = { [1] = { name = 'Test', contents = lines } },
                            },
                        },
                    },
                },
            }
        end

        it('accepts up to 6 lines', function()
            local ok, err = validate.macros(line_data({ '/ja "Provoke" <t>', '/wait 1', '/p Done!', '/echo 4', '/echo 5', '/echo 6' }))
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

        it('rejects a line exceeding 255 chars', function()
            local ok, err = validate.macros(line_data({ string.rep('a', 256) }))
            assert.is_false(ok)
            assert.matches('too long', err)
        end)

        it('accepts a line exactly 255 chars', function()
            local ok, err = validate.macros(line_data({ string.rep('a', 255) }))
            assert.is_true(ok)
            assert.is_nil(err)
        end)

        it('accepts macro with nil contents', function()
            local data = { books = { [0] = { sets = { [0] = { ctrl = { [1] = { name = 'T' } } } } } } }
            local ok, err = validate.macros(data)
            assert.is_true(ok)
            assert.is_nil(err)
        end)
    end)

    describe('full valid structure', function()
        it('accepts a fully populated valid macro table', function()
            local data = {
                books = {
                    [0] = {
                        name = 'Combat',
                        sets = {
                            [0] = {
                                ctrl = {
                                    [1] = { name = 'Provoke', contents = { '/ja "Provoke" <t>', '/wait 1', '/p Provoking!' } },
                                    [2] = { name = 'Pull', contents = { '/ra <t>' } },
                                },
                                alt = {
                                    [1] = { name = 'Dia', contents = { '/ma "Dia" <t>' } },
                                },
                            },
                        },
                    },
                    [1] = {
                        name = 'Support',
                        sets = {
                            [0] = { ctrl = {}, alt = {} },
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
