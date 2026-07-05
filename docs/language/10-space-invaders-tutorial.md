# How to Program a Space Invaders Game

This tutorial explains how to build a small Space Invaders-style game in
Peddle. The finished reference program is:

```text
examples/space_invaders.ped
```

The goal is not only to copy that program. The goal is to show the game
structure that you can reuse for other C64 games:

- fixed screen layout
- game state stored in arrays
- non-blocking input
- joystick and keyboard control
- independent timers for movement
- collision checks
- score and lives
- sound effects
- clear game-over and win states

Each step builds on the previous one.

---

# Step 1: Start With the Screen Plan

The C64 text screen is 40 columns by 25 rows. A simple action game becomes much
easier when you reserve rows for specific jobs.

The reference Space Invaders game uses this layout:

```text
row 0       score and lives
row 1       controls or status text
rows 2-21   alien field and bullets
row 23      player cannon
row 24      key guide
```

Define the important constants first:

```peddle
const COLOR_WHITE = 1

const PLAYER_Y = 23
const ALIEN_ROWS = 4
const ALIEN_COLS = 8
const ALIEN_COUNT = 32

const RESULT_NONE = 0
const RESULT_GAME_OVER = 1
const RESULT_LANDED = 2
const RESULT_WIN = 3
```

For small games, it is often easiest to put runtime state inside `main()` as
local variables and pass arrays to helper functions. Shared state can also be
declared as top-level `var` globals when several functions need to mutate it.
Declare all function locals at the beginning of the function body.

```peddle
fn main() {
    var score int
    var lives byte
    var playerX byte
    var gameOver bool
    var gameResult byte

    score = 0
    lives = 3
    playerX = 19
    gameOver = false
    gameResult = RESULT_NONE
}
```

This shape is deliberate: declarations first, initialization second, game loop
third.

---

# Step 2: Draw Directly to Screen RAM

Use direct screen builtins for games. They do not move the KERNAL text cursor,
so they are predictable inside animation loops.

```peddle
fn drawChar(x byte, y byte, ch char) {
    putchar(x, y, ch)
    putcolor(x, y, COLOR_WHITE)
}
```

The HUD is just direct text plus decimal conversion:

```peddle
fn showHUD(score int, lives byte) {
    var nums char[8]

    putstr(0, 0, "SCORE:")
    clear(nums)
    copy(nums, itoa(score))
    putstr(6, 0, "       ")
    putstr(6, 0, nums)

    putstr(29, 0, "LIVES:")
    clear(nums)
    copy(nums, itoa(lives))
    putstr(35, 0, "  ")
    putstr(35, 0, nums)
}
```

Initialize the screen once before the game begins:

```peddle
cls()
border(6)
background(0)
textcolor(1)
showHUD(score, lives)
putstr(0, 24, "A/D OR JOY:MOVE  SPACE/FIRE:SHOOT  Q:QUIT")
```

The reference game uses simple characters:

```text
A  player cannon
I  player bullet
!  alien bullet
W  top-row alien
M  second-row alien
V  third-row alien
T  fourth-row alien
```

You can replace these later with custom characters or sprites. The game logic
does not need to change.

---

# Step 3: Store Aliens in Parallel Arrays

The alien grid needs three pieces of data per alien:

- alive or destroyed
- x position
- y position

Peddle arrays work well for this:

```peddle
var alienAlive bool[32]
var alienX byte[32]
var alienY byte[32]
```

Initialize the grid row by row:

```peddle
fn initAliens(alive bool[32], ax byte[32], ay byte[32]) {
    var i byte
    var row byte
    var col byte

    i = 0
    row = 0
    while row < 4 {
        col = 0
        while col < 8 {
            alive[i] = true
            ax[i] = col * 4 + 2
            ay[i] = row * 2 + 2
            i = i + 1
            col = col + 1
        }
        row = row + 1
    }
}
```

Draw only aliens that are alive:

```peddle
fn drawAliens(alive bool[32], ax byte[32], ay byte[32]) {
    var i byte
    var row byte
    var ch char

    i = 0
    while i < 32 {
        row = i / 8
        if alive[i] {
            ch = 'T'
            if row == 0 { ch = 'W' }
            if row == 1 { ch = 'M' }
            if row == 2 { ch = 'V' }
            drawChar(ax[i], ay[i], ch)
        }
        i = i + 1
    }
}
```

This parallel-array style is common in small C64 games. It avoids dynamic
allocation and keeps the memory layout obvious.

---

# Step 4: Read Keyboard and Joystick Without Blocking

