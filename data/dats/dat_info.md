# `.dat` File Setups
I'm using the notation `B#S#` where B is the book (1-based) and the S is the set (1-based).

# Hypothesis
Each macro is a C++ struct that looks like this:

```cpp
struct Macro {
    char[6][61] lines; // 6 lines of 60 chars + null terminator
    char[9] name; // 8 + null terminator
}

struct MacroSet {
    Macro[10] ctrl; // [0] is Ctrl 1, [9] is Ctrl 10.
    Macro[10] alt;
}
```

Then each `MacroSet` is serialized to disk as `mcrBS.dat` where `B` is 0-based index of the book, and `S` is 0-based index of the MacroSet in the book. So for 40 macros, there can be `mcr.dat` which is B1S1, `mcr1.dat` is `B1S2`, etc. All the way through `mcr399.dat` which corresponds with `B40S10`.

# Evidence
My RDM book is book 6, I started adding sets there to test in set 9 and set 10. So `mcr58.dat` became `b6s9_pathological_macros.dat` and `mcr59.dat` became `b6s10_struct_test_macros.dat`. I also turned around and picked a book I've never touched, `Book33`. Then in the Ctrl1 macro for each one, I named it `B#S#` so `B33S1`, `B33S2`, etc.

# Gaps
The only thing I'm not sure of is where the macro book names are. But it's somewhere in one of the `.dat` files in `book_names/`. I changed the name of a book, then zoned, and those were all the files that were touched so it's in there somewhere. I named a book "jVE2M4P6MXKYPl0" so that string should be in those `.dat` files somewhere.

## Solution
`mcr.ttl` is the first 20 book titles, `mcr_2.ttl` is the second 20 books set. The first few bytes are consistent between the two files, then some random stuff, then what looks like a `char[20][16] titles;` or something (20 book titles of 15 chars each + null terminator).

I have since deleted the `book_names/` dir and just kept `mcr.ttl` and `mcr_2.ttl`.
