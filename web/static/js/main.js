var currStatus = {
    player: 0,
    board: [
        [9, 9, 9, 9, 9, 9],
        [9, 9, 9, 9, 9],
        [9, 9, 9, 9],
        [9, 9, 9],
        [9, 9],
        [9]
    ],
    unused: [3, 3, 3],
    scores: [0, 0, 0],
    maxScore: 60
};

var gameOver = false;
var botPlayers = [1, 2];

var validMoves = [
    [9, 9, 0, 0],
    [9, 9, 0, 1],
    [9, 9, 0, 2],
    [9, 9, 0, 3],
    [9, 9, 0, 4],
    [9, 9, 0, 5]
]

function movePoints(board, player) {
    for (var i = board.length-1; i >= 0; i--) {
        for (var j = 0; j < board.length; j++) {
            if (board[i][j] == player) {
                return (i + 1)
            }
        }
    }
    return 0
}

function getCoords(elem) {
    var classes = $(elem).attr("class").split(/\s+/);
    var currI = -1;
    var currJ = -1;
    $.each(classes, function(_, v) {
        if (v.charAt(0) == 'i') {
            currI = parseInt(v.substr(1));
        }
        if (v.charAt(0) == 'j') {
            currJ = parseInt(v.substr(1));
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
    for (var i = 0; i < 3; i++) {
        var points = movePoints(currStatus.board, i);
        $(`.player-${i}-score .player-${i}-start .cell-inner`)[0].innerHTML = points;
        $(`.player-${i}-score span`)[0].innerHTML = currStatus.scores[i];
        $(`.player-${i}-score .player-${i}-start`).removeClass("scorable");
        $(`.player-${i}-target`).removeClass(`player-${i}-target`);
        if (points > 0 && points+currStatus.scores[i] <= currStatus.maxScore) {
            $(`.player-${i}-score .player-${i}-start`).addClass("scorable");
        }
    }
    $(".active").removeClass("active");
}

function refreshValidMoves() {
    $.ajax({
        url: "/availableMoves",
        type: "POST",
        data: JSON.stringify(currStatus),
        contentType: "application/json; charset=utf-8",
        success: function(data) {
            validMoves = data;
        }
    });
}

function clickScore() {
    clearPrevs();
    currStatus.scores[currStatus.player] += movePoints(currStatus.board, currStatus.player);
    if (currStatus.scores[currStatus.player] == currStatus.maxScore) {
        alert(`yay ${currStatus.player} wins!`);
        gameOver = true;
    }
    updatePointsScorable();
    $(".curr-turn").addClass(`player-${currStatus.player}-prev`);
    currStatus.player = (currStatus.player + 1) % 3;
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("curr-turn");
    refreshValidMoves();
}

function clickMove(elem) {
    var next = getCoords($(elem));
    var curr = getCoords($(".active")[0]);

    $(`.i${next[0]}.j${next[1]}`).addClass(`player-${currStatus.player}-prev`);
    if (curr[0] == -1) {
        curr = next;
    }
    if (curr[0] == 9) {
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
            var deltaI = getDelta(curr[0], next[0]);
            var deltaJ = getDelta(curr[1], next[1]);
            var temp = [curr[0] + deltaI, curr[1] + deltaJ];
            while (next[0] != temp[0] || next[1] != temp[1]) {
                if (currStatus.board[temp[0]][temp[1]] != 9) {
                    currStatus.unused[currStatus.board[temp[0]][temp[1]]]++;
                    $(`.i${temp[0]}.j${temp[1]}`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}-prev`);
                    $(`.player-${currStatus.board[temp[0]][temp[1]]}-start:not(.player-${currStatus.board[temp[0]][temp[1]]}):last`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}-prev`);
                }
                currStatus.board[temp[0]][temp[1]] = 9;
                $(`.i${temp[0]}.j${temp[1]}`).removeClass('player-0').removeClass('player-1').removeClass('player-2');
                temp[0] += deltaI;
                temp[1] += deltaJ;
            }
        }
    }
    $(".active").removeClass(`player-${currStatus.player}`);
    $(".active").addClass(`player-${currStatus.player}-prev`);
    currStatus.player = (currStatus.player + 1) % 3;
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("curr-turn");
    updatePointsScorable();
    refreshValidMoves();
}

function doBotMove() {
    $.ajax({
        url: "/botMove",
        type: "POST",
        data: JSON.stringify(currStatus),
        contentType: "application/json; charset=utf-8",
        success: function(v) {
            if (v[0] == 9 && v[2] == 9) {
                clickScore();
            } else {
                if (v[0] == 9) {
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

$('body').on('click', ".player-0, .player-1, .player-2", function () {
    if (!botPlayers.includes(currStatus.player)) {
        clearPrevs();
        var classes = $(this).attr("class").split(/\s+/);
        if (classes.includes(`player-${currStatus.player}`)) {
            $(".active").removeClass("active");
            $(".player-0-target").removeClass("player-0-target");
            $(".player-1-target").removeClass("player-1-target");
            $(".player-2-target").removeClass("player-2-target");
            $(this).addClass("active");
            if (classes.includes(`player-${currStatus.player}-start`)) {
                $.each(validMoves, function(_, v) {
                    if (v[0] == 9 && v[2] != 9) {
                        $(`.i${v[2]}.j${v[3]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currStatus.player}-target`);
                    }
                });
            } else {
                var curr = getCoords($(this));
                $.each(validMoves, function(_, v) {
                    if (v[0] == curr[0] && v[1] == curr[1]) {
                        $(`.i${v[2]}.j${v[3]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currStatus.player}-target`);
                    }
                });
            }
        }
    }
});

$('body').on('click', '.player-0-target, .player-1-target, .player-2-target', function () {
    if (!botPlayers.includes(currStatus.player)) {
        clickMove($(this));
    }
});

$('body').on('click', '.curr-turn.scorable', function () {
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