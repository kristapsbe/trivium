const currStatus = {
    player: 0,
    board: [
              [9],
             [9, 9],
           [9, 9, 9],
          [9, 9, 9, 9],
         [9, 9, 9, 9, 9],
        [9, 9, 9, 9, 9, 9]
    ].reverse(),
    unused: [3, 3, 3],
    scores: [0, 0, 0],
    maxScore: 60,
    forceMove: [9, 9]
};

let gameOver = false;
let botPlayers = [];

let validMoves = [
    [9, 9, 0, 0],
    [9, 9, 0, 1],
    [9, 9, 0, 2],
    [9, 9, 0, 3],
    [9, 9, 0, 4],
    [9, 9, 0, 5]
];

$('#bot1, #bot2, #bot3').change(function() {
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
    let currI = 9;
    let currJ = 9;
    $.each(classes, function(_, v) {
        if (v.charAt(0) === 'i') {
            currI = parseInt(v.substring(1));
        }
        if (v.charAt(0) === 'j') {
            currJ = parseInt(v.substring(1));
        }
    });
    return [currI, currJ];
}

function getDelta(currX, nextX) {
    if (currX < nextX) {
        return 1;
    } else if (currX > nextX) {
        return -1;
    }
    return 0;
}

function updatePointsScorable() {
    for (let i = 0; i < 3; i++) {
        const points = availableProgressPoints(currStatus.board, i);
        $(`.player-${i}-score .player-${i}-start .cell-inner`)[0].innerHTML = points;
        $(`.player-${i}-score span`)[0].innerHTML = currStatus.scores[i];
        $(`.player-${i}-score .player-${i}-start`).removeClass("scorable");
        $(`.player-${i}-target`).removeClass(`player-${i}-target`);
        if ((points > 0) && (points+currStatus.scores[i] <= currStatus.maxScore) && (currStatus.forceMove[0] === 9)) {
            $(`.player-${i}-score .player-${i}-start`).addClass("scorable");
        }
    }
    $(".active").removeClass("active");
}

function refreshValidMoves(next) {
    if (next[0] === 9) {
        currStatus.player = ++currStatus.player % 3;
    }
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("curr-turn");
    currStatus.forceMove = next;
    updatePointsScorable();
    $.ajax({
        url: "/availableMoves",
        type: "POST",
        data: JSON.stringify(currStatus),
        contentType: "application/json; charset=utf-8",
        success: function(data) {
            if (data.length === 1 && data[0][0] !== 9 && (data[0][0] === data[0][2] && data[0][1] === data[0][3])) {
                refreshValidMoves([9, 9]);
            } else {
                validMoves = data;
                if (next[0] !== 9) {
                    setActive($(`.i${currStatus.forceMove[0]}.j${currStatus.forceMove[1]}`)[0]);
                }
            }
        }
    });
}

const playerColors = {0: "Red", 1: "Green", 2: "Blue"};
function clickScore() {
    clearPrevs();
    currStatus.scores[currStatus.player] += availableProgressPoints(currStatus.board, currStatus.player);
    if (currStatus.scores[currStatus.player] === currStatus.maxScore) {
        alert(`Yay, ${playerColors[currStatus.player]} wins!`);
        gameOver = true;
    }
    $(".curr-turn").addClass(`player-${currStatus.player}-prev`);
    refreshValidMoves([9, 9]);
}

function clickMove(elem) {
    const next = getCoords($(elem));
    const active = $(".active");
    let curr = getCoords(active[0]);

    $(`.i${next[0]}.j${next[1]}`).addClass(`player-${currStatus.player}-prev`);
    if (curr[0] === -1) {
        curr = next;
    }
    if (curr[0] === 9) {
        currStatus.unused[currStatus.player]--;
        currStatus.board[next[0]][next[1]] = currStatus.player;
        $(elem).addClass(`player-${currStatus.player}`);
    } else {
        currStatus.board[curr[0]][curr[1]] = 9;
        if (next[0] < 0 || next[0] >= currStatus.board.length || next[1] < 0 || next[1] >= currStatus.board[next[0]].length) {
            currStatus.unused[currStatus.player]++;
            $(`.player-${currStatus.player}-start:not(.player-${currStatus.player}):last`).addClass(`player-${currStatus.player}`);
        } else {
            currStatus.board[next[0]][next[1]] = currStatus.player;
            $(elem).addClass(`player-${currStatus.player}`);
        }
        if (Math.abs(curr[0]-next[0]) > 1 || Math.abs(curr[1]-next[1]) > 1) {
            const temp = [curr[0] + getDelta(curr[0], next[0]), curr[1] + getDelta(curr[1], next[1])];
            if (currStatus.board[temp[0]][temp[1]] !== currStatus.player) {
                currStatus.unused[currStatus.board[temp[0]][temp[1]]]++;
                $(`.i${temp[0]}.j${temp[1]}`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}-prev`);
                $(`.player-${currStatus.board[temp[0]][temp[1]]}-start:not(.player-${currStatus.board[temp[0]][temp[1]]}):last`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}-prev`);
                currStatus.board[temp[0]][temp[1]] = 9;
                $(`.i${temp[0]}.j${temp[1]}`).removeClass('player-0').removeClass('player-1').removeClass('player-2');
            }
        }
    }
    active.removeClass(`player-${currStatus.player}`);
    active.addClass(`player-${currStatus.player}-prev`);
    refreshValidMoves((curr[0] === 9 || Math.abs(curr[0]-next[0]) === 1 || Math.abs(curr[1]-next[1]) === 1) ? [9, 9] : next);
}

