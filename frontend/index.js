var chapterOptions = []
var currentPages = []
var currentPage = 0;
var currentChapterName;
var currentChapterIndex;
var pageURL = window.location.href;

var drawerStatus = true;

var series = window.location.href.substring(("https://www.chemistry-tutor.com/reader/".length))
var seriesDisplay;

$(document).ready(function () {
    loadChapterOptions();
});
document.addEventListener('keydown', logKey);

function logKey(e) {
    if (e.code == 'ArrowLeft') {
        next();
    } else if (e.code == 'ArrowRight') {
        previous();
    }
}

$('#chapter-select').on('input', function () {
    var query = $("#chapter-select").val()
});

$('#chapter-select').keypress(function (event) {
    if (event.keyCode == 13) {
        selectChapter();
    }
});

async function loadChapterOptions() {
    await axios
        .get('https://www.chemistry-tutor.com/json/' + series)
        .then(response => {
            chapters = response.data.chapters;
            seriesDisplay = response.data.title;
            for (i = 0; i < chapters.length; i++) {
                chapterOptions.push(parseFloat(chapters[i].chapter.replace('-', '.')))
            }
            chapterOptions.sort(function (a, b) {
                return b - a;
            });

        })

    $("#chapter-select").attr('placeholder', '1...' + chapterOptions[0])
    console.log(chapterOptions)
}

async function loadChapter() {
    currentPages = []
    console.log("loading...")
    chapterTitle = currentChapterName.toString().replace(".", '-')
    axios
        .get('https://www.chemistry-tutor.com/getChapter/' + series + '/' + chapterTitle)
        .then(response => {
            pageLinks = response.data.links;

            axios
                .get('https://dwmc7ixdnoavh.cloudfront.net/Series/' + +series + '/' + chapterTitle + '/' + chapterTitle + '.json')
                .then(keyresponse => {
                    keys = keyresponse.data.keys
                    for (i = 0; i < pageLinks.length; i++) {
                        currentPages.push({
                            id: i,
                            url: pageLinks[i],
                            key: keys[i]
                        })

                        var img = $('<img />', {
                            id: i,
                            src: pageLinks[i],
                        });
                        $("#imgcontainer").append(img)

                        $('#imgcontainer').imagesLoaded(function () {
                            console.log("Images Loaded");
                            displayToCanvas();
                        });


                        if (currentChapterIndex + 1 < chapterOptions.length) {
                            sideLoadChapter(currentChapterIndex + 1);
                        }
                        if (currentChapterIndex - 1 >= 0) {
                            sideLoadChapter(currentChapterIndex - 1);
                        }
                    }
                })
        })
}





function displayToCanvas() {
    leftCtx = document.getElementById("left").getContext("2d");
    rightCtx = document.getElementById("right").getContext("2d");

    rightImage = document.getElementById("" + currentPage);
    leftImage = document.getElementById("" + (currentPage + 1));

    console.log(currentPage);
    drawPage(rightImage, rightCtx, currentPages[currentPage].key);
    drawPage(leftImage, leftCtx, currentPages[currentPage+1].key);
}


function drawPage(image, ctx, key) {
    _ = parseInt(image.naturalWidth - 90);
    v = parseInt(image.naturalHeight - 140);

    ctx.canvas.width = _;
    ctx.canvas.height = v;

    P = key.split(":");

    w = Math.floor(_ / 10);
    b = Math.floor(v / 15);

    ctx.clearRect(0, 0, _, v);

    for (ctx.drawImage(image, 0, 0, _, b, 0, 0, _, b),
        ctx.drawImage(image, 0, b + 10, w, v - 2 * b, 0, b, w, v - 2 * b),
        ctx.drawImage(image, 0, 14 * (b + 10), _, image.height - 14 * (b + 10), 0, 14 * b, _, image.height - 14 * (b + 10)),
        ctx.drawImage(image, 9 * (w + 10), b + 10, w + (_ - 10 * w), v - 2 * b, 9 * w, b, w + (_ - 10 * w), v - 2 * b),
        m = 0; m < P.length; m++)
        P[m] = parseInt(P[m], 16),
        ctx.drawImage(image, Math.floor((m % 8 + 1) * (w + 10)), Math.floor((Math.floor(m / 8) + 1) * (b + 10)), Math.floor(w), Math.floor(b), Math.floor((P[m] % 8 + 1) * w), Math.floor((Math.floor(P[m] / 8) + 1) * b), Math.floor(w), Math.floor(b));
}

