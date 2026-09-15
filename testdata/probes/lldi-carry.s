.name "lldi-carry"
.description "Expose lldi carry behavior as an in-code memory write"

        live %-1
        lldi %100, %100, r2
        zjmp %:taken

        ld %1, r3
        and r1, %0, r4
        zjmp %:write

taken:  ld %2, r3
write:  st r3, 5
        nop r1