function doBotMove() {
    $.ajax({
        url: "/botMove",
        type: "POST",
        data: JSON.stringify(currStatus),
        contentType: "application/json; charset=utf-8",
        success: function(v) {
            console.log("Incoming: " + v)
            if (v[0] === 9 && v[2] === 9) {
                clickScore();
            } else if (v[0] === v[2] && v[1] === v[3]) {
                // a bot is allowed to stop multi-moves early
                console.log(`${currStatus.player} - stop early`);
                refreshValidMoves([9, 9]);
            } else {
                if (v[0] === 9) {
                    $(`.player-${currStatus.player}-start.player-${currStatus.player}:last`).addClass("active");
                } else {
                    $(`.i${v[0]}.j${v[1]}`).addClass("active");
                }
                clickMove($(`.i${v[2]}.j${v[3]}`));
            }
            console.log(`${currStatus.player} - turn taken`);
        }
    });
}

function clearPrevs() {
    $(".cell").removeClass("player-0-prev").removeClass("player-1-prev").removeClass("player-2-prev");
}

function setActive(elem) {
    if (currStatus.forceMove[0] === 9) {
        clearPrevs();
    }
    const classes = $(elem).attr("class").split(/\s+/);
    if (classes.includes(`player-${currStatus.player}`)) {
        $(".active").removeClass("active");
        $(".player-0-target").removeClass("player-0-target");
        $(".player-1-target").removeClass("player-1-target");
        $(".player-2-target").removeClass("player-2-target");
        $(elem).addClass("active");
        if (classes.includes(`player-${currStatus.player}-start`)) {
            $.each(validMoves, function(_, v) {
                if (v[0] === 9 && v[2] !== 9) {
                    $(`.i${v[2]}.j${v[3]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currStatus.player}-target`);
                }
            });
        } else {
            const curr = getCoords($(elem));
            $.each(validMoves, function(_, v) {
                if (v[0] === curr[0] && v[1] === curr[1]) {
                    $(`.i${v[2]}.j${v[3]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currStatus.player}-target`);
                }
            });
        }
    }
}

const body = $('body');
body.on('click', ".player-0, .player-1, .player-2", function () {
    if (!botPlayers.includes(currStatus.player)) {
        if (currStatus.forceMove[0] === 9) {
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
    if (!botPlayers.includes(currStatus.player)) {
        clickMove($(this));
    }
});

body.on('click', '.curr-turn.scorable', function () {
    if (!botPlayers.includes(currStatus.player)) {
        clickScore();
    }
});

function botLoop() {
    setTimeout(function() {
        if (!botPlayers.includes(currStatus.player)) {
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