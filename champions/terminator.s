.name "terminator"
.description "Hijacks ameba's one-time self patch so every future live credits this champion"

        sti r1,%:live,%1
        and r1,%0,r2
live:   live %0
        lfork %2028
        zjmp %-8