Games should not call `waitkey()` inside the main loop. Use `key()` for
non-blocking keyboard input and `joy(port)` for joystick input.

Keyboard:

```peddle
k = key()
```

Joystick:

```peddle
j = joy(2) & 31
```

The joystick bits are active-low. A direction is pressed when its bit is `0`.

Common checks:

```peddle
if (j & 4) == 0 {
    # left
}

if (j & 8) == 0 {
    # right
}

if (j & 16) == 0 {
    # fire
}
```

Declare input flags at the beginning of `main()`:

```peddle
var k char
var j byte
var moveLeft bool
var moveRight bool
var firePressed bool
```

Then reset and fill those flags each time through the game loop:

```peddle
moveLeft = false
moveRight = false
firePressed = false

k = key()
j = joy(2) & 31

if k == 'q' { gameOver = true }
if k == 'Q' { gameOver = true }

if k == 'a' { moveLeft = true }
if k == 'A' { moveLeft = true }
if (j & 4) == 0 { moveLeft = true }

if k == 'd' { moveRight = true }
if k == 'D' { moveRight = true }
if (j & 8) == 0 { moveRight = true }

if k == ' ' { firePressed = true }
if (j & 16) == 0 { firePressed = true }
```

Then apply the input to the player:

```peddle
if moveLeft {
    if playerX > 0 {
        oldPlayerX = playerX
        playerX = playerX - 1
        drawChar(playerX, PLAYER_Y, 'A')
        putchar(oldPlayerX, PLAYER_Y, ' ')
    }
}

if moveRight {
    if playerX < 39 {
        oldPlayerX = playerX
        playerX = playerX + 1
        drawChar(playerX, PLAYER_Y, 'A')
        putchar(oldPlayerX, PLAYER_Y, ' ')
    }
}
```

The current reference program uses `A`, `D`, space, and `Q`. Adding joystick
support follows this exact pattern.

---

# Step 5: Use Timers Instead of Delay Loops

Action games need several things moving at different speeds:

- player input every frame
- bullets every few ticks
- aliens more slowly
- alien shots even more slowly

Use `ticks()` and `tickdue(last, interval)` for this.

```peddle
var bulletTimer int
var alienTimer int
var fireTimer int

bulletTimer = ticks()
alienTimer = ticks()
fireTimer = ticks()
```

Inside the main loop:

```peddle
if tickdue(bulletTimer, 4) {
    bulletTimer = ticks()
    # move bullets
}

if tickdue(alienTimer, alienMoveInterval) {
    alienTimer = ticks()
    # move aliens
}

if tickdue(fireTimer, 40) {
    fireTimer = ticks()
    # create alien bullet
}
```

This keeps the game responsive because the loop never sleeps. It only performs
work when a timer is due.

---

# Step 6: Add the Player Bullet

Use one active player bullet at a time:

```peddle
var pbActive bool
var pbX byte
var pbY byte
var oldPbX byte
var oldPbY byte

pbActive = false
```

Fire only when no player bullet is already active:

```peddle
if firePressed {
    if pbActive == false {
        pbX = playerX
        pbY = 22
        pbActive = true
        drawChar(pbX, pbY, 'I')
    }
}
```

Move it upward on the bullet timer:

```peddle
if pbActive {
    oldPbX = pbX
    oldPbY = pbY

    if pbY <= 2 {
        pbActive = false
        putchar(oldPbX, oldPbY, ' ')
    } else {
        pbY = pbY - 1
        drawChar(pbX, pbY, 'I')
        putchar(oldPbX, oldPbY, ' ')
    }
}
```

Keep movement and drawing close together. That makes it easier to avoid leaving
old characters on the screen.

---

# Step 7: Check Bullet and Alien Collision

After moving the player bullet, compare it with each alive alien.

```peddle
i = 0
while i < 32 {
    if alienAlive[i] {
        if alienX[i] == pbX {
            if alienY[i] == pbY {
                alienAlive[i] = false
                pbActive = false
                putchar(pbX, pbY, ' ')
                putchar(oldPbX, oldPbY, ' ')
                # add score here
                break
            }
        }
    }
    i = i + 1
}
```

The reference game scores aliens by row:

```peddle
row = i / 8
pts = 10
if row == 0 { pts = 30 }
if row == 1 { pts = 20 }
score = score + pts
showHUD(score, lives)
```

After a hit, reduce the alien movement interval as fewer aliens remain:

```peddle
aliveCount = countAlive(alienAlive)
if aliveCount < 24 { alienMoveInterval = 16 }
if aliveCount < 16 { alienMoveInterval = 12 }
if aliveCount < 8  { alienMoveInterval = 8  }
if aliveCount < 4  { alienMoveInterval = 5  }
```

