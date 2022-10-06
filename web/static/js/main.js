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

/**
 * For a given player, this function returns the number of moves
 * possible on the progress board given the current situation on
 * the strategy board.
 */
function availableProgressPoints(board, player) {
    for (let i = board.length-1; i >= 0; i--) {
        for (let j = 0; j < board.length; j++) {
            if (board[i][j] === player) {
                return (i + 1)
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
        if (v.charAt(0) === 'i') {
            currY = parseInt(v.substring(1));
        }
        if (v.charAt(0) === 'j') {
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

function updatePointsScorable() {
    for (let i = 0; i < 3; i++) {
        const points = availableProgressPoints(currState.board, i);
        $(`.player-${i}-score .player-${i}-start .cell-inner`)[0].innerHTML = points;
        $(`.player-${i}-score span`)[0].innerHTML = currState.scores[i];
        const scorable = $(`.player-${i}-score .player-${i}-start`).removeClass("scorable");
        scorable.removeClass("scorable");
        $(`.player-${i}-target`).removeClass(`player-${i}-target`);
        if ((points > 0) && (points+currState.scores[i] <= TARGET_SCORE) && (currState.forceMovePawn[0] === 9)) {
            scorable.addClass("scorable");
        }
    }
    $(".active").removeClass("active");
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
                    setActive($(`.i${currState.forceMovePawn[0]}.j${currState.forceMovePawn[1]}`)[0]);
                }
            }
        }
    });
}

const playerColors = {0: "Red", 1: "Green", 2: "Blue"};
function clickScore() {
    clearPrevs();
    currState.scores[currState.player] += availableProgressPoints(currState.board, currState.player);
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

    $(`.i${next[0]}.j${next[1]}`).addClass(`player-${currState.player}-prev`);
    if (curr[0] === -1) {
        curr = next;
    }
    if (curr[0] === 9) {
        currState.unusedPawns[currState.player]--;
        currState.board[next[0]][next[1]] = currState.player;
        $(elem).addClass(`player-${currState.player}`);
    } else {
        currState.board[curr[0]][curr[1]] = 9;
        if (next[0] < 0 || next[0] >= currState.board.length || next[1] < 0 || next[1] >= currState.board[next[0]].length) {
            currState.unusedPawns[currState.player]++;
            $(`.player-${currState.player}-start:not(.player-${currState.player}):last`).addClass(`player-${currState.player}`);
        } else {
            currState.board[next[0]][next[1]] = currState.player;
            $(elem).addClass(`player-${currState.player}`);
        }
        if (Math.abs(curr[0]-next[0]) > 1 || Math.abs(curr[1]-next[1]) > 1) {
            const temp = [curr[0] + getDelta(curr[0], next[0]), curr[1] + getDelta(curr[1], next[1])];
            if (currState.board[temp[0]][temp[1]] !== currState.player) {
                currState.unusedPawns[currState.board[temp[0]][temp[1]]]++;
                const prev = $(`.i${temp[0]}.j${temp[1]}`);
                prev.addClass(`player-${currState.board[temp[0]][temp[1]]}-prev`);
                $(`.player-${currState.board[temp[0]][temp[1]]}-start:not(.player-${currState.board[temp[0]][temp[1]]}):last`).addClass(`player-${currState.board[temp[0]][temp[1]]}`).addClass(`player-${currState.board[temp[0]][temp[1]]}-prev`);
                currState.board[temp[0]][temp[1]] = 9;
                prev.removeClass('player-0').removeClass('player-1').removeClass('player-2');
            }
        }
    }
    active.removeClass(`player-${currState.player}`);
    active.addClass(`player-${currState.player}-prev`);
    console.log("This could have been it:");
    console.log((curr[0] === 9 || Math.abs(curr[0]-next[0]) === 1 || Math.abs(curr[1]-next[1]) === 1) ? [9, 9] : next);
    refreshValidMoves([9, 9]);
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
                console.log(`${currState.player} - stop early`);
                refreshValidMoves([9, 9]);
            } else {
                if (path[0][0] === 9) {
                    $(`.player-${currState.player}-start.player-${currState.player}:last`).addClass("active");
                } else {
                    $(`.i${path[0][0]}.j${path[0][1]}`).addClass("active");
                }
                clickMove($(`.i${path[1][0]}.j${path[1][1]}`));
            }
            console.log(`${currState.player} - turn taken`);
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
                    $(`.i${path[1][0]}.j${path[1][1]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currState.player}-target`);
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
                    $(`.i${path[1][0]}.j${path[1][1]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currState.player}-target`);
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