function selectChapter() {
    $("#imgcontainer").empty();
    var selected = $("#chapter-select").val();
    console.log("Selected: ", selected)
    var found = false;
    for (i = 0; i < chapterOptions.length; i++) {
        if (selected == chapterOptions[i]) {
            currentChapterName = chapterOptions[i];
            currentChapterIndex = i;
            found = true;
            break;
        }
    }
    if (!found) {
        if (selected > chapterOptions[0]) {
            $("#chapter-select").val(chapterOptions[0]);
        } else if (selected <= 0) {
            $("#chapter-select").val(1);
        } else {
            alert("that chapter does not exist")
        }

    } else {
        console.log("Valid Selection!")

        $("#title").text(seriesDisplay)
        $("#chapter").text("Ch. " + $('#chapter-select').val())
        //$('#chapter-select').val('');
        $('#drawer').css({
            top: '97vh',
            'box-shadow': 'none'
        });

        drawerStatus = false;

        loadChapter();

        setTimeout(function () {
            $('#chapter-select').val('');
        }, 1000)
    }
}

function pullUp() {
    if (drawerStatus == false) {
        $('#drawer').css({
            top: '0vh',
            'box-shadow': 'rgba(50, 50, 93, 0.25) 0px 50px 150px -20px, rgba(0, 0, 0, 0.3) 0px 30px 120px -30px'
        });
        drawerStatus = true;
    } else {
        $('#drawer').css({
            top: '97vh',
            'box-shadow': 'none'
        });
        drawerStatus = false;
    }

}

function next() {
    leftCtx = document.getElementById("left").getContext("2d");
    rightCtx = document.getElementById("right").getContext("2d");
    if (currentPage + 2 < currentPages.length) {
        currentPage += 2;
        displayToCanvas();
    }

}

function previous() {
    leftCtx = document.getElementById("left").getContext("2d");
    rightCtx = document.getElementById("right").getContext("2d");
    if (currentPage > 1) {
        currentPage -= 2;
        displayToCanvas();
    }

}

function nextChapter() {
    if (currentChapterIndex - 1 >= 0) {
        $("#imgcontainer").empty();
        leftCtx = document.getElementById("left").getContext("2d");
        rightCtx = document.getElementById("right").getContext("2d");
        currentChapterIndex--;
        currentChapterName = chapterOptions[currentChapterIndex];
        currentPage = 0;
        $("#title").text(seriesDisplay)
        $("#chapter").text("Ch. " + currentChapterName)
        loadChapter();
    }
}

function previousChapter() {
    if (currentChapterIndex + 1 < chapterOptions.length) {
        $("#imgcontainer").empty();
        leftCtx = document.getElementById("left").getContext("2d");
        rightCtx = document.getElementById("right").getContext("2d");
        currentChapterIndex++;
        currentChapterName = chapterOptions[currentChapterIndex];
        currentPage = 0;
        $("#title").text(seriesDisplay)
        $("#chapter").text("Ch. " + currentChapterName)
        loadChapter();
    }
}

document.getElementById('canvases').onclick = function clickEvent(e) {
    var x = e.clientX;
    var screenW = $(document).width() / 2;

    var offset = $("#canvases").offset();
    var objWidth = $("#canvases").width() / 8;
    objWidth *= 3;

    relative = x - screenW;
    relativeAbs = Math.abs(relative)

    if (relative < 0 && relativeAbs > objWidth) {
        next();
    } else if (relativeAbs > objWidth) {
        previous();
    }
}

function sideLoadChapter(i) {
    chapter = chapterOptions[i]
    axios
        .get('https://www.chemistry-tutor.com/getChapter/' + series + '/' + chapter)
        .then(response => {
            console.log("Side Chapter Loaded: " + chapter)
        })
}  
//comment