This is an important game-feel trick. Difficulty rises naturally as the player
gets closer to winning.

---

# Step 8: Move the Alien Formation

The alien formation has one shared direction:

```peddle
var alienGoRight bool
var alienMoveInterval int
var hitWall bool

alienGoRight = true
alienMoveInterval = 20
```

Before moving, check if any alive alien would hit a side wall:

```peddle
hitWall = false
i = 0
while i < 32 {
    if alienAlive[i] {
        if alienGoRight {
            if alienX[i] >= 38 { hitWall = true }
        } else {
            if alienX[i] <= 1 { hitWall = true }
        }
    }
    i = i + 1
}
```

If a wall is hit, reverse direction and move down one row:

```peddle
if hitWall {
    alienGoRight = !alienGoRight
    i = 0
    while i < 32 {
        if alienAlive[i] {
            alienY[i] = alienY[i] + 1
        }
        i = i + 1
    }
}
```

Otherwise, move horizontally:

```peddle
i = 0
while i < 32 {
    if alienAlive[i] {
        if alienGoRight {
            alienX[i] = alienX[i] + 1
        } else {
            alienX[i] = alienX[i] - 1
        }
    }
    i = i + 1
}
```

Erase old positions after drawing the new formation. The reference game copies
the old alien arrays before movement, then erases old positions afterward.

```peddle
copyAlienPositions(alienX, alienY, oldAlienX, oldAlienY)
# move aliens
drawAliens(alienAlive, alienX, alienY)
eraseAliens(alienAlive, oldAlienX, oldAlienY)
```

This avoids flicker from clearing the entire screen.

---

# Step 9: Add Alien Bullets

The reference game allows three alien bullets at once:

```peddle
var abActive bool[3]
var abX byte[3]
var abY byte[3]
```

Initialize them:

```peddle
i = 0
while i < 3 {
    abActive[i] = false
    abX[i] = 0
    abY[i] = 0
    i = i + 1
}
```

Use a round-robin search so different aliens get a chance to fire:

```peddle
var fireSearch byte
var fireSlot byte
var gotSlot bool
var searched byte

fireSearch = 0
```

On the fire timer, find a free slot:

```peddle
gotSlot = false
i = 0
while i < 3 {
    if gotSlot == false {
        if abActive[i] == false {
            fireSlot = i
            gotSlot = true
        }
    }
    i = i + 1
}
```

Then find the next alive alien:

```peddle
searched = 0
while searched < 32 {
    if alienAlive[fireSearch] {
        abX[fireSlot] = alienX[fireSearch]
        abY[fireSlot] = alienY[fireSearch] + 1
        abActive[fireSlot] = true
        drawChar(abX[fireSlot], abY[fireSlot], '!')
        break
    }

    fireSearch = fireSearch + 1
    if fireSearch >= 32 { fireSearch = 0 }
    searched = searched + 1
}
```

Move alien bullets downward on the bullet timer. If a bullet reaches the player
row at the player x position, remove a life.

```peddle
if abX[i] == playerX {
    if abY[i] == PLAYER_Y {
        abActive[i] = false
        lives = lives - 1
        showHUD(score, lives)
        if lives == 0 {
            gameOver = true
            gameResult = RESULT_GAME_OVER
        }
    }
}
```

This is the same state pattern as the player bullet, just repeated in arrays.

---

# Step 10: Add Sound Effects

Peddle sound uses a byte stream and a sound pool. The reference game builds two
effects:

- rocket launch on voice 1
- explosion noise on voice 2

Define the sound constants in your program or import them from your own library:

```peddle
const SOUND_STREAM = 1
const SOUND_VOICE1 = 1
const SOUND_VOICE2 = 2
const SOUND_REPLACE = 1
const SOUND_OVERLAY = 2
const ERR_OK = 0

const VOICE1 = 0
const VOICE2 = 1
const WAVE_TRIANGLE = 16
const WAVE_NOISE = 128
const GATE = 1
```

Small helper functions make the stream readable:

```peddle
fn waveform(data byte[128], voice byte, value byte) {
    append(data, 4)
    append(data, voice)
    append(data, value)
}

fn soundWait(data byte[128], frames byte) {
    append(data, 1)
    append(data, frames)
}

fn env(data byte[128], voice byte, ad byte, sr byte) {
    append(data, 5)
    append(data, voice)
    append(data, ad)

    append(data, 6)
    append(data, voice)
    append(data, sr)
}

fn volume(data byte[128], value byte) {
    append(data, 7)
    append(data, value)
}

fn freq(data byte[128], voice byte, lo byte, hi byte) {
    append(data, 8)
    append(data, voice)
    append(data, lo)
    append(data, hi)
}
```

