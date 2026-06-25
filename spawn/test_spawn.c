#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>

#include "helpers.h"

static int tests_run = 0, tests_failed = 0;

#define CHECK(cond, msg)                                                                           \
    do {                                                                                           \
        tests_run++;                                                                               \
        if (!(cond)) {                                                                             \
            fprintf(stderr, "FAIL: %s\n", (msg));                                                  \
            tests_failed++;                                                                        \
        }                                                                                          \
    } while (0)

#define CHECK_STR(got, want, msg)                                                                  \
    do {                                                                                           \
        tests_run++;                                                                               \
        if (strcmp((got), (want)) != 0) {                                                          \
            fprintf(stderr, "FAIL: %s\n  got  \"%s\"\n  want \"%s\"\n", (msg), (got), (want));     \
            tests_failed++;                                                                        \
        }                                                                                          \
    } while (0)

/* ── push_quoted ─────────────────────────────────────────────────────── */

static void test_push_quoted_simple(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "macromog.exe");
    CHECK_STR(buf, "macromog.exe", "simple arg — no quotes needed");
}

static void test_push_quoted_space(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "Program Files/foo.exe");
    CHECK_STR(buf, "\"Program Files/foo.exe\"", "arg with space — outer quotes added");
}

static void test_push_quoted_embedded_quote(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "say \"hi\"");
    CHECK_STR(buf, "\"say \\\"hi\\\"\"", "arg with embedded quotes — backslash-escaped");
}

static void test_push_quoted_multiple_args(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "cmd.exe");
    push_quoted(buf, &pos, sizeof(buf), "--output");
    push_quoted(buf, &pos, sizeof(buf), "json");
    CHECK_STR(buf, "cmd.exe --output json", "multiple args — space-separated");
}

static void test_push_quoted_second_arg_with_space(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "cmd.exe");
    push_quoted(buf, &pos, sizeof(buf), "foo bar");
    CHECK_STR(buf, "cmd.exe \"foo bar\"", "second arg with space — only that arg quoted");
}

static void test_push_quoted_empty_arg(void)
{
    char buf[256];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "cmd.exe");
    push_quoted(buf, &pos, sizeof(buf), "");
    /* empty arg has no special char — no quotes added */
    CHECK_STR(buf, "cmd.exe ", "empty arg — space separator only");
}

static void test_push_quoted_overflow(void)
{
    char buf[10];
    size_t pos = 0;
    buf[0] = '\0';
    push_quoted(buf, &pos, sizeof(buf), "12345678901234567890");
    CHECK(pos < sizeof(buf), "overflow — pos stays within capacity");
    /* Guard before indexing: CHECK does not halt execution on failure. */
    if (pos < sizeof(buf))
        CHECK(buf[pos] == '\0', "overflow — null terminator present");
}

/* ── mem_copy ────────────────────────────────────────────────────────── */

static void test_mem_copy_basic(void)
{
    char src[] = "hello";
    char dst[6] = {0};
    mem_copy(dst, src, 5);
    dst[5] = '\0';
    CHECK_STR(dst, "hello", "mem_copy — copies bytes correctly");
}

static void test_mem_copy_zero(void)
{
    char dst[] = "untouched";
    mem_copy(dst, "xyz", 0);
    CHECK_STR(dst, "untouched", "mem_copy — zero length leaves dst unchanged");
}

/* ── filetime_to_hex ─────────────────────────────────────────────────── */

static void test_filetime_to_hex_zeros(void)
{
    char stamp[17];
    filetime_to_hex(0, 0, stamp);
    CHECK_STR(stamp, "0000000000000000", "filetime_to_hex — all zeros");
}

static void test_filetime_to_hex_all_f(void)
{
    char stamp[17];
    filetime_to_hex(0xFFFFFFFF, 0xFFFFFFFF, stamp);
    CHECK_STR(stamp, "FFFFFFFFFFFFFFFF", "filetime_to_hex — all 0xF");
}

static void test_filetime_to_hex_known(void)
{
    char stamp[17];
    filetime_to_hex(0x01D8A0E3, 0xF1234567, stamp);
    CHECK_STR(stamp, "01D8A0E3F1234567", "filetime_to_hex — known value");
}

static void test_filetime_to_hex_split(void)
{
    char stamp[17];
    filetime_to_hex(0xDEADBEEF, 0x12345678, stamp);
    CHECK_STR(stamp, "DEADBEEF12345678", "filetime_to_hex — hi/lo boundary correct");
}

/* ── main ────────────────────────────────────────────────────────────── */

int main(void)
{
    test_push_quoted_simple();
    test_push_quoted_space();
    test_push_quoted_embedded_quote();
    test_push_quoted_multiple_args();
    test_push_quoted_second_arg_with_space();
    test_push_quoted_empty_arg();
    test_push_quoted_overflow();
    test_mem_copy_basic();
    test_mem_copy_zero();
    test_filetime_to_hex_zeros();
    test_filetime_to_hex_all_f();
    test_filetime_to_hex_known();
    test_filetime_to_hex_split();

    if (tests_failed == 0) {
        printf("%d/%d passed\n", tests_run, tests_run);
        return 0;
    }
    printf("%d/%d passed, %d FAILED\n", tests_run - tests_failed, tests_run, tests_failed);
    return 1;
}
