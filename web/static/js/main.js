const currState = {
    player: 0,
    board: [
              [9],
             [9, 9],
           [9, 9, 9],
          [9, 9, 9, 9],
         [9, 9, 9, 9, 9],
        [9, 9, 9, 9, 9, 9]
    ].reverse(),
    unusedPawns: [3, 3, 3],
    scores: [0, 0, 0],
    forceMovePawn: [9, 9]
};
const TARGET_SCORE = 60;

let gameOver = false;
let botPlayers = [];

let validMoves = [
    {"Player":0,"Board":0,"Path":[[9,9],[0,0]]},
    {"Player":0,"Board":0,"Path":[[9,9],[0,1]]},
    {"Player":0,"Board":0,"Path":[[9,9],[0,2]]},
    {"Player":0,"Board":0,"Path":[[9,9],[0,3]]},
    {"Player":0,"Board":0,"Path":[[9,9],[0,4]]},
    {"Player":0,"Board":0,"Path":[[9,9],[0,5]]}
];


$('#bot1, #bot2, #bot3').on("change", function() {
    botPlayers = [];
    $.each($('#bot1, #bot2, #bot3'), function(_, elem) {
        if ($(elem).is(':checked')) {
            botPlayers.push(parseInt($(elem).val()));
        }
    });
});

function repaintBoard() {
    // Given the current state of the game, repaint the whole thing

}

function updatePointsScorable() {
    for (let p = 0; p < 3; p++) {
        const points = topPawnScorePointsValue(currState.board, p);
        $(`.player-${p}-score .player-${p}-start .cell-inner`)[0].innerHTML = points;
        $(`.player-${p}-score span`)[0].innerHTML = currState.scores[p];
        const scorable = $(`.player-${p}-score .player-${p}-start`).removeClass("scorable");
        scorable.removeClass("scorable");
        $(`.player-${p}-target`).removeClass(`player-${p}-target`);
        if ((points > 0) && (points+currState.scores[p] <= TARGET_SCORE) && (currState.forceMovePawn[0] === 9)) {
            scorable.addClass("scorable");
        }
    }
    $(".active").removeClass("active");
}

function topPawnScorePointsValue(board, player) {
    for (let y = board.length-1; y >= 0; y--) {
        for (let x = 0; x < board.length; x++) {
            if (board[y][x] === player) {
                return (y + 1)
            }
        }
    }
    return 0
}

function availableScorePoints(board, player) {
    for (let y = board.length-1; y >= 0; y--) {
        if (y + currState.scores[player] <= TARGET_SCORE && (currState.forceMovePawn[0] === 9)) {
            for (let x = 0; x < board.length; x++) {
                if (board[y][x] === player) {
                    return (y + 1)
                }
            }
        }
    }
    return 0
}

function getCoords(elem) {
    const classes = $(elem).attr("class").split(/\s+/);
    let currY = 9;
    let currX = 9;
    $.each(classes, function(_, v) {
        if (v.charAt(0) === 'y') {
            currY = parseInt(v.substring(1));
        }
        if (v.charAt(0) === 'x') {
            currX = parseInt(v.substring(1));
        }
    });
    return [currY, currX];
}

function getDelta(current, next) {
    if (current < next) {
        return 1;
    } else if (current > next) {
        return -1;
    }
    return 0;
}

