#pragma once
#include <stddef.h>
#include <stdint.h>

/* Append src as a properly quoted cmdline token into dst[0..cap). */
static inline void push_quoted(char *dst, size_t *pos, size_t cap, const char *src)
{
    const char *p;
    int quote = 0;
    for (p = src; *p; p++)
        if (*p == ' ' || *p == '"') {
            quote = 1;
            break;
        }

    if (*pos > 0 && *pos < cap - 1)
        dst[(*pos)++] = ' ';
    if (quote && *pos < cap - 1)
        dst[(*pos)++] = '"';
    for (p = src; *p && *pos < cap - 2; p++) {
        if (*p == '"')
            dst[(*pos)++] = '\\';
        dst[(*pos)++] = *p;
    }
    if (quote && *pos < cap - 1)
        dst[(*pos)++] = '"';
    dst[*pos] = '\0';
}

/* Byte-by-byte copy — no CRT dependency. */
static inline void mem_copy(void *dst, const void *src, size_t n)
{
    char *d = (char *)dst;
    const char *s = (const char *)src;
    while (n--)
        *d++ = *s++;
}

/* Encode two 32-bit FILETIME halves as a 16-char uppercase hex string. */
static inline void filetime_to_hex(uint32_t hi, uint32_t lo, char out[17])
{
    static const char hex[] = "0123456789ABCDEF";
    int j;
    for (j = 0; j < 8; j++) {
        out[j] = hex[(hi >> (28 - 4 * j)) & 0xF];
        out[8 + j] = hex[(lo >> (28 - 4 * j)) & 0xF];
    }
    out[16] = '\0';
}