Build a short rocket sound:

```peddle
fn buildRocketSound(data byte[128]) {
    clear(data)

    volume(data, 15)
    env(data, VOICE1, 8, 128)
    freq(data, VOICE1, 103, 17)
    waveform(data, VOICE1, WAVE_TRIANGLE + GATE)
    soundWait(data, 4)
    waveform(data, VOICE1, WAVE_TRIANGLE)
    append(data, 0)
}
```

Load sounds during initialization:

```peddle
var soundPool byte[1024]
var rocketData byte[128]
var rocketId uint
var sErr int

sound_init(soundPool)
buildRocketSound(rocketData)
rocketId, sErr = sound_load(rocketData, SOUND_STREAM)
```

Play the effect when the player fires:

```peddle
if sErr == ERR_OK {
    playErr = sound_play(rocketId, SOUND_VOICE1, 4, SOUND_OVERLAY)
}
```

Use `SOUND_OVERLAY` for short effects so the explosion can play while the rocket
effect is still active. Use different voice masks for different effects.

---

# Step 11: End Conditions

There are three natural end conditions:

- lives reached zero
- aliens reached the bottom
- all aliens were destroyed

Use a result code so the end screen can show the right message:

```peddle
if lives == 0 {
    gameOver = true
    gameResult = RESULT_GAME_OVER
}

if aliensAtBottom(alienAlive, alienY) {
    gameOver = true
    gameResult = RESULT_LANDED
}

aliveCount = countAlive(alienAlive)
if aliveCount == 0 {
    gameOver = true
    gameResult = RESULT_WIN
}
```

After the loop, display the result:

```peddle
if gameResult == RESULT_GAME_OVER {
    putstr(11, 14, "GAME OVER")
}
if gameResult == RESULT_LANDED {
    putstr(10, 12, "THEY LANDED!")
}
if gameResult == RESULT_WIN {
    putstr(12, 12, "YOU WIN!")
}
putstr(10, 16, "PRESS ANY KEY")
waitkey()
```

Separating game logic from the final display keeps the main loop easier to
understand.

---

# Step 12: Main Loop Template

The final loop has a simple rhythm:

```peddle
while gameOver == false {
    # 1. read input
    k = key()
    j = joy(2) & 31

    # 2. apply player movement

    # 3. create player bullet if fire was pressed

    # 4. update player and alien bullets when bulletTimer is due

    # 5. update alien formation when alienTimer is due

    # 6. create alien bullet when fireTimer is due

    # 7. set gameOver and gameResult when an end condition is reached
}
```

This is the most useful template to carry into other games. Most C64 arcade
games can start with this structure.

---

# Step 13: Build and Run

Compile the reference game:

```sh
./peddle.sh examples/space_invaders.ped
```

Run it in VICE:

```sh
./peddle.sh --run examples/space_invaders.ped
```

The reference controls are:

```text
A / D  move left and right
SPACE  fire
Q      quit
```

To add joystick control, add the `joy(2)` input checks from Step 4 to the main
loop.

---

# Design Checklist

When building your own game from this template, decide these things first:

- Which rows belong to the HUD, playfield, and player?
- Which objects are single variables?
- Which objects need arrays?
- Which things move on separate timers?
- Which objects collide?
- Which events need sound?
- Which result states can end the game?

For Space Invaders, the answers are:

| Concept | Implementation |
|---|---|
| Player | `playerX`, fixed `PLAYER_Y` |
| Aliens | `alienAlive[32]`, `alienX[32]`, `alienY[32]` |
| Player bullet | one active bullet with `pbActive`, `pbX`, `pbY` |
| Alien bullets | three active slots with `abActive[3]`, `abX[3]`, `abY[3]` |
| Timing | `bulletTimer`, `alienTimer`, `fireTimer` |
| Difficulty | reduce `alienMoveInterval` as `countAlive()` falls |
| Sound | one voice for rocket, one voice for explosion |
| Game result | `RESULT_GAME_OVER`, `RESULT_LANDED`, `RESULT_WIN` |

---

# Next Improvements

Good next steps after the reference version:

- add joystick support directly to `examples/space_invaders.ped`
- add bunkers between player and aliens
- use custom character graphics for aliens
- add a start screen difficulty selector
- add a high score
- add a saucer bonus enemy
- move repeated sound helpers into an imported `lib/sound.ped`

Keep each improvement isolated. A good C64 game grows best as a series of small,
working changes.