function refreshValidMoves(next) {
    if (next[0] === 9) {
        currState.player = ++currState.player % 3;
    }
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currState.player}-score .player-${currState.player}-start`).addClass("curr-turn");
    //currState.forceMovePawn = next;
    updatePointsScorable();
    $.ajax({
        url: "/availableMoves",
        type: "POST",
        data: JSON.stringify(currState),
        contentType: "application/json; charset=utf-8",
        success: function(data) {
            if (data.length === 1 && data[0].Path[0][0] !== 9 && (data[0].Path[0] === data[0].Path[2] && data[0].Path[1] === data[0].Path[3])) {
                console.log("Here's a situation (we were only given one choice!) ...");
                refreshValidMoves([9, 9]); // since [0] is 9 here, we're not passing on the tour to the next player
            } else {
                validMoves = data;
                if (next[0] !== 9) {
                    setActive($(`.y${currState.forceMovePawn[0]}.x${currState.forceMovePawn[1]}`)[0]);
                }
            }
        },
        error: function(_1, _2, _3) {
            console.log("Received a 500 error from server. Calling it game over.");
            gameOver = true;
        }
    });
}

const playerColors = {0: "Red", 1: "Green", 2: "Blue"};
function clickScore() {
    clearPrevs();
    currState.scores[currState.player] += availableScorePoints(currState.board, currState.player);
    if (currState.scores[currState.player] === TARGET_SCORE) {
        alert(`Yay, ${playerColors[currState.player]} wins!`);
        gameOver = true;
    }
    $(".curr-turn").addClass(`player-${currState.player}-prev`);
    refreshValidMoves([9, 9]);
}

function clickMove(elem) {
    const next = getCoords($(elem));
    const active = $(".active");
    let curr = getCoords(active[0]);

    $(`.y${next[0]}.x${next[1]}`).addClass(`player-${currState.player}-prev`);
    if (curr[0] === -1) {
        alert("I don't get it. Kristaps, what's happening here?");
        curr = next;
    }
    if (curr[0] === 9) {
        currState.unusedPawns[currState.player]--;
        currState.board[next[0]][next[1]] = currState.player;
        $(elem).addClass(`player-${currState.player}`);
    } else {
        currState.board[curr[0]][curr[1]] = 9; // removed from current position
        if (next[0] < 0 || next[0] >= currState.board.length ||
            next[1] < 0 || next[1] >= currState.board[next[0]].length) {
            // New position is in limbo
            // Remove pawn from curren state's board:
            currState.unusedPawns[currState.player]++;
            // Add a pawn to the bottom limbo line of unused pawns:
            $(`.player-${currState.player}-start:not(.player-${currState.player}):last`).addClass(`player-${currState.player}`);
        } else {
            // New position is on board
            currState.board[next[0]][next[1]] = currState.player;
            $(elem).addClass(`player-${currState.player}`);
        }
        // New positions are now occupied. Time to remove a pawn, if we jumped over one:
        if (Math.abs(curr[0]-next[0]) > 1 || Math.abs(curr[1]-next[1]) > 1) {
            const temp = [curr[0] + getDelta(curr[0], next[0]), curr[1] + getDelta(curr[1], next[1])];
            const goner = currState.board[temp[0]][temp[1]];
            if (goner !== currState.player) {
                currState.unusedPawns[goner]++;
                const prev = $(`.y${temp[0]}.x${temp[1]}`);
                prev.addClass(`player-${goner}-prev`);
                $(`.player-${goner}-start:not(.player-${goner}):last`).addClass(`player-${goner}`).addClass(`player-${goner}-prev`);
                currState.board[temp[0]][temp[1]] = 9;
                prev.removeClass('player-0').removeClass('player-1').removeClass('player-2');
            }
        }
    }
    active.removeClass(`player-${currState.player}`);
    active.addClass(`player-${currState.player}-prev`);
}

function doBotMove() {
    $.ajax({
        url: "/suggestBotMove",
        type: "POST",
        data: JSON.stringify(currState),
        contentType: "application/json; charset=utf-8",
        success: function(move) {
            if (move.Board === 1) {
                clickScore();
                return;
            }

            const path = move.Path;
            if (path[0] === path[1]) {
                // a bot is allowed to stop multi-moves early
                console.log(`${currState.player} - stop early - THIS CAN NO LONGER OCCUR! HELP! WHAT HAPPENED?`);
                refreshValidMoves([9, 9]);
            } else {
                if (path[0][0] === 9 && path.length !== 2) {
                    console.error("When entering the board, one should not be able to move further!")
                }

                function setActiveAndCallClickMove(index) {
                    if (path[index][0] === 9) {
                        // Moving a pawn into the board
                        $(`.player-${currState.player}-start.player-${currState.player}:last`).addClass("active");
                    } else {
                        $(`.y${path[index][0]}.x${path[index][1]}`).addClass("active");
                    }
                    clickMove($(`.y${path[index+1][0]}.x${path[index+1][1]}`));
                }

                for (let i = 0; i < path.length-1; i++) {
                    console.log(`${currState.player} - no of moves in move: ${path.length-1} (and now we're at #${i+1})`);
                    setActiveAndCallClickMove(i);
                    //console.log(`  ... and after moving we have these pawns in the game: ${}`);
                }
                refreshValidMoves([9, 9]);
                console.log(`${currState.player} - turn taken. Board is now thus:`);
                console.dir(currState.board);
            }
        },
        error: function(_, _, _) {
            gameOver = true;
        }
    });
}

