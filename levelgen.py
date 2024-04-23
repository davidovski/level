#!/usr/bin/env python
array = [
        "#########################",
        "#########################",
        "#########################",
        "#########################",
        "#########################",
        "#########################",
        "#########################",
        "                         ",
        "                         ",
        "                         ",
        " #####             ##### ",
        " #####     ##      ##### ",
        " #####     ##      ##### ",
        " #####     ##      ##### ",
        " #####     ##      ##### ",
        " #####     ##      ##### ",
        " #####     ##      ##### ",
        ]
#array = [
#        "                         ",
#        "                         ",
#        "                         ",
#        "                         ",
#        " ###################     ",
#        " ###################     ",
#        " ##                      ",
#        " ##                      ",
#        " ##                      ",
#        " ##                      ",
#        " ##                      ",
#        " ##                      ",
#        " ##                      ",
#        " ########        ####### ",
#        " ########       ######## ",
#        " ########       ######## ",
#        " ########       ######## ",
#        ]

def layer0(x, y, solid, topSurface, left, right):

    n = 0
    if solid and not topSurface:
        n = 37
        n += x % 3
        n += (y % 3)*16

    if topSurface and solid:
        n = 34
        if left:
            n += 16
        elif right:
            n += 32

    print ("%d, " % n, end="")

def layer1(x, y, solid, topSurface, left, right):

    n = 0
    if topSurface and solid:
        n = 82
        if left:
            n += 16
        elif right:
            n += 32

    print ("%d, " % n, end="")


def calc(x, y):

        solid = array[y][x] == "#"

        topSurface = solid and (y != 0 and array[y-1][x] == " ")
        left = x == 0 or array[y][x-1] == " "
        right = (x+1) == len(array[0]) or array[y][x+1] == " "
        return solid, topSurface, left, right

print("{")
for y in range(len(array)):
    for x in range(len(array[y])):
        solid, topSurface, left, right = calc(x, y)
        layer0(x, y, solid, topSurface, left, right)

    print("")
print("}, ")
print("{")
for y in range(len(array)):
    for x in range(len(array[y])):
        solid, topSurface, left, right = calc(x, y)
        layer1(x, y, solid, topSurface, left, right)

    print("")
print("}, ")
