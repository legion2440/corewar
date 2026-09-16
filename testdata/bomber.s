.name "bomber"
.description "Visualizer demo: bombs both sides of its code and invades the opponent's copy"

        ld %128,r2
        ld %-128,r3
        ld %4,r4
        ld %-4,r5

        sti r1,%:seed_live,%1
seed_live:
        live %0

        and r1,%0,r6
        fork %:bomb_loop
        # Two-player spacing is 2048 bytes; +3 lands at bomb_loop after this lfork.
        lfork %2051

bomb_loop:
        sti r1,%:live,%1
live:
        live %0

        sti r1,r2,%0
        sti r1,r3,%0
        add r2,r4,r2
        add r3,r5,r3

        xor r2,%508,r6
        zjmp %:reset
        and r1,%0,r6
        zjmp %:bomb_loop

reset:
        ld %128,r2
        ld %-128,r3
        and r1,%0,r6
        zjmp %:bomb_loop