function clearPrevs() {
    $(".cell").removeClass("player-0-prev").removeClass("player-1-prev").removeClass("player-2-prev");
}

function setActive(elem) {
    if (currState.forceMovePawn[0] === 9) {
        clearPrevs();
    }
    const classes = $(elem).attr("class").split(/\s+/);
    if (classes.includes(`player-${currState.player}`)) {
        $(".active").removeClass("active");
        $(".player-0-target").removeClass("player-0-target");
        $(".player-1-target").removeClass("player-1-target");
        $(".player-2-target").removeClass("player-2-target");
        $(elem).addClass("active");
        if (classes.includes(`player-${currState.player}-start`)) {
            $.each(validMoves, function(_, move) {
                const path = move.Path;
                if (path[0][0] === 9 && path[1][0] !== 9) {
                    // This is for coming onto the board. If first path element is 9,9, the pawn us unused
                    $(`.y${path[1][0]}.x${path[1][1]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currState.player}-target`);
                }
            });
        } else {
            const current = getCoords($(elem));
            $.each(validMoves, function(_, move) {
                const path = move.Path;
                if (path[0][0] === current[0] && path[0][1] === current[1]) {
                    // This path starts from the current cell, so concerns the current pawn
                    if (path.length > 2) {
                        console.log("Kristaps, we need to colorize all path elements here, not just te first jump target.")
                        console.log("We also need to make the newly welcomed path elements clickable, and we should be able,")
                        console.log("somehow, to distinguish the different paths, since it will be possible to hit the same")
                        console.log("final cell via different paths ...")
                    }
                    $(`.y${path[1][0]}.x${path[1][1]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currState.player}-target`);
                }
            });
        }
    }
}

const body = $('body');
body.on('click', ".player-0, .player-1, .player-2", function () {
    if (!botPlayers.includes(currState.player)) {
        if (currState.forceMovePawn[0] === 9) {
            setActive($(this));
        } else {
            const classes = $(this).attr("class").split(/\s+/);
            if (classes.includes("active")) {
                refreshValidMoves([9, 9]);
            }
        }
    }
});

body.on('click', '.player-0-target, .player-1-target, .player-2-target', function () {
    if (!botPlayers.includes(currState.player)) {
        clickMove($(this));
        refreshValidMoves([9, 9]);
    } else {
        alert("Please do not interfere while the bot is thinking!")
    }
});

body.on('click', '.curr-turn.scorable', function () {
    if (!botPlayers.includes(currState.player)) {
        clickScore();
    }
});

function botLoop() {
    setTimeout(function() {
        if (!botPlayers.includes(currState.player)) {
            console.log('wait for hooman');
            botLoop();
        } else {
            doBotMove();
            if (!gameOver) {
                botLoop();
            }
        }
    }, 1000)
}

botLoop();