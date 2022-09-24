currStatus = {
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

validMoves = [
    [9, 9, 0, 0],
    [9, 9, 0, 1],
    [9, 9, 0, 2],
    [9, 9, 0, 3],
    [9, 9, 0, 4],
    [9, 9, 0, 5]
]

function movePoints(board, player) {
    for (i = board.length-1; i >= 0; i--) {
        console.log(i)
        for (j = 0; j < board.length; j++) {
            if (board[i][j] == player) {
                return (i + 1)
            }
        }
    }
    return 0
}

function getCoords(elem) {
    classes = $(elem).attr("class").split(/\s+/);
    currI = -1;
    currJ = -1;
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

$('body').on('click', ".player-0, .player-1, .player-2", function () {
    classes = $(this).attr("class").split(/\s+/);
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
            curr = getCoords($(this));
            $.each(validMoves, function(_, v) {
                if (v[0] == curr[0] && v[1] == curr[1]) {
                    $(`.i${v[2]}.j${v[3]}:not(.player-0):not(.player-1):not(.player-2)`).addClass(`player-${currStatus.player}-target`);
                }
            });
        }
    }
});

$('body').on('click', '.player-0-target, .player-1-target, .player-2-target', function () {
    next = getCoords($(this));
    curr = getCoords($(".active")[0]);
    if (curr[0] == -1) {
        curr = next;
    }
    if (curr[0] == 9) {
        currStatus.unused[currStatus.player]--;
        currStatus.board[next[0]][next[1]] = currStatus.player;
        $(this).addClass(`player-${currStatus.player}`);
    } else {
        currStatus.board[curr[0]][curr[1]] = 9;
        if (next[0] < 0 || next[0] >= currStatus.board.len || next[1] < 0 || next[1] >= currStatus.board[next[0]].len) {
            currStatus.unused[currStatus.player]++;
            $(`.player-${currStatus.player}-start:not(.player-${currStatus.player}):first`).addClass(`player-${currStatus.player}`);
        } else {
            currStatus.board[next[0]][next[1]] = currStatus.player;
            $(this).addClass(`player-${currStatus.player}`);
        }
        if (Math.abs(curr[0]-next[0]) > 1 || Math.abs(curr[1]-next[1]) > 1) {
            deltaI = getDelta(curr[0], next[0]);
            deltaJ = getDelta(curr[1], next[1]);
            temp = [curr[0] + deltaI, curr[1] + deltaJ];
            while (next[0] != temp[0] || next[1] != temp[1]) {
                if (currStatus.board[temp[0]][temp[1]] != 9) {
                    currStatus.unused[currStatus.board[temp[0]][temp[1]]]++;
                    $(`.player-${currStatus.board[temp[0]][temp[1]]}-start:not(.player-${currStatus.board[temp[0]][temp[1]]}):first`).addClass(`player-${currStatus.board[temp[0]][temp[1]]}`);
                }
                currStatus.board[temp[0]][temp[1]] = 9;
                $(`.i${temp[0]}.j${temp[1]}`).removeClass('player-0').removeClass('player-1').removeClass('player-2');
                temp[0] += deltaI;
                temp[1] += deltaJ;
            }
        }
    }
    points = movePoints(currStatus.board, currStatus.player);
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start .cell-inner`)[0].innerHTML = points;
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).removeClass("scorable");
    if (points+currStatus.scores[currStatus.player] <= currStatus.maxScore) {
        $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("scorable");
    }
    currStatus.player = (currStatus.player + 1) % 3;
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("curr-turn");
    $(".active").removeClass("player-0").removeClass("player-1").removeClass("player-2").removeClass("active");
    $(".player-0-target").removeClass("player-0-target");
    $(".player-1-target").removeClass("player-1-target");
    $(".player-2-target").removeClass("player-2-target");
    $.ajax({
        url: "/availableMoves",
        type: "POST",
        data: JSON.stringify(currStatus),
        contentType: "application/json; charset=utf-8",
        success: function(data) {
            validMoves = data;
        }
    });
});


$('body').on('click', '.curr-turn.scorable', function () {
    points = movePoints(currStatus.board, currStatus.player);
    currStatus.scores[currStatus.player] += points;
    $(`.player-${currStatus.player}-score span`)[0].innerHTML = currStatus.scores[currStatus.player];
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).removeClass("scorable");
    if (points+currStatus.scores[currStatus.player] <= currStatus.maxScore) {
        $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("scorable");
    }
    currStatus.player = (currStatus.player + 1) % 3;
    $(".curr-turn").removeClass("curr-turn");
    $(`.player-${currStatus.player}-score .player-${currStatus.player}-start`).addClass("curr-turn");
});